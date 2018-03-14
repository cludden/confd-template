package backend

import "confd-template/template"

// Backend provides access to a KV configuration store
type Backend interface {
	Keys(*template.Template) chan *Key
}

// Key represents an invidvidual key value pair
type Key struct {
	Error error
	Name  string
	Value string
}
