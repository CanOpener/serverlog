package serverlog

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
	logToConsol = false // should log to consol
	logToFile   = false // should log to file
	logDir      = ""    // directory in which to store logfiles
	maxDays     = -1    // total number of logfiles at any time
	// if exeeded will delete oldest

	//log queue
	logChan     = make(chan logItem, 100)
	logNameChan = make(chan string, 2)
	killChan    = make(chan bool, 1)

	//colour functions
	startupColourFunc = ansi.ColorFunc("green+b:black")
	fatalColourFunc   = ansi.ColorFunc("red+b:black")
	generalColourFunc = ansi.ColorFunc("blue+b:black")
	warningColourFunc = ansi.ColorFunc("yellow+b:black")
)

// The clolour type enumerator
const (
	startupColour int = iota
	fatalColour
	generalColour
	warningColour
)

// Init initialises the srvlog package. if either consoleLog or fileLog
// is true it will start the logger in another gorutine ready to log
func Init(consolLog, fileLog bool, maxLogDays int, pathToLogDir string) {
	logToConsol = consolLog
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
		content:          args,
	}
	logChan <- logItem
}

// Fatal is used to log something a server killing circumstance
// same as log.Fatalln()
// This will terminate the process with an exit code of 1
func Fatal(args ...interface{}) {
	logItem := logItem{
		prefix:           "FATAL:  ",
		prefixColourFunc: fatalColour,
		content:          args,
	}
	logChan <- logItem
}

// General is used to log general stuff
func General(args ...interface{}) {
	logItem := logItem{
		prefix:           "GENERAL:",
		prefixColourFunc: generalColour,
		content:          args,
	}
	logChan <- logItem
}

// Warning is used to log warnings
func Warning(args ...interface{}) {
	logItem := logItem{
		prefix:           "WARNING:",
		prefixColourFunc: warningColour,
		content:          args,
	}
	logChan <- logItem
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
	content          []interface{}
}

// listen is the listener which runs in its own gorutine and logs messages
func listen() {
	currentLogPath := path.Join(logDir, time.Now().Format("02-01-2006-{csrv}.log"))

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

		line := time.Now().Format("15:04:05") + " " + item.prefix + " " + fmt.Sprintln(item.content)
		_, err = file.WriteString(line)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// writeToConsole logs a message to the console
func writeToConsole(item logItem) {
	if logToConsol {
		fmt.Print(time.Now().Format("15:04:05"), colourInText(item.prefix, item.prefixColourFunc), " ")
		fmt.Println(item.content...)
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
			// tell listener new logfile name.
			logNameChan <- path.Join(logDir, time.Now().Format("02-01-2006-{csrv}.log"))

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
