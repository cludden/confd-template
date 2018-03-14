package ssm

import (
	"confd-template/template"
	"confd-template/validation"
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/sirupsen/logrus"
)

// Backend provides access to a KV store backed by SSM
type Backend struct {
	keys chan *template.Key
	log  logrus.FieldLogger
	svc  ssmiface.SSMAPI
}

// New returns a new Backend value
func New(c *Config) (*Backend, error) {
	if err := validation.Validate.Struct(c); err != nil {
		return nil, err
	}
	b := &Backend{
		keys: make(chan *template.Key, 500),
		log:  c.Logger,
		svc:  c.SSM,
	}
	return b, nil
}

// Keys returns a buffered channel of keys
func (b *Backend) Keys(t *template.Template) chan *template.Key {
	go b.streamKeys(t)
	return b.keys
}

// streamKeys asynchronously retrieves all keys that match the template configuraton and
// emits them on the given channel
func (b *Backend) streamKeys(t *template.Template) {
	// define ssm request parameters
	params := &ssm.GetParametersByPathInput{
		MaxResults:     aws.Int64(10),
		Path:           aws.String(t.Prefix),
		Recursive:      aws.Bool(true),
		WithDecryption: aws.Bool(true),
	}

	// define parameter page handler
	handler := func(output *ssm.GetParametersByPathOutput, last bool) bool {
		for _, p := range output.Parameters {
			b.keys <- &template.Key{
				Name:  *p.Name,
				Value: *p.Value,
			}
		}
		return !last
	}

	// process all parameters until there are no parameters left to process
	if err := b.svc.GetParametersByPathPagesWithContext(context.Background(), params, handler); err != nil {
		b.log.WithError(err).Errorln("ssm error detected")
		b.keys <- &template.Key{
			Error: err,
		}
	}
	close(b.keys)
}

// Config defines the input to a New backend operation
type Config struct {
	Logger logrus.FieldLogger `validate:"required"`
	SSM    ssmiface.SSMAPI    `validate:"required"`
}
