package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/jessevdk/go-flags"

	log "github.com/christian-blades-cb/log-multiplexer/_vendor/logrus"
	"github.com/christian-blades-cb/log-multiplexer/loghose"
	"github.com/christian-blades-cb/log-multiplexer/logio"
	"strings"
)

func main() {
	var opts struct {
		LogioServer string `short:"s" long:"server" description:"logio server" default:"localhost:28777" env:"LOGIO_SERVER"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal("Unable to parse arguments.")
	}

	logioServer := opts.LogioServer
	log.WithField("server", logioServer).Info("Connecting to logio server")

	client, err := logio.NewClient(logioServer)
	if err != nil {
		log.WithField("err", err.Error()).Fatal("error connecting to logio")
	}
	defer client.Close()

	loglines := make(chan *loghose.LoghoseLine)

	go sendToLogio(client, loglines)
	readParseLoop(loglines)
}

func sendToLogio(client *logio.Client, loglines <-chan *loghose.LoghoseLine) {
	for line := range loglines {

		level := "Info"
		if l, ok := line.Logline["line.level"]; ok {
			if lvl, ok := l.(string); ok && lvl != "" {
				level = l.(string)
			}
		}

		logline := &logio.LogLine{
			Node:    strings.Replace(line.Image, ":", "__", -1),
			Stream:  fmt.Sprintf("%s-%s", line.ContainerName, line.ContainerId),
			Level:   level,
			Message: line.LogfmtLine(),
		}

		log.Infoln(logline.Serialize())
		client.Log(logline)
	}
}

func readParseLoop(loglines chan<- *loghose.LoghoseLine) {
	consoleReader := bufio.NewReader(os.Stdin)
	for {
		rawline, ioErr := consoleReader.ReadSlice('\n')

		loghoseLine, err := loghose.Parse(rawline)
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err.Error(),
				"line": string(rawline[:]),
			}).Error("unable to parse line")
			continue
		}

		loglines <- loghoseLine

		if ioErr == io.EOF {
			break
		} else if ioErr != nil {
			log.WithField("err", ioErr).Fatal("error reading from stream")
		}
	}
}
