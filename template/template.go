package template

// Template defines a template to render
type Template struct {
	Delimiter string `validate:"required"`
	Filter    string
	Format    string `validate:"required,oneof=yaml"`
	Outfile   string
	Prefix    string `validate:"required"`
}
