package main

import (
	"crypto/tls"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	log "github.com/asyrafduyshart/go-reverse-proxy/log"
)

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
