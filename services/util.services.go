package services

import (
	"bytes"
	"context"

	"github.com/a-h/templ"
)

func RenderComponent(component templ.Component) string {
	buf := new(bytes.Buffer)
	component.Render(context.Background(), buf)

	return buf.String()
}
