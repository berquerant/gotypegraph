package logger

import (
	"fmt"
	"log"
)

type Level int

const (
	Error Level = iota
	Warn
	Info
	Debug
	Verbose
)

func (s Level) LessEqual(other Level) bool { return int(s) <= int(other) }

var (
	level = Info
)

func SetLevel(lev Level) { level = lev }

func toBeLogged(lev Level) bool { return lev.LessEqual(level) }

func outputf(v string) { log.Println(v) }

func Verbosef(format string, v ...interface{}) {
	if toBeLogged(Verbose) {
		outputf(fmt.Sprintf("[V] %s", fmt.Sprintf(format, v...)))
	}
}

func Debugf(format string, v ...interface{}) {
	if toBeLogged(Debug) {
		outputf(fmt.Sprintf("[D] %s", fmt.Sprintf(format, v...)))
	}
}

func Infof(format string, v ...interface{}) {
	if toBeLogged(Info) {
		outputf(fmt.Sprintf("[I] %s", fmt.Sprintf(format, v...)))
	}
}

func Warnf(format string, v ...interface{}) {
	if toBeLogged(Warn) {
		outputf(fmt.Sprintf("[W] %s", fmt.Sprintf(format, v...)))
	}
}

func Errorf(format string, v ...interface{}) {
	if toBeLogged(Error) {
		outputf(fmt.Sprintf("[E] %s", fmt.Sprintf(format, v...)))
	}
}
