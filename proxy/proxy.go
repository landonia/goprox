// Copyright 2016 Landonia Ltd. All rights reserved.

package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/landonia/golog"
)

const (
	// DefaultServerHostname returns the default hostname which is 0.0.0.0
	DefaultServerHostname = "0.0.0.0"
	// DefaultServerPort returns the default port which is 8080, not used
	DefaultServerPort = 8080
)

var (
	logger = golog.New("proxy.Proxy")
)

var (
	// DefaultServerAddr the default server addr which is: 0.0.0.0:8080
	DefaultServerAddr = DefaultServerHostname + ":" + strconv.Itoa(DefaultServerPort)
)

// Proxy is the root server
type Proxy struct {
	rs           *http.Server  // The actual server
	config       Configuration // The configuration
	proxies      []*CP         // The proxies to the host->proxy
	proxyHandler http.Handler  // The root proxy handler
	exit         chan error    // When to shutdown the server
}

// CP holds the proxy and the path being proxied
type CP struct {
	Path  string                 // The path being proxied
	Proxy *httputil.ReverseProxy // The actual proxy
}

// NewReverseRewriteProxy returns a new ReverseProxy that rewrites
// URLs to the scheme, host, and base path provided in target. It will rewrite
// URL to match the request path
func NewReverseRewriteProxy(source string, target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		// We need to extract the rest of the path from the source and then
		// forward this to target location

		if v := strings.Split(req.URL.Path, source); len(v) == 2 {
			req.URL.Path = singleJoiningSlash(target.Path, v[1])
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// Setup will initialise the proxy and must be called before any other functions
func Setup(config Configuration) (*Proxy, error) {
	gm := &Proxy{}
	gm.config = config
	gm.proxies = make([]*CP, 0, len(config.Proxies))

	// If there are any proxies then we need to set them up as well
	for _, proxy := range config.Proxies {
		if u, err := url.Parse(proxy.To); err == nil {
			gm.proxies = append(gm.proxies, &CP{
				Path:  proxy.ProxyPath,
				Proxy: NewReverseRewriteProxy(proxy.ProxyPath, u),
			})
		} else {
			logger.Warn("Could not parse proxy to address: %s", err.Error())
		}
	}

	// Create the root handler
	gm.proxyHandler = http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

		// Check if the proxy exists for this path
		found := false
		for _, proxy := range gm.proxies {
			if strings.HasPrefix(req.URL.Path, proxy.Path) {
				logger.Trace("Proxy: %v: to Path: %s", proxy.Path, req.URL.String())

				// Forward to the proxy
				proxy.Proxy.ServeHTTP(resp, req)
				found = true
				break
			}
		}

		if !found {
			if gm.config.StaticDir != "" {
				filePath := path.Join(gm.config.StaticDir, req.URL.EscapedPath())
				logger.Trace("Serve: %v: Path: %s", req.Host, filePath)

				// Just attempt to serve the file/directory specified by the host
				http.ServeFile(resp, req, filePath)
			} else {
				logger.Trace("Serve: %v: Notfound: %s", req.Host, req.URL.String())
				resp.WriteHeader(http.StatusNotFound)
			}
		}
	})
	return gm, nil
}

// Service will start the server and handle the requests
func (gm *Proxy) Service() (err error) {

	// Initialise the server if one has not been provided
	gm.rs = &http.Server{
		Addr:    gm.config.Addr,
		Handler: gm.proxyHandler,
	}

	// Attempt to start the service
	if gm.rs == nil {
		err = fmt.Errorf("Setup() must be called")
	} else {
		logger.Info("Starting Proxy server at address: %s", gm.config.Addr)
		gm.exit = make(chan error)

		// Launch the server
		go func() {
			gm.exit <- gm.Listen()
		}()

		// Block until we receive the exit
		err = <-gm.exit
		logger.Info("Proxy server has shutdown at address: %s", gm.config.Addr)
	}
	return
}

// Listen will create the handler to listen for requests and proxy them accordingly
func (gm *Proxy) Listen() error {
	addr := ParseHost(gm.config.Addr)
	logger.Info("Address: %s", addr)
	return gm.rs.ListenAndServe()
}

// Shutdown will force the Service function to exit
func (gm *Proxy) Shutdown() {
	gm.exit <- nil
}

// ParseHost tries to convert a given string to an address which is compatible with net.Listener and server
func ParseHost(addr string) string {
	// check if addr has :port, if not do it +:80 ,we need the hostname for many cases
	a := addr
	if a == "" {
		// check for os environments
		if oshost := os.Getenv("ADDR"); oshost != "" {
			a = oshost
		} else if oshost := os.Getenv("HOST"); oshost != "" {
			a = oshost
		} else if oshost := os.Getenv("HOSTNAME"); oshost != "" {
			a = oshost
			// check for port also here
			if osport := os.Getenv("PORT"); osport != "" {
				a += ":" + osport
			}
		} else if osport := os.Getenv("PORT"); osport != "" {
			a = ":" + osport
		} else {
			a = DefaultServerAddr
		}
	}
	if portIdx := strings.IndexByte(a, ':'); portIdx == 0 {
		if a[portIdx:] == ":https" {
			a = DefaultServerHostname + ":443"
		} else {
			// if contains only :port	,then the : is the first letter, so we dont have setted a hostname, lets set it
			a = DefaultServerHostname + a
		}
	}

	return a
}
