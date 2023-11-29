package vi

import (
	"context"
	"fmt"
	"net/http"

	"github.com/diontr00/vi/internal/color"
)

// type of the context key
type ctxKey struct{}

// use to set the key of a context
var contextKey = ctxKey{}

type Config struct {
	// When set to false , this will turn off the banner
	Banner bool
	// Set the custom not found error handler , if not set the default not fault will be use
	NotFoundHandler http.HandlerFunc
}

type middleware func(next http.HandlerFunc) http.HandlerFunc

type vi struct {
	// hold prefix that relevant for particular vi instance
	prefixes []string
	// routing tree
	trees map[string]*tree
	// map between prefix and middleware
	middlewares map[string][]middleware
	// not found error handler
	notfoundhandler http.HandlerFunc
}

// Return new vi
func New(config *Config) *vi {
	v := new(vi)
	v.prefixes = []string{"/"}
	v.middlewares = map[string][]middleware{"/": {}}
	v.trees = make(map[string]*tree)

	if config != nil && config.Banner {
		fmt.Println(color.Green(banner, color.Blue(Version), color.Red(website)))
	}
	if config.NotFoundHandler != nil {
		v.notfoundhandler = config.NotFoundHandler
	} else {
		v.notfoundhandler = func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}
	return v
}

// HTTP get routing along "pattern"
func (v *vi) GET(path string, handler http.HandlerFunc) {
	v.Add("GET", path, handler)
}

// HTTP post routing along "pattern"
func (v *vi) POST(path string, handler http.HandlerFunc) {
	v.Add("POST", path, handler)
}

// HTTP put routing along "pattern"
func (v *vi) PUT(path string, handler http.HandlerFunc) {
	v.Add("PUT", path, handler)
}

// HTTP delete routing along "pattern"
func (v *vi) DELETE(path string, handler http.HandlerFunc) {
	v.Add("DELETE", path, handler)
}

// HTTP path routin  along "pattern"
func (v *vi) PATCH(path string, handler http.HandlerFunc) {
	v.Add("PATCH", path, handler)
}

// register new  HTTP verb routing along pattern
func (v *vi) Add(method, path string, handler http.HandlerFunc) {
	if method == "" {
		panic(color.Red("method must not be empty"))
	}

	if len(path) < 1 {
		panic(color.Red("path must not be empty"))
	}
	if string(path[0]) != "/" {
		path = "/" + path
	}
	if handler == nil {
		panic(color.Red("handler must not be nil"))
	}

	if v.trees == nil {
		v.trees = make(map[string]*tree)
	}

	tree := v.trees[method]
	if tree == nil {
		tree = newTree()
		v.trees[method] = tree
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
		prefixes:        prefixes,
		trees:           v.trees,
		middlewares:     v.middlewares,
		notfoundhandler: v.notfoundhandler,
	}
}

// use to register middlewares
func (v *vi) Use(middlewares ...middleware) {
	prefix := v.prefixes[len(v.prefixes)-1]

	if len(middlewares) > 0 {
		v.middlewares[prefix] = append(v.middlewares[prefix], middlewares...)
	}
}

// chain all middlewares associate with prefixes
func (v *vi) chain(w http.ResponseWriter, r *http.Request, handler http.HandlerFunc, prefixes []string) {
	var allMiddleware []middleware
	for _, p := range prefixes {
		allMiddleware = append(allMiddleware, v.middlewares[p]...)
	}

	for i := len(allMiddleware) - 1; i >= 0; i-- {
		handler = allMiddleware[i](handler)
	}

	handler(w, r)
}

// Get the matched  param that store inside request context
func GetParam(r *http.Request, key string) (paramValue string) {
	values, ok := r.Context().Value(contextKey).(matchParams)
	if ok {
		paramValue, ok = values[matchKey(key)]
	}

	if !ok {
		paramValue = ""
	}

	return paramValue
}

func (v *vi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rqUrl := r.URL.Path
	nodes := v.trees[r.Method].find(rqUrl)
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
		// match against any regex match
		nodes := v.trees[r.Method].find("/")

		for i := range nodes {
			handler := nodes[i].handler

			if handler != nil {
				isMatch, matchParams := match(rqUrl, nodes[i].path)
				if isMatch {
					ctx := context.WithValue(r.Context(), contextKey, matchParams)
					r = r.WithContext(ctx)
					v.chain(w, r, handler, nodes[i].prefixes)
					return
				}
			}
		}
	}

	v.notfoundhandler(w, r)
}
