package main

import (
	"fmt"
	"time"

	//"fmt"
	"os"
	//"strings"
	//"time"
	logging "github.com/ipfs/go-log"
	//logwriter "github.com/ipfs/go-log/writer"

	logwriter "github.com/ipfs/go-log/writer"
)

var (
	logfilename string
	logfile     *os.File

	log = logging.Logger("cmdtool")
)

// startDebugging starts logging of kad-dht
func startDebugging(logfilename string) (*os.File, error) {

	if len(logfilename) == 0 {

		// Prepare logfile for logging
		year, month, day := time.Now().Date()
		hour, minute, second := time.Now().Clock()
		logfilename = fmt.Sprintf("debug-%s-%v%02d%02d%02d%02d%02d.log", name,
			year, int(month), int(day), int(hour), int(minute), int(second))
	}

	file, err := os.OpenFile(logfilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file %q for debugging failed: %v\n", logfilename, err)
	}

	logwriter.Configure(logwriter.Output(file))

	return file, nil
}

//// startLogging start logging for cmdtool itself
//func startLogging(name string) (*os.File, error) {
//
//	// Config logging
//	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
//	log.SetPrefix("DEBUG: ")
//
//	// Prepare logfile for logging
//	year, month, day := time.Now().Date()
//	hour, minute, second := time.Now().Clock()
//	logfilename = fmt.Sprintf("cmdtool-%s-%v%02d%02d%02d%02d%02d.log", name,
//		year, int(month), int(day), int(hour), int(minute), int(second))
//
//	logfile, err := os.OpenFile(logfilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
//	if err != nil {
//		return nil, fmt.Errorf("error opening logfile %v: %v", logfilename, err)
//	}
//
//	// Switch logging to logfile
//	log.SetOutput(logfile)
//
//	// First entry in the individual log file
//	log.Printf("Session starting\n")
//
//	return logfile, nil
//}
