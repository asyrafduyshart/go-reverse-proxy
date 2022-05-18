package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	log "github.com/asyrafduyshart/go-reverse-proxy/log"
	"github.com/jpillora/ipfilter"
)

func InitIpFilter(conf *Config) *ipfilter.IPFilter {

	ipUrlFilterEnabled := len(conf.IpWhiteListUrl) != 0
	ipDefaultEnabled := len(conf.DefaultIpWhitelist) != 0

	defaultIps := strings.Split(conf.DefaultIpWhitelist, ",")

	redisUrlStatus := len(conf.Redis.Url) != 0
	redisKeyStatus := len(conf.Redis.Key) != 0
	redisFieldStatus := len(conf.Redis.Field) != 0
	ipRedisEnabled := false

	if redisUrlStatus {
		if !(redisKeyStatus && redisFieldStatus) {
			log.Error("error %v", "Please set redis config properly")
		} else {
			ipRedisEnabled = true
		}
	}

	f := ipfilter.New(ipfilter.Options{
		BlockByDefault: ipDefaultEnabled,
		AllowedIPs:     defaultIps,
		TrustProxy:     true,
	})

	// Periodically check ip for time based
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		var ips []string
		for {
			select {
			case <-ticker.C:
				// Get Ip from url if enabled
				if ipUrlFilterEnabled {
					urlres := ipsUrl(conf)
					ips = append(ips, urlres...)
				}

				// Get Ip from redis if enabled
				if ipRedisEnabled {
					urlres := ipsRedis(conf)
					ips = append(ips, urlres...)
				}

				// Assign as allowed ip
				for i := range ips {
					f.AllowIP(ips[i])
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return f
}

func ipsUrl(conf *Config) []string {
	resp, err := http.Get(conf.IpWhiteListUrl)
	if err != nil {
		log.Error("error %v", err)
	}

	// read json http response
	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("error %v", err)
	}
	var ips []string

	err = json.Unmarshal([]byte(jsonDataFromHttp), &ips)
	if err != nil {
		log.Error("error %v", err)
	}

	return ips
}

func ipsRedis(conf *Config) []string {
	redisIp, errRedis := GetKeyField(conf.Redis.Key, conf.Redis.Field)
	if errRedis != nil {
		log.Error("error %v", errRedis)
	}

	match, _ := regexp.MatchString("^[0-9.,/]*$", redisIp)

	if !match {
		log.Error("error %v", "redis "+conf.Redis.Field+" not valid. eg : 10.0.0.1,10.0.0.2,10.0.0/4")
	}

	redisIps := strings.Split(redisIp, ",")

	return redisIps
}
