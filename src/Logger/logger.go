package Logger

import (
	"Config"
	"log"
)

var blacklist = []string{"Start", "End"}

type Msg []string
type ErrMsg struct {
	Err    error
	Status string
}

var Verbose chan Msg
var VeryVerbose chan Msg
var Debug chan Msg
var Error chan ErrMsg
var Assert chan Msg
var Event chan Msg
var Warning chan Msg

func Start() {
	Verbose = make(chan Msg)
	StartLog(&Verbose, "verbose")
	VeryVerbose = make(chan Msg)
	StartLog(&VeryVerbose, "very_verbose")
	Debug = make(chan Msg)
	StartLog(&Debug, "debug")
	Assert = make(chan Msg)
	StartLog(&Assert, "assert")
	Event = make(chan Msg)
	StartLog(&Event, "event")
	Warning = make(chan Msg)
	StartLog(&Warning, "warning")

	Error = make(chan ErrMsg)
	logMsg := func(s1 error, s2 string) {
		if !Include(s2, blacklist) {
			log.Fatalf("<error> :%s:%s", s1, s2)
		}
	}
	go func() {
		logMsg(nil, "Start")
		for message := range Error {
			if message.Err != nil {
				logMsg(message.Err, message.Status)
			}
		}
	}()
}
func Close() {
	StopLog(&Verbose, "verbose")
	StopLog(&VeryVerbose, "very_verbose")
	StopLog(&Debug, "debug")
	StopLog(&Event, "event")
	StopLog(&Warning, "warning")
	StopLog(&Assert, "assert")
	Error <- ErrMsg{nil, "End"}
	close(Error)
}
func Include(compare string, against []string) bool {
	for _, s := range against {
		if s == compare {
			return true
		}
	}
	return false
}

//Startlog will start the log
func StartLog(logChannel *chan Msg, logState string) {
	logMsg := func(s1 string, s2 string) {
		if !Include(s1, blacklist) {
			log.Printf(" <%s> %s:%s", logState, s1, s2)
		}
	}
	if Config.Log.HasState(logState) {
		go func() {
			logMsg("Start", "Logging")
			for message := range *logChannel {
				if len(message) > 1 {
					logMsg(message[1], message[0])
				} else {
					logMsg("Log", message[0])
				}

			}
		}()
	}
}
func StopLog(logChannel *chan Msg, logState string) {
	if Config.Log.HasState(logState) {
		*logChannel <- Msg{"Logging", "End"}
		close(*logChannel)
	} else {
		close(*logChannel)
	}
}
