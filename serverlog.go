package serverlog

import (
	"fmt"
	"github.com/mgutz/ansi"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	logToConsol = false
	logToFile   = false
	filePath    = "/home/mladen/Desktop/test.log"

	//log queue
	logChan = make(chan logItem, 1024)

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
func Init(consolLog, fileLog bool, pathToFile string) {
	logToConsol = consolLog
	logToFile = fileLog
	filePath = pathToFile

	if logToConsol || logToFile {
		go listen()
	}
}

// Startup used to log the startup message
// example "Startup("Server listening on port:", PORT)"
func Startup(args ...interface{}) {
	logChan <- logItem{
		longTime:         true,
		preset:           "STARTUP:",
		presetColourFunc: startupColour,
		content:          stringFromArgs(args...),
	}
}

// Fatal is used to log something fatal
// This will terminate the process with an exit code of 1
func Fatal(args ...interface{}) {
	logChan <- logItem{
		longTime:         false,
		preset:           "FATAL:",
		presetColourFunc: fatalColour,
		content:          stringFromArgs(args...),
	}
}

// General is used to log general stuff
func General(args ...interface{}) {
	logChan <- logItem{
		longTime:         false,
		preset:           "GENERAL:",
		presetColourFunc: generalColour,
		content:          stringFromArgs(args...),
	}
}

// Warning is used to log warnings
func Warning(args ...interface{}) {
	logChan <- logItem{
		longTime:         false,
		preset:           "WARNING:",
		presetColourFunc: warningColour,
		content:          stringFromArgs(args...),
	}
}

// stringFromArgs returns a string of the combination of the arguments given
func stringFromArgs(args ...interface{}) string {
	var formatStr string
	for i := range args {
		if i != 0 {
			formatStr += " "
		}
		formatStr += "%v"
	}
	return fmt.Sprintf(formatStr, args...)
}

// logitem is the struct passed to the logger function
type logItem struct {
	longTime         bool
	preset           string
	presetColourFunc int
	content          string
}

// listen is the listener which runs in its own gorutine and logs messages
func listen() {
	for {
		select {
		case item := <-logChan:
			writeToConsole(item)
			writeToFile(item)
			if item.presetColourFunc == fatalColour {
				os.Exit(1)
			}
		}
	}
}

// writeToFile logs a message to the logfile
func writeToFile(item logItem) {
	if logToFile {
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalln(err)
		}

		line := getTimestamp(item.longTime) + " " + item.preset + " " + item.content + "\n"
		_, err = file.WriteString(line)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// writeToConsole logs a message to the console
func writeToConsole(item logItem) {
	if logToConsol {
		fmt.Println(getTimestamp(item.longTime), colourInText(item.preset, item.presetColourFunc), item.content)
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

// returns the string form of the current time
// specify true for presetting with the date too
func getTimestamp(long bool) string {
	now := time.Now().Local()
	hour, min, sec := now.Clock()
	clockString := strconv.Itoa(hour) + ":" + strconv.Itoa(min) + ":" + strconv.Itoa(sec)
	if long {
		year, month, day := now.Date()
		dateString := strconv.Itoa(year) + "/" + month.String() + "/" + strconv.Itoa(day)
		return dateString + " " + clockString
	}
	return clockString
}
