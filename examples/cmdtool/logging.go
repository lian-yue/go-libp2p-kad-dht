package main

import (
	"fmt"
	"os"
	"time"

	logging "github.com/ipfs/go-log"
	logwriter "github.com/ipfs/go-log/writer"
)

var (
	log = logging.Logger("cmdtool")
)

// startLogging starts logging of ipfs/go-log
func startLogging(logfilename string) (*os.File, error) {

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
