package yaml

import (
	"bytes"
	"confd-template/template"
	"confd-template/validation"
	"fmt"
	"io"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/sirupsen/logrus"
)

// Engine manages rendering a single confd yaml template
type Engine struct {
	log      logrus.FieldLogger
	optional bool
	template *template.Template
}

// New returns a new Engine value
func New(c *Config) (*Engine, error) {
	// validate config
	if err := validation.Validate.Struct(c); err != nil {
		return nil, err
	}

	// create new renderer value
	e := &Engine{
		log:      c.Logger,
		optional: c.Optional,
		template: c.Template,
	}

	return e, nil
}

// Render processes the template using the specified backend
func (e *Engine) Render(keys chan *template.Key, w io.Writer) error {
	// create generic key value hierarchy
	data := make(map[string]interface{})

	// convert keys into template data
	for k := range keys {
		if err := e.appendKey(data, k); err != nil {
			e.log.WithField("key", k.Name).Errorln("error appending key")
			return err
		}
	}

	// render template data as yaml
	tmpl, err := e.render(data, 0)
	if err != nil {
		return err
	}

	// write rendered template to writer
	if _, err = w.Write(tmpl); err != nil {
		e.log.WithError(err).Errorln("error writing rendered template")
		return err
	}

	// render
	return nil
}

// appendKey assigns a confd key value pair to the yaml data map
func (e *Engine) appendKey(data map[string]interface{}, key *template.Key) error {
	container := data
	last := key.Path[len(key.Path)-1]
	// build nested container
	var c map[string]interface{}
	for i := 0; i < len(key.Path)-1; i++ {
		k := key.Path[i]
		tmp, ok := container[k]
		if !ok {
			c = make(map[string]interface{})
			container[k] = c
		} else {
			c, ok = tmp.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unable to coerce nested container to map[string]interface{} for key %s", k)
			}
		}
		container = c
	}

	var prefix string
	var suffix string
	var format string

	// if value is not a JSON scalar or stringified data structure, wrap in quotes
	if !isJSON(key.Value) {
		prefix = "\""
		suffix = "\""
	}

	if e.optional {
		format = fmt.Sprintf("{{ getv \"%s\" \"null\" }}", key.Name)
	} else {
		format = fmt.Sprintf("{{ getv \"%s\" }}", key.Name)
	}

	// add key interpolation to container
	container[last] = fmt.Sprintf("%s%s%s", prefix, format, suffix)

	return nil
}

// render the provided data as yaml with the specified indent level
func (e *Engine) render(data map[string]interface{}, i int) ([]byte, error) {
	whitespace := strings.Repeat("  ", i)
	tmpl := bytes.Buffer{}
	// iterate through top level keys
	for k, v := range data {
		switch x := v.(type) {
		case map[string]interface{}:
			p, err := e.render(x, i+1)
			if err != nil {
				e.log.WithError(err).WithField("key", k).Errorln("render error detected")
				break
			}
			tmpl.Write([]byte(fmt.Sprintf("%s%s:\n%s", whitespace, k, p)))
		default:
			tmpl.Write([]byte(fmt.Sprintf("%s%s: %v\n", whitespace, k, v)))
		}
	}
	return tmpl.Bytes(), nil
}

// Config defines the input to a NewRenderer operation
type Config struct {
	Logger   logrus.FieldLogger `validate:"required"`
	Optional bool               `validate:"required"`
	Template *template.Template `validate:"required"`
}

// isJSON determines whether or not a given string represents a valid JSON
// scalar value
func isJSON(s string) bool {
	_, err := gabs.ParseJSON([]byte(s))
	return err == nil
}
