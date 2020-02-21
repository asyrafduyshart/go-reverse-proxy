package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	log "github.com/asyrafduyshart/go-reverse-proxy/log"
)

// Proxy struct file
type Proxy struct {
	ProxyPass      *string             `yaml:"proxy_pass"`
	ProxyPath      *string             `yaml:"proxy_path"`
	RequestHeaders []map[string]string `yaml:"request_headers"`
}

func (p Proxy) setup(w http.ResponseWriter, r *http.Request) {

	realurl := *p.ProxyPass + r.RequestURI
	log.Info("RealURL: %s", realurl)

	req, err := http.NewRequest(r.Method, realurl, r.Body)
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Println(name, value)
			req.Header.Set(name, value)
		}
	}

	for k := range p.RequestHeaders {
		headerReq := Keys(p.RequestHeaders[k])
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

// Server struct file
type Server struct {
	Name     string   `yaml:"name"`
	Listen   string   `yaml:"listen"`
	Domains  []string `yaml:"domains"`
	Root     *string  `yaml:"root"`
	SSL      bool     `yaml:"ssl"`
	GZIP     bool     `yaml:"gzip"`
	GFW      bool     `yaml:"gfw"`
	Proxies  []Proxy  `yaml:"proxies"`
	KeyFile  string   `yaml:"key_file"`
	CertFile string   `yaml:"cert_file"`
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

	if s.Root != nil {
		log.Info("%s listen %s, ssl: %v, static dir %s", s.Name, s.Listen, s.SSL, *s.Root)
	}

	r := mux.NewRouter()
	for _, proxy := range s.Proxies {
		if proxy.ProxyPass != nil {
			log.Info("%s listen %s, ssl: %v, proxy to %s", s.Name, s.Listen, s.SSL, *proxy.ProxyPass)
		}

		fmt.Println("PROXY", *proxy.ProxyPath)

		r.PathPrefix(*proxy.ProxyPath).Subrouter()
		r.PathPrefix(*proxy.ProxyPath).Subrouter().HandleFunc("", proxy.setup)
		r.PathPrefix(*proxy.ProxyPath).Subrouter().HandleFunc("/", proxy.setup)
		r.PathPrefix(*proxy.ProxyPath).Subrouter().HandleFunc(`/{rest:[a-zA-Z0-9=\-\/]+}`, proxy.setup)
	}

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

var transport = &http.Transport{
	ResponseHeaderTimeout: 30 * time.Second,
	TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
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
