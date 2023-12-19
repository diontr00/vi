package vi

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/diontr00/vi/internal/color"
	"github.com/diontr00/vi/internal/utils"
	"io/fs"
	"log"
	"net/http"
	ospath "path"
	"strconv"
	"strings"
)

// type of the context key
type ctxKey struct{}

// use to set the key of a context
var contextKey = ctxKey{}

// Configuration for vi instance
type Config struct {
	// When set to false , this will turn off the banner
	Banner bool
	// Set the custom not found error handler , if not set the default not fault will be use
	NotFoundHandler http.HandlerFunc
}

// Static defines configuration options when defining static route
type StaticConfig struct {
	// Root is a filesystem  that provides access to a
	// collection of files and directories , use http.Dir("folder name") or embed FS  with http.FS()
	// Required
	Root http.FileSystem
	// Defines prefix that will be add to be when reading a file from FileSystem
	// Use only when using go embed FS for Root
	// Optional  default to ""
	Prefix string
	// Name of the index file for serving
	// Optional default to index.html
	Index string
	// The value for the cache-control HTTP-Header when response , its define in term of second , default value to 0
	// Optional default to 0
	MaxAge int
	//  Next defines a function  that allow to skip a scenario when it return true
	// Optional default to nil
	Next func(w http.ResponseWriter, r *http.Request) bool
	// File to return if path is not found. Useful for SPA's
	// Optional default to 404 not found
	NotFoundFile string
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

func (v *vi) registerStatic(path string, cfg *StaticConfig) {
	cacheControl := "public, max-age=" + strconv.Itoa(cfg.MaxAge)

	handler := func(w http.ResponseWriter, r *http.Request) {
		if cfg.Next != nil && cfg.Next(w, r) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// path to search for static files
		var searchp string = ospath.Clean(r.URL.Path)

		if cfg.Prefix != "" {
			searchp = cfg.Prefix + searchp
		}
		if len(searchp) > 1 {
			searchp = strings.TrimSuffix(searchp, "/")
		}

		if cfg.MaxAge > 0 {
			w.Header().Set("Cache-Control", cacheControl)
		}

		file, err := cfg.Root.Open(searchp)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "text/plain")
				if cfg.NotFoundFile == "" {
					w.Write([]byte("404 Not Found"))
					return
				}

				nffile, err := cfg.Root.Open(cfg.NotFoundFile)
				if err != nil {
					w.Write([]byte("404 Not Found"))
					log.Print(color.Red("[Warning] , not found file couldn't be open : %v \n", err))

					return
				}

				defer nffile.Close()
				w.Header().Set("Content-Type", utils.GetMIME(utils.GetFileExtension(cfg.NotFoundFile)))
				bufio.NewReader(nffile).WriteTo(w)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "text/plain")
				w.Write([]byte("500 server internal error"))
				log.Print(color.Red("[Error] , could not open file %s for reading : %v \n", color.Bold(searchp), err))
				return
			}
		}

		defer file.Close()
		stat, _ := file.Stat()

		// Serve index if path is directory
		if stat.IsDir() {
			index, err := cfg.Root.Open(cfg.Index)
			if err != nil {
				log.Panicf(color.Red("Index file couldn't be open : %v \n", err))
			}

			defer index.Close()
			file = index
			searchp = cfg.Index
		}

		mimeType := utils.GetFileExtension(searchp)
		w.Header().Set("Content-Type", utils.GetMIME(mimeType))
		if cfg.MaxAge > 0 {
			w.Header().Set("Cache-Control", cacheControl)
		}

		bufio.NewReader(file).WriteTo(w)
	}
	v.Add(http.MethodGet, path, handler)
}

// Static will create a file server serving the static file
// if path present the path pattern
func (v *vi) Static(path string, config *StaticConfig) {
	var cfg = &StaticConfig{
		Index:        "/index.html",
		MaxAge:       0,
		Next:         nil,
		NotFoundFile: "",
		Root:         nil,
		Prefix:       "",
	}

	if config != nil {
		cfg = config
		if config.Index == "" {
			cfg.Index = "index.html"
		}
		if !strings.HasPrefix(cfg.Index, "/") {
			cfg.Index = "/" + cfg.Index
		}
		if cfg.NotFoundFile != "" && !strings.HasPrefix(cfg.NotFoundFile, "/") {
			cfg.NotFoundFile = "/" + cfg.NotFoundFile
		}
	}

	if cfg.Root == nil {
		panic("Http file server root cannot be nil")
	}

	if cfg.Prefix != "" && !strings.HasPrefix(cfg.Prefix, "/") {
		cfg.Prefix = "/" + cfg.Prefix
	}
	v.registerStatic(path, cfg)
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
	// get the last prefix added
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
	if _, ok := v.trees[r.Method]; !ok {
		v.trees[r.Method] = newTree()
	}
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
