package app

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

//go:embed page.html.gotpl
var tplStrPage string
var tplPage = template.New("page")

// initialize page template
func init() {
	tplPage = tplPage.Funcs(map[string]interface{}{
		"asjson": func(vars interface{}) string {
			b, err := json.Marshal(vars)
			if err != nil {
				panic(err)
			}
			return string(b)
		},
		"unsafe": func(str string) template.HTML {
			return template.HTML(str)
		},
		"unsafeAttr": func(str string) template.HTMLAttr {
			return template.HTMLAttr(str)
		},
	})

	tplPage = template.Must(tplPage.Parse(tplStrPage))
}

// HTML preload links
type Preload struct {
	Href string
	As   string
	Type string
	Flag string
}

type PageConfig struct {
	Route   string
	Script  string
	Style   string
	Vendor  string
	Preload []*Preload
	Inject  []string
	Data    interface{}
}

func (c *PageConfig) TemplateVariables(ctx context.Context, dist string) map[string]interface{} {
	vars := map[string]interface{}{}

	if c.Script != "" {
		vars["js"] = filepath.Join(dist, c.Script)
	}
	if c.Style != "" {
		vars["css"] = filepath.Join(dist, c.Style)
	}
	if c.Vendor != "" {
		vars["vendor"] = filepath.Join(dist, c.Vendor)
	}
	if len(c.Preload) > 0 {
		modifiedPreload := make([]*Preload, len(c.Preload))
		for i, curr := range c.Preload {
			modifiedPreload[i] = &Preload{
				Href: filepath.Join(dist, curr.Href),
				As:   curr.As,
				Type: curr.Type,
				Flag: curr.Flag,
			}
		}
		vars["preload"] = modifiedPreload
	}

	if len(c.Inject) > 0 {
		vars["inject"] = c.Inject
	}

	if mode := ctx.Value(CtxMode{}); mode != nil {
		vars["mode"] = fmt.Sprint(mode)
	}
	if version := ctx.Value(CtxVersion{}); version != nil {
		vars["version"] = fmt.Sprint(version)
	}
	if c.Data != nil {
		vars["data"] = c.Data
	}

	return vars
}

func (c *PageConfig) Reader(ctx context.Context, dist string) (io.Reader, error) {
	w := bytes.NewBuffer([]byte{})
	if err := tplPage.Execute(w, c.TemplateVariables(ctx, dist)); err != nil {
		return nil, err
	}
	return w, nil
}

func page(config *PageConfig, dist string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tplPage.Execute(w, config.TemplateVariables(r.Context(), dist)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
