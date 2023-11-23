package vi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/diontr00/vi/internal/color"
)

type method string

const (
	MethodGet     method = method(http.MethodGet)
	MethodPost    method = method(http.MethodPost)
	MethodPut     method = method(http.MethodPut)
	MethodDelete  method = method(http.MethodDelete)
	MethodPatch   method = method(http.MethodPatch)
	MethodConnect method = method(http.MethodConnect)
	MethodHead    method = method(http.MethodHead)
	MethodTrace   method = method(http.MethodTrace)
)

type Config struct {
	banner bool
}

type middleware func(next http.HandlerFunc) http.HandlerFunc

type vi struct {
	prefixes []string
	// routing tree
	trees map[method]*tree
	// map between prefix and middleware
	middlewares map[string][]middleware
}

// Return new vi
func New(config *Config) *vi {
	if config != nil && config.banner {
		fmt.Println(color.Green(banner, color.Blue(Version), color.Red(website)))
	}

	prefix := "/"
	middlewares := make(map[string][]middleware)
	middlewares[prefix] = []middleware{}
	return &vi{
		prefixes:    []string{prefix},
		middlewares: middlewares,
		trees:       make(map[method]*tree),
	}
}

// HTTP get routing along "pattern"
func (v *vi) GET(path string, handler http.HandlerFunc) {
	v.Add(MethodGet, path, handler)
}

// HTTP post routing along "pattern"
func (v *vi) POST(path string, handler http.HandlerFunc) {
	v.Add(MethodPost, path, handler)
}

// HTTP put routing along "pattern"
func (v *vi) PUT(path string, handler http.HandlerFunc) {
	v.Add(MethodPut, path, handler)
}

// HTTP delete routing along "pattern"
func (v *vi) DELETE(path string, handler http.HandlerFunc) {
	v.Add(MethodDelete, path, handler)
}

// HTTP path routin  along "pattern"
func (v *vi) PATCH(path string, handler http.HandlerFunc) {
	v.Add(MethodPatch, path, handler)
}

// register new  HTTP verb routing along pattern
func (v *vi) Add(m method, path string, handler http.HandlerFunc) {
	if path == "" {
		path = "/"
	}
	if string(path[0]) != "/" {
		path = "/" + path
	}

	if v.trees == nil {
		v.trees = make(map[method]*tree)
	}

	tree := v.trees[method(m)]
	if tree == nil {
		tree = newTree()
		v.trees[method(m)] = tree
	}

	tree.add(path, handler, v.prefixes)
}

// use to group route under prefix
func (v *vi) Group(prefix string) *vi {
	if string(prefix[0]) != "/" {
		prefix = "/" + prefix
	}
	prefixes := v.prefixes
	if _, ok := v.middlewares[prefix]; !ok {
		v.middlewares[prefix] = make([]middleware, 0)
		prefixes = append(prefixes, prefix)
	}

	return &vi{
		prefixes:    prefixes,
		trees:       v.trees,
		middlewares: v.middlewares,
	}
}

// use to register middlewares
func (v *vi) Use(middlewares ...middleware) {
	prefix := v.prefixes[len(v.prefixes)-1]

	if len(middlewares) > 0 {
		v.middlewares[prefix] = append(v.middlewares[prefix], middlewares...)
	}
}

func (v *vi) chain(w http.ResponseWriter, r *http.Request, handler http.HandlerFunc, prefixes []string) {
	for _, p := range prefixes {
		for _, m := range v.middlewares[p] {
			handler = m(handler)
		}
	}

	handler(w, r)
}

func (v *vi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rqUrl := r.URL.Path
	nodes := v.trees[method(r.Method)].find(rqUrl)
	for i := range nodes {
		handler := nodes[i].handler
		if handler != nil {
			if nodes[i].path == rqUrl {
				v.chain(w, r, handler, nodes[i].prefixes)
				return
			}
		}
	}

	if nodes == nil {
		// match against any static match
		nodes := v.trees[method(r.Method)].find("/")

		for i := range nodes {
			handler := nodes[i].handler

			if handler != nil && nodes[i].path != rqUrl {
				isMatch, matchParams := match(rqUrl, nodes[i].path)
				if isMatch {
					for k, v := range matchParams {
						ctx := context.WithValue(r.Context(), k, v)
						r = r.WithContext(ctx)
					}
					v.chain(w, r, handler, nodes[i].prefixes)
					return
				}
			}
		}
	}

	http.NotFound(w, r)
}
