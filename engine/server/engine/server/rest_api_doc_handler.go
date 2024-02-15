package server

// Based on https://github.com/go-goyave/openapi3

import (
	"io"
	"net/http"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

type UIConfig struct {
	Title        string
	Favicon32x32 string
	Favicon16x16 string
	BundleURL    string
	PresetURL    string
	StylesURL    string
	Spec         string
}

const uiTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
	<title>{{ .Title }}</title>
    <link rel="stylesheet" type="text/css" href="{{ .StylesURL }}" >
    <link rel="icon" type="image/png" href="{{ .Favicon32x32 }}" sizes="32x32" />
    <link rel="icon" type="image/png" href="{{ .Favicon16x16 }}" sizes="16x16" />
    <style>
      html
      {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }
      *,
      *:before,
      *:after
      {
        box-sizing: inherit;
      }
      body
      {
        margin:0;
        background: #fafafa;
      }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="{{ .BundleURL }}"> </script>
    <script src="{{ .PresetURL }}"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        spec: {{ .Spec }},
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      })
      // End Swagger UI call region
      window.ui = ui
    }
  </script>
  </body>
</html>
`

func NewSwaggerUIConfig(spec *openapi3.T) UIConfig {
	var json []byte
	if spec == nil {
		json = []byte{}
	} else {
		json, _ = spec.MarshalJSON()
	}
	return UIConfig{
		Title:        spec.Info.Title + " API Documentation",
		Favicon16x16: "https://raw.githubusercontent.com/kurtosis-tech/kurtosis/70f492c6e3bfa46caa78e93b90b70f4f005ed2fc/docs/static/img/favicon.ico",
		Favicon32x32: "https://raw.githubusercontent.com/kurtosis-tech/kurtosis/70f492c6e3bfa46caa78e93b90b70f4f005ed2fc/docs/static/img/favicon.ico",
		BundleURL:    "https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js",
		PresetURL:    "https://unpkg.com/swagger-ui-dist/swagger-ui-standalone-preset.js",
		StylesURL:    "https://unpkg.com/swagger-ui-dist/swagger-ui.css",
		Spec:         string(json),
	}
}

type TemplateRenderer struct {
	templates *template.Template
}

func (templateRender *TemplateRenderer) Render(writer io.Writer, name string, data interface{}, ctx echo.Context) error {
	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = ctx.Echo().Reverse
	}
	return templateRender.templates.ExecuteTemplate(writer, name, data)
}

func ServeSwaggerUI(echoRouter *echo.Echo, groupPath string, uri string, opts UIConfig) {
	const template_name = "swaggerui"

	tmpl := TemplateRenderer{
		templates: template.Must(template.New(template_name).Parse(uiTemplate)),
	}
	echoRouter.Renderer = &tmpl

	echoRouter.Group(groupPath).GET(uri, func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, template_name, opts)
	})
}
