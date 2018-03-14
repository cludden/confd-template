package yaml

import (
	"bytes"
	"confd-template/backend"
	"confd-template/template"
	"confd-template/validation"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/sirupsen/logrus"
)

// Renderer manages rendering a single confd template
type Renderer struct {
	backend   backend.Backend
	data      map[string]interface{}
	delimiter string
	filter    *regexp.Regexp
	log       logrus.FieldLogger
	template  *template.Template
}

// NewRenderer returns a new Renderer value
func NewRenderer(c *Config) (*Renderer, error) {
	// validate config
	if err := validation.Validate.Struct(c); err != nil {
		return nil, err
	}

	// create new renderer value
	r := &Renderer{
		backend:   c.Backend,
		data:      make(map[string]interface{}),
		delimiter: "/",
		log:       c.Logger,
		template:  c.Template,
	}

	// assign delimiter
	if c.Template.Delimiter != "" {
		r.delimiter = c.Template.Delimiter
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

// Render processes the template using the specified backend
func (r *Renderer) Render() error {
	// convert keys into template data
	if err := r.buildTemplateData(); err != nil {
		return err
	}

	// render template data as yaml
	tmpl, err := r.render(r.data, 0)
	if err != nil {
		return err
	}

	// write rendered template to outfile
	if r.template.Outfile != "" {
		file, err := os.Create(r.template.Outfile)
		if err != nil {
			r.log.WithError(err).Errorln("error opening outfilefile")
			return err
		}
		defer file.Close()

		if _, err = file.Write(tmpl); err != nil {
			r.log.WithError(err).Errorln("error writing to outfile")
			return err
		}
	}

	// render
	return nil
}

// appendKey assigns a confd key value pair to the renderer data map
func (r *Renderer) appendKey(path []string, key *backend.Key) error {
	container := r.data
	last := path[len(path)-1]
	// build nested container
	var c map[string]interface{}
	for i := 0; i < len(path)-1; i++ {
		k := path[i]
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

	if isJSON(key.Value) {
		container[last] = fmt.Sprintf("{{ getv \"%s\" }}", key.Name)
	} else {
		container[last] = fmt.Sprintf("\"{{ getv \"%s\" }}\"", key.Name)
	}

	return nil
}

// iterates through backend keys and builds template data
func (r *Renderer) buildTemplateData() error {
	// build renderer data
	keys := r.backend.Keys(r.template)
	for k := range keys {
		// handle error
		if k.Error != nil {
			r.log.WithError(k.Error).Errorln("key error detected")
			return k.Error
		}

		// apply filters
		k.Name = strings.Replace(k.Name, r.template.Prefix, "", 1)
		if r.filter != nil && !r.filter.MatchString(k.Name) {
			r.log.WithField("key", k.Name).Debugln("filtering key")
			continue
		}

		// append key to renderer data
		path := strings.Split(k.Name, r.delimiter)
		if path[0] == "" {
			path = path[1:]
		}
		if err := r.appendKey(path, k); err != nil {
			r.log.WithField("key", k.Name).Errorln("error appending key")
			return err
		}
	}
	return nil
}

// render the provided data as yaml with the specified indent level
func (r *Renderer) render(data map[string]interface{}, i int) ([]byte, error) {
	whitespace := strings.Repeat("  ", i)
	tmpl := bytes.Buffer{}
	// iterate through top level keys
	for k, v := range data {
		switch x := v.(type) {
		case map[string]interface{}:
			p, err := r.render(x, i+1)
			if err != nil {
				r.log.WithError(err).WithField("key", k).Errorln("render error detected")
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
	Backend  backend.Backend    `validate:"required"`
	Logger   logrus.FieldLogger `validate:"required"`
	Template *template.Template `validate:"required"`
}

func isJSON(s string) bool {
	_, err := gabs.ParseJSON([]byte(s))
	return err == nil
}
