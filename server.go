package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"

	log "github.com/asyrafduyshart/go-reverse-proxy/log"
)

// Server struct file
type Server struct {
	Name           string              `yaml:"name"`
	Listen         string              `yaml:"listen"`
	Domains        []string            `yaml:"domains"`
	Root           *string             `yaml:"root"`
	SSL            bool                `yaml:"ssl"`
	GZIP           bool                `yaml:"gzip"`
	GFW            bool                `yaml:"gfw"`
	ProxyPass      *string             `yaml:"proxy_pass"`
	ProxyPath      *string             `yaml:"proxy_path"`
	RequestHeaders []map[string]string `yaml:"request_headers"`
	KeyFile        string              `yaml:"key_file"`
	CertFile       string              `yaml:"cert_file"`
}

// Keys Return keys of the given map
func Keys(m map[string]string) map[string]interface{} {
	po := make(map[string]interface{})
	for k := range m {
		po[k] = m[k]
	}
	return po
}

// Start the server
func (s *Server) Start() {
	if s.ProxyPass != nil {
		log.Info("%s listen %s, ssl: %v, proxy to %s", s.Name, s.Listen, s.SSL, *s.ProxyPass)
	} else {
		if s.Root != nil {
			log.Info("%s listen %s, ssl: %v, static dir %s", s.Name, s.Listen, s.SSL, *s.Root)
		}
	}

	r := mux.NewRouter()

	api := r.PathPrefix(*s.ProxyPath).Subrouter()
	api.HandleFunc("", s.fuckGFW)
	api.HandleFunc("/", s.fuckGFW)
	api.HandleFunc(`/{rest:[a-zA-Z0-9=\-\/]+}`, s.fuckGFW)

	if s.Root != nil {
		pathLocation := *s.Root
		log.Info("Config location %s", pathLocation)
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("." + pathLocation)))
	}

	var err error
	if s.SSL {
		err = http.ListenAndServeTLS(s.Listen, s.CertFile, s.KeyFile, r)
	} else {
		err = http.ListenAndServe(s.Listen, r)
	}

	if err != nil {
		log.Error("%v", err)
	}
}

// Handler the request
func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", ProjectName+"/"+Version)
	if !Contains(s.Domains, strings.Replace(r.Host, s.Listen, "", -1)) {
		log.Error("Request [%s] RemoteAddr: %s, Header: %v", r.Host, r.RemoteAddr, r.Header)
		w.Write([]byte("Bad Request."))
		return
	}
	log.Info("Request URI: %s", r.URL.RequestURI())
	log.Info("Request [%s] RemoteAddr: %s, Header: %v", r.Host, r.RemoteAddr, r.Header)

	if s.GFW {
		s.fuckGFW(w, r)
	} else {
		if s.ProxyPass == nil {
			s.Static(w, r)
		} else {
			s.Proxy(w, r)
		}
	}
}

// Static server
func (s *Server) Static(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:]
	data, err := ioutil.ReadFile(*s.Root + "/" + string(path))
	if err == nil {
		var contentType string
		if strings.HasSuffix(path, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(path, ".html") {
			contentType = "text/html"
		} else if strings.HasSuffix(path, ".js") {
			contentType = "application/javascript"
		} else if strings.HasSuffix(path, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(path, ".jpg") {
			contentType = "image/jepg"
		} else if strings.HasSuffix(path, ".jepg") {
			contentType = "image/jepg"
		} else if strings.HasSuffix(path, ".svg") {
			contentType = "image/svg+xml"
		} else {
			contentType = "text/plain"
		}

		w.Header().Add("Content Type", contentType)
		w.Write(data)
	} else {
		w.WriteHeader(404)
		w.Write([]byte("404 My dear - " + http.StatusText(404)))
	}
}

// IndexHandler index handler
func (s *Server) IndexHandler() func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		location := *s.Root
		log.Info("Web Location %s", location)
		http.ServeFile(w, r, location)
	}
	return http.HandlerFunc(fn)
}

// Proxy server
func (s *Server) Proxy(w http.ResponseWriter, r *http.Request) {
	realurl := *s.ProxyPass + r.RequestURI
	log.Info("RealURL: %s", realurl)

	o := new(http.Request)

	*o = *r
	targetURL, err := url.Parse(*s.ProxyPass)
	log.Info("Target Url: %s", targetURL)

	o.Host = targetURL.Host
	o.URL.Scheme = targetURL.Scheme
	o.URL.Host = targetURL.Host
	o.URL.Path = singleJoiningSlash(targetURL.Path, o.URL.Path)

	if q := o.URL.RawQuery; q != "" {
		o.URL.RawPath = o.URL.Path + "?" + q
	} else {
		o.URL.RawPath = o.URL.Path
	}

	o.URL.RawQuery = targetURL.RawQuery

	o.Proto = "HTTP/1.1"
	o.ProtoMajor = 1
	o.ProtoMinor = 1
	o.Close = false

	transport := http.DefaultTransport
	res, err := transport.RoundTrip(o)

	if err != nil {
		log.Error("http: proxy error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hdr := w.Header()

	for k, vv := range res.Header {
		for _, v := range vv {
			hdr.Add(k, v)
		}
	}

	for _, c := range res.Cookies() {
		hdr.Add("Set-Cookie", c.Raw)
	}

	if s.GZIP {
		hdr.Add("Content-Encoding", "gzip")
	}

	w.WriteHeader(res.StatusCode)

	if res.Body != nil {
		io.Copy(w, res.Body)
	}

}

var transport = &http.Transport{
	ResponseHeaderTimeout: 30 * time.Second,
	TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
}

func (s *Server) fuckGFW(w http.ResponseWriter, r *http.Request) {
	realurl := *s.ProxyPass + r.RequestURI
	log.Info("RealURL: %s", realurl)

	req, err := http.NewRequest(r.Method, realurl, r.Body)
	req.Header.Set("Accept", r.Header.Get("Accept"))
	req.Header.Set("Accept-Encoding", r.Header.Get("Accept-Encoding"))
	req.Header.Set("Accept-Language", r.Header.Get("Accept-Language"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", r.Header.Get("User-Agent"))
	req.Header.Set("Cookie", r.Header.Get("Cookie"))

	for k := range s.RequestHeaders {
		headerReq := Keys(s.RequestHeaders[k])
		for j := range headerReq {
			log.Info("Header: %s | Value: %s", j, headerReq[j])
			str := fmt.Sprintf("%v", headerReq[j])
			req.Header.Set(j, str)
		}
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		log.Error("%v", err)
		return
	}

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		log.Info("Location: %s", location)
		req, err = http.NewRequest(r.Method, location, r.Body)
		req.Header.Set("Accept", r.Header.Get("Accept"))
		req.Header.Set("Accept-Encoding", r.Header.Get("Accept-Encoding"))
		req.Header.Set("Accept-Language", r.Header.Get("Accept-Language"))
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Pragma", "no-cache")
		req.Header.Set("User-Agent", r.Header.Get("User-Agent"))
		req.Header.Set("Cookie", r.Header.Get("Cookie"))

		resp, err = transport.RoundTrip(req)
		if err != nil {
			log.Error("%v", err)
			return
		}
	}

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", resp.Header.Get("Content-Type"))

	if err != nil {
		log.Error("%v", err)
		return
	}

	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	w.Write(data)

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
