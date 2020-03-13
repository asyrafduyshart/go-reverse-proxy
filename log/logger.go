package log

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevelNum level
var LogLevelNum = 2

const (
	colorRed = uint8(iota + 91)
	colorGreen
	colorYellow
	colorBlue
	colorMagenta //洋红

	info  = "[INFO]"
	debug = "[DEBUG]"
	erro  = "[ERROR]"
	warn  = "[WARN]"
)

var (
	file *os.File
)

func init() {
	f, err := os.OpenFile("text.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	file = f
}

// Debug see complete color rules in document in https://en.wikipedia.org/wiki/ANSI_escape_code#cite_note-ecma48-13
func Debug(format string, a ...interface{}) {
	logger := log.New(file, "prefix", log.LstdFlags)
	if LogLevelNum <= 1 {
		prefix := yellow(debug)
		logger.Println(formatLog(prefix), fmt.Sprintf(format, a...))
		fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
	}
}

// Info to infi log
func Info(format string, a ...interface{}) {
	logger := log.New(file, "prefix", log.LstdFlags)
	if LogLevelNum <= 2 {
		prefix := green(info)
		fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
		logger.Println(formatLog(prefix), fmt.Sprintf(format, a...))
	}
}

// func Success(format string, a ...interface{}) {
// 	prefix := blue(info)
// 	fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
// }

// Warning to warn
func Warning(format string, a ...interface{}) {
	logger := log.New(file, "prefix", log.LstdFlags)
	if LogLevelNum <= 3 {
		prefix := magenta(warn)
		logger.Println(formatLog(prefix), fmt.Sprintf(format, a...))
		fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
	}
}

// Error to error
func Error(format string, a ...interface{}) {
	logger := log.New(file, "prefix", log.LstdFlags)
	if LogLevelNum <= 4 {
		prefix := red(erro)
		logger.Println(formatLog(prefix), fmt.Sprintf(format, a...))
		fmt.Println(formatLog(prefix), fmt.Sprintf(format, a...))
	}
}

func red(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorRed, s)
}

func green(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorGreen, s)
}

func yellow(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorYellow, s)
}

func blue(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorBlue, s)
}

func magenta(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", colorMagenta, s)
}

func formatLog(prefix string) string {
	return time.Now().Format("2006/01/02 15:04:05") + " " + prefix
}
