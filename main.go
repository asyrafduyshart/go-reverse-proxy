package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/asyrafduyshart/go-reverse-proxy/log"
)

// Config ...
type Config struct {
	AccessLog          string `json:"access_log"`
	LogLevel           string `json:"log_level"`
	IpWhiteListUrl     string `json:"ip_whitelist_url"`
	IpCheckInterval    string `json:"ip_check_interval"`
	DefaultIpWhitelist string `json:"default_ip_whitelist"`
	HTTP               struct {
		Servers []Server `json:"servers"`
	}
	Redis struct {
		Url   string `json:"url"`
		Key   string `json:"key"`
		Field string `json:"field"`
	}
}

const (
	// ProjectName ...
	ProjectName = "Goinx"
	// Version ...
	Version = "0.0.1"
	// PidFile ...
	PidFile = "goinx.pid"
)

var (
	configPath = flag.String("config", "config.json", "Configuration Path")
	cmds       = []string{"start", "stop", "restart"}
)

func usage() {
	fmt.Printf("💖  %s %s\n", ProjectName, Version)
	fmt.Println("Author: Asyraf Duyshart")
	fmt.Println("Github: https://github.com/asyrafduyshart/go-reverse-proxy")
	fmt.Printf("\nUsage: goinx [start|stop|restart]\n")
	fmt.Println("Options:")
	fmt.Println("    --config\tConfiguration path")
	fmt.Println("    --help\tHelp info")
}

func startArgs() *Config {
	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}

	cmd := os.Args[1]

	if !Contains(cmds, cmd) {
		usage()
		os.Exit(0)
	}

	// start goinx
	if cmd == cmds[0] {
		return start()
	}
	// stop goinx
	if cmd == cmds[1] {
		stop()
	}
	if cmd == cmds[2] {
		stop()
		return start()
	}

	return nil
}

func start() *Config {

	if Exist(PidFile) {
		log.Warning("Goinx has bean started.")
		os.Exit(0)
	}

	conf := Config{}
	if pid := os.Getpid(); pid != 1 {
		err := ioutil.WriteFile(PidFile, []byte(strconv.Itoa(pid)), 0777)
		if err != nil {
			log.Error("%v", err)
		}
	}

	flag.Usage = usage
	flag.Parse()

	var bytes []byte

	// Take length of string with len.

	if mp := os.Getenv("CONFIG_SETTING"); mp != "" {
		bytes = []byte(mp)
	} else {
		cl := *configPath

		if len(os.Args) > 3 && os.Args[2] == "--config" {
			if os.Args[3] != "" {
				cl = os.Args[3]
			}
		}

		result, err := ioutil.ReadFile(cl)
		if err != nil {
			log.Error("%v", err)
			os.Remove("goinx.pid")
			os.Exit(0)
		}
		bytes = result
	}

	err := json.Unmarshal([]byte(bytes), &conf)
	if err != nil {
		log.Error("%v", err)
		os.Remove("goinx.pid")
		os.Exit(0)
	}
	return &conf
}

func stop() {
	bytes, err := ioutil.ReadFile(PidFile)
	if err != nil {
		log.Error("%v", err)
		os.Exit(0)
	}

	pid, err := strconv.Atoi(string(bytes))
	log.Info("Stopping " + strconv.Itoa(pid))

	if err != nil {
		log.Error("%v", err)
	}
	os.Exit(0)
	os.Remove("goinx.pid")
}

func shutdownHook() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				os.Remove("goinx.pid")
				log.Info("Shutown Goinx.")
				os.Exit(0)
			default:
				log.Info("other", s)
			}
		}
	}()
}

func main() {

	shutdownHook()
	conf := startArgs()
	log.Info("Start Goinx.")
	if len(conf.Redis.Url) != 0 {
		RedisInit(conf.Redis.Url)
	}

	if conf.LogLevel == "debug" {
		log.LogLevelNum = 1
	}
	if conf.LogLevel == "info" {
		log.LogLevelNum = 2
	}
	if conf.LogLevel == "warn" {
		log.LogLevelNum = 3
	}
	if conf.LogLevel == "error" {
		log.LogLevelNum = 4
	}

	ipf := InitIpFilter(conf)

	count := 0
	exitChan := make(chan int)
	for _, server := range conf.HTTP.Servers {
		go func(s Server) {
			// loop through
			s.Start(conf, ipf)
			exitChan <- 1
		}(server)
		count++
	}

	for i := 0; i < count; i++ {
		<-exitChan
	}

}
