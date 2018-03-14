// Package template defines application ontology including domain
// types, interfaces, and core methods
package template

import (
	"confd-template/validation"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// Backend provides access to a KV configuration store
type Backend interface {
	Keys(*Template) chan *Key
}

// Config defines a Renderer configuration
type Config struct {
	Backend  Backend            `validate:"required"`
	Engine   Engine             `validate:"required"`
	Logger   logrus.FieldLogger `validate:"required"`
	Template *Template          `validate:"required"`
}

// Engine represents a template formatting engine
type Engine interface {
	Render(chan *Key, io.Writer) error
}

// Key represents an invidvidual key value pair
type Key struct {
	Error error
	Name  string
	Path  []string
	Value string
}

// Renderer coordinates template rendering using the specified
// backend and formatting engine
type Renderer struct {
	backend  Backend
	engine   Engine
	filter   *regexp.Regexp
	log      logrus.FieldLogger
	template *Template
}

// NewRenderer returns a new Renderer value
func NewRenderer(c *Config) (*Renderer, error) {
	// validate config
	if err := validation.Validate.Struct(c); err != nil {
		return nil, err
	}

	// create new renderer value
	r := &Renderer{
		backend:  c.Backend,
		engine:   c.Engine,
		log:      c.Logger,
		template: c.Template,
	}

	// add filter
	if c.Template.Filter != "" {
		filter, err := regexp.Compile(c.Template.Filter)
		if err != nil {
			r.log.WithError(err).Errorln("unable to compile filter expression")
			return r, err
		}
		r.filter = filter
	}

	return r, nil
}

// Render initiates the rendering process by piping keys from the backend
// to the appropriate formatting engine
func (r *Renderer) Render() error {
	// define input and output channels
	out := make(chan *Key, 500)
	in := r.backend.Keys(r.template)

	// define render target
	var target io.WriteCloser
	var err error
	if r.template.Outfile != "" {
		target, err = os.Create(r.template.Outfile)
		if err != nil {
			r.log.WithError(err).Errorln("error opening outfilefile")
			return err
		}
		defer target.Close()
	} else {
		target = os.Stdout
	}

	// iterate through `in` keys, handle errors and apply filters
	// before emitting to engine
	go r.pipe(in, out)

	return r.engine.Render(out, target)
}

// pipe applies filters and transformations to backend keys before piping
// to the downstream formatting engine
func (r *Renderer) pipe(in, out chan *Key) {
	for k := range in {
		// handle error
		if k.Error != nil {
			r.log.WithError(k.Error).Errorln("key error detected")
			break
		}

		// apply filters
		k.Name = strings.Replace(k.Name, r.template.Prefix, "", 1)
		if r.filter != nil && !r.filter.MatchString(k.Name) {
			r.log.WithField("key", k.Name).Debugln("filtering key")
			continue
		}

		// append path to key
		k.Path = strings.Split(strings.TrimPrefix(k.Name, r.template.Delimiter), r.template.Delimiter)
		out <- k
	}
	close(out)
}

// Template defines a template to render
type Template struct {
	Delimiter string `validate:"required"`
	Filter    string
	Format    string `validate:"required,oneof=yaml"`
	Outfile   string
	Prefix    string `validate:"required"`
}
