package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
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
		BlockByDefault: ipDefaultEnabled || ipUrlFilterEnabled || ipRedisEnabled,
		AllowedIPs:     defaultIps,
		TrustProxy:     true,
	})

	// Periodically check ip for time based
	var ips []string
	var blockedIps []string

	ipInterval := 300
	if len(conf.IpCheckInterval) > 0 {
		ipInterval, _ = strconv.Atoi(conf.IpCheckInterval)
	}

	ttl := time.Duration(ipInterval) * time.Second
	ticker := time.NewTicker(ttl)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// Get Ip from url if enabled
				if ipUrlFilterEnabled {
					urlres := ipsUrl(conf)
					ips = appendIp(ips, urlres)
					blockedIps = difference(ips, urlres)
					for _, item := range blockedIps {
						ips = FindAndDelete(ips, item)
					}
				}

				// Get Ip from redis if enabled
				if ipRedisEnabled {
					urlres := ipsRedis(conf)
					ips = appendIp(ips, urlres)
					blockedIps = difference(ips, urlres)
					for _, item := range blockedIps {
						ips = FindAndDelete(ips, item)
					}
				}

				// Assign as allowed ip
				for i := range ips {
					f.AllowIP(ips[i])
				}

				for i := range blockedIps {
					f.BlockIP(blockedIps[i])
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
		return []string{}
	}
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("error %v", err.Error())
		return []string{}
	}

	responseString := string(responseData)

	ips := strings.Split(responseString, "\n")

	return ips
}

func ipsRedis(conf *Config) []string {
	var redisIps []string
	redisIp, errRedis := GetKeyField(conf.Redis.Key, conf.Redis.Field)
	if errRedis != nil {
		log.Error("error get key field %v", errRedis)
	} else {
		match, _ := regexp.MatchString("^[0-9.,/]*$", redisIp)
		if !match {
			log.Error("error %s redis not valid. eg : 10.0.0.1,10.0.0.2,10.0.0/4", conf.Redis.Field)
		} else {
			redisIps := strings.Split(redisIp, ",")
			return redisIps
		}
	}

	return redisIps
}

func appendIp(ips []string, newsIp []string) []string {

	for _, item := range newsIp {
		if !Contains(ips, item) {
			ips = append(ips, item)
		}
	}

	return ips
}

func difference(ips, newsIp []string) []string {
	var diff []string
	for i := 0; i < 2; i++ {
		for _, s1 := range ips {
			found := false
			for _, s2 := range newsIp {
				if s1 == s2 {
					found = true
					break
				}
			}

			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			ips, newsIp = newsIp, ips
		}
	}
	return diff
}
