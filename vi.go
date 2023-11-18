package vi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type method string

const (
	MethodGet    method = method(http.MethodGet)
	MethodPost   method = method(http.MethodPost)
	MethodPut    method = method(http.MethodPut)
	MethodDelete method = method(http.MethodDelete)
)

type vi struct {
	trees map[method]*tree
}

// Return new vi
func New() *vi {
	if boff := os.Getenv("BANNEROFF"); boff == "" {
		fmt.Printf(banner, Version, website)
	}

	return &vi{
		trees: make(map[method]*tree),
	}
}

// HTTP get routing along "pattern"
func (v *vi) GET(path string, handler http.HandlerFunc) {
	v.Add(http.MethodGet, path, handler)
}

// HTTP post routing along "pattern"
func (v *vi) POST(path string, handler http.HandlerFunc) {
	v.Add(http.MethodPost, path, handler)
}

// HTTP put routing along "pattern"
func (v *vi) PUT(path string, handler http.HandlerFunc) {
	v.Add(http.MethodPut, path, handler)
}

// HTTP delete routing along "pattern"
func (v *vi) DELETE(path string, handler http.HandlerFunc) {
	v.Add(http.MethodDelete, path, handler)
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

	tree.add(path, handler)
}

func (v *vi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rqUrl := r.URL.Path
	nodes := v.trees[method(r.Method)].find(rqUrl)

	for i := range nodes {
		handler := nodes[i].handler
		if handler != nil {
			if nodes[i].path == rqUrl {
				handler(w, r)
				return
			}
		}
	}

	if nodes == nil {
		res := strings.Split(rqUrl, "/")
		prefix := "/" + res[0]
		// match against any static match
		nodes := v.trees[method(r.Method)].find(prefix)

		for _, node := range nodes {
			handler := node.handler

			if handler != nil && node.path != rqUrl {
				isMatch, matchParams := match(rqUrl, node.path)
				if isMatch {
					for k, v := range matchParams {
						ctx := context.WithValue(r.Context(), k, v)
						r = r.WithContext(ctx)
					}
					handler(w, r)
					return
				}
			}
		}
	}

	http.NotFound(w, r)
}
