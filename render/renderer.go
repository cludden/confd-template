package render

// Renderer describes a generic template renderer
type Renderer interface {
	Render() error
}
