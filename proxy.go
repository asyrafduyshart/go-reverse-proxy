package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	log "github.com/asyrafduyshart/go-reverse-proxy/log"
)

// Proxy struct file
type Proxy struct {
	ProxyPass      *string             `json:"proxy_pass"`
	ProxyPath      *string             `json:"proxy_path"`
	RetainPath     *bool               `json:"retain_path" default:"false"`
	RequestHeaders []map[string]string `json:"request_headers"`
}

// IsRetainPath proxy path
func (p Proxy) IsRetainPath() bool {
	return p.RetainPath == nil || *p.RetainPath
}

func (p Proxy) getProxyPath(strEx string) string {
	reStr := regexp.MustCompile("^(.*?)" + *p.ProxyPath + "(.*)$")
	repStr := "${1}$2"
	return reStr.ReplaceAllString(strEx, repStr)
}

// setup create proxy
func (p Proxy) setup(w http.ResponseWriter, r *http.Request) {

	pathURL := *p.ProxyPass + p.getProxyPath(r.RequestURI)

	if !p.IsRetainPath() {
		pathURL = *p.ProxyPass + r.RequestURI
	}

	log.Info("Real URL Path: %s", r.RequestURI)
	log.Info("Proxied URL: %s", pathURL)

	req, err := http.NewRequest(r.Method, pathURL, r.Body)
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			// fmt.Println(name, value)
			req.Header.Set(name, value)
		}
	}

	for k := range p.RequestHeaders {
		headerReq := Keys(p.RequestHeaders[k])
		for j := range headerReq {
			str := fmt.Sprintf("%v", headerReq[j])
			req.Header.Set(j, str)
		}
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		log.Error("%v", err)
		return
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
