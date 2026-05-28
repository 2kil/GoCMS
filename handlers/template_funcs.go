package handlers

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"safe":  func(s string) template.HTML { return template.HTML(s) },
		"asset": assetPath,
		"dict": func(values ...interface{}) map[string]interface{} {
			m := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					if key, ok := values[i].(string); ok {
						m[key] = values[i+1]
					}
				}
			}
			return m
		},
		"list": func(values ...interface{}) []interface{} {
			return values
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},
		"md": func(s string) template.HTML {
			p := parser.NewWithExtensions(parser.CommonExtensions)
			renderer := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags})
			output := markdown.ToHTML([]byte(s), p, renderer)
			return template.HTML(output)
		},
		"render": renderTemplateString,
	}
}

func TemplateFuncsForGin() template.FuncMap {
	return templateFuncs()
}

func assetPath(s string) string {
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	return "/static/" + strings.TrimPrefix(s, "/")
}

func renderTemplateString(s string, data interface{}) template.HTML {
	tpl, err := template.New("column-content").Funcs(template.FuncMap{
		"safe":  func(s string) template.HTML { return template.HTML(s) },
		"asset": assetPath,
	}).Parse(s)
	if err != nil {
		return template.HTML(s)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return template.HTML(s)
	}
	return template.HTML(buf.String())
}
