package cmd

import (
	ssmbackend "confd-template/backend/ssm"
	"confd-template/engine/yaml"
	"confd-template/template"
	"confd-template/validation"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	backendtype string
	delimiter   string
	filter      string
	format      string
	loglevel    string
	optional    bool
	outfile     string
	prefix      string
)

// define cli entrypoint
var rootCmd = &cobra.Command{
	Use:   "confd-template",
	Short: "a cli for generating confd templates using a populated KV backend",
	Run: func(cmd *cobra.Command, args []string) {
		configureLogging()

		t := loadTemplate()
		b := loadBackend()
		e := loadEngine(t)
		r, err := template.NewRenderer(&template.Config{
			Backend:  b,
			Engine:   e,
			Logger:   logrus.StandardLogger(),
			Template: t,
		})
		if err != nil {
			logrus.WithError(err).Fatalln("error creating renderer...")
		}

		// render template
		err = r.Render()
		if err != nil {
			logrus.WithError(err).Fatalln("error rendering template")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&backendtype, "backend", "ssm", "configuration backend")
	rootCmd.PersistentFlags().StringVar(&delimiter, "delimiter", "/", "key delimiter")
	rootCmd.PersistentFlags().StringVar(&filter, "filter", "", "optional regex key filter")
	rootCmd.PersistentFlags().StringVar(&format, "format", "yaml", "template format")
	rootCmd.PersistentFlags().StringVar(&loglevel, "level", "info", "log level")
	rootCmd.PersistentFlags().BoolVar(&optional, "optional", false, "wrap template interpolations with conditional wrapper`")
	rootCmd.PersistentFlags().StringVar(&outfile, "out", "", "output template path")
	rootCmd.PersistentFlags().StringVar(&prefix, "prefix", "/", "key prefix to scan")
}

// Execute the cli
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// configureLogging configures logging format and verbosity
func configureLogging() {
	// set log verbosity
	switch strings.ToLower(strings.TrimSpace(loglevel)) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

// loadBackend creates and validates a new backend value
func loadBackend() template.Backend {
	var b template.Backend
	var err error
	switch backendtype {
	case "ssm":
		sess := session.Must(session.NewSession())
		svc := ssm.New(sess)
		b, err = ssmbackend.New(&ssmbackend.Config{
			Logger: logrus.StandardLogger(),
			SSM:    svc,
		})
		if err != nil {
			logrus.WithError(err).Fatalln("error creating ssm backend")
		}
	default:
		logrus.Fatalf("unsupported backend type: %s", backendtype)
	}
	return b
}

// loadEngine creates and loads formatting engine
func loadEngine(t *template.Template) template.Engine {
	var e template.Engine
	var err error
	switch format {
	case "yaml":
		e, err = yaml.New(&yaml.Config{
			Logger:   logrus.StandardLogger(),
			Optional: optional,
			Template: t,
		})
		if err != nil {
			logrus.WithError(err).Fatalln("error creating renderer")
		}
	default:
		logrus.Fatalf("unsupported renderer format: %s", format)
	}
	return e
}

// loadTemplate creates and validates a new template value
func loadTemplate() *template.Template {
	// define template
	t := &template.Template{
		Delimiter: delimiter,
		Filter:    filter,
		Format:    format,
		Outfile:   outfile,
		Prefix:    prefix,
	}
	if err := validation.Validate.Struct(t); err != nil {
		logrus.WithError(err).Fatalln("error loading template...")
	}
	return t
}
