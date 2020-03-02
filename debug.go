package sshtest

import "log"

var debugEnabled bool

func DebugOn() {
	debugEnabled = true
}

func DebugOff() {
	debugEnabled = false
}

func debug(v ...interface{}) {
	if debugEnabled {
		log.Print(v...)
	}
}

func debugf(format string, v ...interface{}) {
	if debugEnabled {
		log.Printf(format, v...)
	}
}
