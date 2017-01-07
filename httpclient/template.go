package httpclient

import (
	"bytes"
	"text/template"
)

type TemplateWrapper struct {
	template *template.Template
}

// implement encoding/TextUnmarshaler
func (t *TemplateWrapper) UnmarshalText(text []byte) error {
	str := string(text)
	tmpl := template.New(str)
	tmpl, err := tmpl.Parse(str)
	t.template = tmpl
	return err
}

// Apply template to the passed object
func (w *TemplateWrapper) Apply(o interface{}) (string, error) {
	buffer := new(bytes.Buffer)
	err := w.template.Execute(buffer, o)
	return buffer.String(), err
}
