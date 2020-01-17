# goinx

Multiple domain reverse proxy with golang

## Feature

- Support static server
- Support multi domain proxy
- Support HTTPS

## Usage

**By Golang**

```bash
» ./goinx
Author: asyrafduyshart
Github: https://github.com/asyrafduyshart/go-reverse-proxy

Usage: goinx [start|stop|restart]

Options:

    --config    Configuration path
    --help      Help info
```

## Config File

```yaml
log_level: info
access_log:
http:
  servers:
    - name: site1
      listen: ":9001"
      domains: [localhost]
      proxy_pass: http://httpbin.org/get
      cert_file:
      key_file:
    - name: site2
      listen: ":9002"
      gfw: true
      domains: [http://nmhclbwvtxzkdfsr.neverssl.com/]
      proxy_pass: "http://nmhclbwvtxzkdfsr.neverssl.com/"
```
