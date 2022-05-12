# goinx

Multiple domain reverse proxy with golang

## Feature

- Support static server
- Support multi domain proxy
- Support HTTPS

## Usage

**By Golang**

## Build File
```bash
go install

go build //Windows

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build // To run in alpine

```


## Run File

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

```json
{
    "log_level": "debug",
    "access_log": null,
    "redis" : {
        "url" : "redis://localhost:6379",
        "key" : "user:config",
        "field": "ip_whitelist"
    },
    "http": {
        "servers": [
            {
                "name": "site2"
            },
            {
                "name": "site3",
                "listen": "9003",
                "files": [
                    {
                        "path" : "example-path",
                        "location" : "index-path",
                        "index" : "index.html"
                    }
                ]
            },
            {
                "name": "site",
                "listen": "9001",
                "proxies": [
                    {
                        "proxy_pass": "https://httpbin.org",
                        "proxy_path": "/httpbin",
                        "request_headers": [
                            {
                                "Authentication": "Basic 123456"
                            },
                            {
                                "Yomama": "Pk"
                            }
                        ]
                    },
                    {
                        "proxy_pass": "https://postman-echo.com",
                        "proxy_path": "/postman",
                        "request_headers": [
                            {
                                "Authentication": "Basic 11235453"
                            }
                        ]
                    }
                ],
                "domains": [
                    "localhost"
                ],
                "root": "www"
            }
        ]
    }
}
```