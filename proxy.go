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
	ProxyPass      *string             `yaml:"proxy_pass"`
	ProxyPath      *string             `yaml:"proxy_path"`
	RetainPath     *bool               `default:"false" yaml:"retain_path"`
	RequestHeaders []map[string]string `yaml:"request_headers"`
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

	// if resp.StatusCode == 301 || resp.StatusCode == 302 {
	// 	location := resp.Header.Get("Location")
	// 	pathURL = location + replaceStringFirstOccurance(r.RequestURI, *p.ProxyPath, "")
	// 	for k, v := range resp.Header {
	// 		fmt.Print(k)
	// 		fmt.Print(" : ")
	// 		fmt.Println(v)
	// 	}
	// 	log.Info("Location: %s", location)
	// 	req, err = http.NewRequest(r.Method, pathURL, r.Body)
	// 	req.Header.Set("Accept", r.Header.Get("Accept"))
	// 	req.Header.Set("Accept-Encoding", r.Header.Get("Accept-Encoding"))
	// 	req.Header.Set("Accept-Language", r.Header.Get("Accept-Language"))
	// 	req.Header.Set("Cache-Control", "no-cache")
	// 	req.Header.Set("Pragma", "no-cache")
	// 	req.Header.Set("User-Agent", r.Header.Get("User-Agent"))
	// 	req.Header.Set("Cookie", r.Header.Get("Cookie"))

	// 	resp, err = transport.RoundTrip(req)
	// 	if err != nil {
	// 		log.Error("%v", err)
	// 		return
	// 	}
	// }

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
