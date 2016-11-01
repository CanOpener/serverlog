package serverlog

/*
The MIT License (MIT)
Copyright (c) 2015 Mladen Kajic
See https://github.com/CanOpener/serverlog
*/

import (
	"fmt"
	"github.com/mgutz/ansi"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"syscall"
	"time"
)

var (
	logToConsole = false // should log to console
	logToFile    = false // should log to file
	logDir       = ""    // directory in which to store logfiles
	maxDays      = -1    // total number of logfiles at any time
	// if exeeded will delete oldest

	//log queue
	logChan     = make(chan logItem, 100)
	logNameChan = make(chan string, 2)
	killChan    = make(chan bool, 1)

	//colour functions
	startupColourFunc = ansi.ColorFunc("green+b")
	fatalColourFunc   = ansi.ColorFunc("red+b")
	generalColourFunc = ansi.ColorFunc("blue+b")
	warningColourFunc = ansi.ColorFunc("yellow+b")
)

// The colour type enumerator
const (
	startupColour int = iota
	fatalColour
	generalColour
	warningColour
)

// Init initialises the srvlog package. if either consoleLog or fileLog
// is true it will start the logger in another gorutine ready to log
func Init(consoleLog bool, fileLog bool, maxLogDays int, pathToLogDir string) {
	logToConsole = consoleLog
	logToFile = fileLog
	logDir = pathToLogDir
	maxDays = maxLogDays

	if logToFile {
		// make sure log directory exists
		info, err := os.Stat(logDir)
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatalln("The directory specified to serverlog does not exist.")
			}
			log.Fatalln(err)
		}
		if !info.IsDir() {
			log.Fatalln("The path specified to serverlog is not a directory.")
		}

		// make sure have premissions to log directory
		err = syscall.Access(logDir, syscall.O_RDWR)
		if err != nil {
			log.Fatalln("Serverlog needs read and write premissions to the specified directory.")
		}

		// manage logfile names and number of logfiles at any one time
		// only needed if logging to files
		go logFileOverseer()
	}

	go listen()
}

// Startup used to log the startup message
// example "Startup("Server listening on port:", PORT)"
func Startup(args ...interface{}) {
	logItem := logItem{
		prefix:           "STARTUP:",
		prefixColourFunc: startupColour,
		time:             time.Now(),
		content:          args,
	}
	logChan <- logItem
}

// Startupf used to log the startup message
// Formats the text before sending it
func Startupf(format string, args ...interface{}) {
	Startup(fmt.Sprintf(format, args...))
}

// Fatal is used to log something a server killing circumstance
// same as log.Fatalln()
// This will terminate the process with an exit code of 1
func Fatal(args ...interface{}) {
	logItem := logItem{
		prefix:           "FATAL:  ",
		prefixColourFunc: fatalColour,
		time:             time.Now(),
		content:          args,
	}
	logChan <- logItem
}

// Fatalf is used to log something a server killing circumstance
// Formats the text before sending it
func Fatalf(format string, args ...interface{}) {
	Fatal(fmt.Sprintf(format, args...))
}

// General is used to log general stuff
func General(args ...interface{}) {
	logItem := logItem{
		prefix:           "GENERAL:",
		prefixColourFunc: generalColour,
		time:             time.Now(),
		content:          args,
	}
	logChan <- logItem
}

// Generalf is used to log general stuff
// Formats the text before sending it
func Generalf(format string, args ...interface{}) {
	General(fmt.Sprintf(format, args...))
}

// Warning is used to log warnings
func Warning(args ...interface{}) {
	logItem := logItem{
		prefix:           "WARNING:",
		prefixColourFunc: warningColour,
		time:             time.Now(),
		content:          args,
	}
	logChan <- logItem
}

// Warningf is used to log warnings
// Formats the text before sending it
func Warningf(format string, args ...interface{}) {
	Warning(fmt.Sprintf(format, args...))
}

// Kill will terminate the listener and logFileOverseer
func Kill() {
	killChan <- false
	logToConsol = false
	logToFile = false
}

// logitem is the struct passed to the logger function
type logItem struct {
	prefix           string
	prefixColourFunc int
	time             time.Time
	content          []interface{}
}

// listen is the listener which runs in its own gorutine and logs messages
func listen() {
	currentLogPath := path.Join(logDir, time.Now().Format("2006-01-02.crsv.log"))

	for {
		select {
		case item := <-logChan:
			writeToConsole(item)
			writeToFile(item, currentLogPath)
			if item.prefixColourFunc == fatalColour {
				os.Exit(1)
			}
		case newLogPath := <-logNameChan:
			currentLogPath = newLogPath
		case <-killChan:
			return
		}
	}
}

// writeToFile logs a message to the logfile
func writeToFile(item logItem, logPath string) {
	if logToFile {
		file, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		defer file.Close()
		if err != nil {
			log.Fatalln(err)
		}

		line := item.time.Format("15:04:05") + " " + item.prefix + " " + fmt.Sprintln(item.content...)
		_, err = file.WriteString(line)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// writeToConsole logs a message to the console
func writeToConsole(item logItem) {
	if logToConsol {
		fmt.Print(item.time.Format("15:04:05"), " ", colourInText(item.prefix, item.prefixColourFunc), " ", fmt.Sprintln(item.content...))
	}
}

// logFileOverseer makes sure that there are never more logfiles for each day than
// maxLogDays. also makes sure to update the listener as to new logfile names
// for each day.
func logFileOverseer() {
	for {
		in24Hr := time.Now().AddDate(0, 0, 1) // AddDate is used in case the next day is next month or year
		tomorrow := time.Date(in24Hr.Year(), in24Hr.Month(), in24Hr.Day(), 0, 0, 0, 0, in24Hr.Location())
		timeToWait := tomorrow.Sub(time.Now())
		newDay := time.After(timeToWait)

		select {
		case <-newDay:
			newLogFile := path.Join(logDir, tomorrow.Format("2006-01-02.crsv.log"))

			// create new logfile
			_, err := os.Create(newLogFile)
			if err != nil {
				Warning("Serverlog failed to create new logfile:", newLogFile, ":", err)
				break
			}

			// tell listener new logfile name.
			logNameChan <- newLogFile

			// check number of logfiles
			if maxDays > 0 {
				files, err := ioutil.ReadDir(logDir)
				if err != nil {
					Warning("Server log failed to read from log directory :", err)
					break
				}

				logs := make([]string, 0, maxDays*2)
				index := 0
				for _, file := range files {
					if strings.Contains(file.Name(), "{csrv}") {
						logs[index] = file.Name()
						index++
					}
				}
				sort.Strings(logs)

				numberLogsLeft := len(logs)
				for i := 0; (numberLogsLeft > maxDays) && (i < len(logs)); i++ {
					logToDelete := path.Join(logDir, logs[i])
					err := os.Remove(logToDelete)
					if err != nil {
						Warning("Server log failed to delete logfile :", logToDelete, ":", err)
						break
					}
					numberLogsLeft--
				}
			}
		case <-killChan:
			return
		}
	}
}

// colourInText returns the text coloured in based on the colour func enumerator
func colourInText(text string, colourFunc int) string {
	switch colourFunc {
	case startupColour:
		return startupColourFunc(text)
	case fatalColour:
		return fatalColourFunc(text)
	case generalColour:
		return generalColourFunc(text)
	case warningColour:
		return warningColourFunc(text)
	default:
		return text
	}
}
