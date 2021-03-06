package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
)

type Level int

const (
	Quiet Level = iota
	Error
	Warn
	Info
	Verbose
	Debug
)

func (s Level) LessEqual(other Level) bool { return int(s) <= int(other) }

var (
	level        = Info
	filterRegexp *regexp.Regexp
)

// SetLevel sets logging level.
func SetLevel(lev Level) { level = lev }

// SetFilter sets regexp to select logs.
// If re is not nil, logs that matched with re are displayed.
func SetFilter(re *regexp.Regexp) { filterRegexp = re }

func toBeLogged(lev Level) bool { return lev.LessEqual(level) }

func outputf(v string) {
	if filterRegexp == nil || filterRegexp.MatchString(v) {
		log.Println(v)
	}
}

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

func JSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
