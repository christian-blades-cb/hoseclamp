package main // import "github.com/christian-blades-cb/hoseclamp"

import (
	"strings"

	"github.com/jessevdk/go-flags"

	log "github.com/Sirupsen/logrus"
	"github.com/christian-blades-cb/hoseclamp/firehose"
	"github.com/christian-blades-cb/hoseclamp/logio"
	"github.com/christian-blades-cb/hoseclamp/sumo"
)

func main() {
	var opts struct {
		LogioServer string `short:"s" long:"server" description:"logio server" default:"localhost:28777" env:"LOGIO_SERVER"`
		SumoServer  string `short:"u" long:"sumourl" description:"http collector endpoint for Sumologic" env:"SUMOLOGIC_ENDPOINT"`

		DockerHost     string `long:"host" description:"Docker Host" required:"true" default:"unix:///var/run/docker.sock" env:"DOCKER_HOST"`
		DockerCertPath string `long:"certpath" description:"Docker TLS Certificate path" env:"DOCKER_CERT_PATH"`
		Verbose        bool   `long:"verbose" short:"v" description:"all the logs"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal("Unable to parse arguments.")
	}

	if opts.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	rawLines := make(chan *firehose.ContainerLine, 128)

	sumoBatcher := sumo.NewSumoClient(opts.SumoServer)
	log.WithField("sumologic_endpoint", opts.SumoServer).Info("using sumologic")

	go func() {
		sumoBatcher.Start()
		sendToSumoBatcher(sumoBatcher.Work, rawLines)
		sumoBatcher.Stop()
	}()

	err = firehose.StartLoglineStream(opts.DockerHost, opts.DockerCertPath, rawLines)
	if err != nil {
		log.WithField("err", err).Warn("error on startup")
	}
}

func sendToLogio(client *logio.Client, loglines <-chan *firehose.ContainerLine) {
	for line := range loglines {
		line.Parse()

		level := "Info"
		if l, ok := line.ParsedLine["line.level"]; ok {
			if lvl, ok := l.(string); ok && lvl != "" {
				level = lvl
			}
		}

		logline := &logio.LogLine{
			Node:    strings.Replace(line.Image, ":", "__", -1),
			Stream:  line.ContainerId,
			Level:   level,
			Message: logfmtMap(line.ParsedLine),
		}

		log.Debugln(logline.Serialize())
		client.Log(logline)
	}
}

func getLevel(line *firehose.ContainerLine) string {
	if l, ok := line.ParsedLine["line.level"]; ok {
		if lvl, ok := l.(string); ok && lvl != "" {
			return lvl
		}
	}

	return "Info"
}

func sendToSumoLogic(client *sumo.Client, loglines <-chan *firehose.ContainerLine) {
	for line := range loglines {
		line.Parse()

		logline := &sumo.LogLine{
			Image:         line.Image,
			Container:     line.ContainerId,
			RawMessage:    nil,
			Level:         getLevel(line),
			ParsedMessage: line.ParsedLine,
		}

		log.Debugln(logline.Serialize())
		client.Log(logline)
	}
}

func sendToSumoBatcher(workChannel chan<- interface{}, loglines <-chan *firehose.ContainerLine) {
	for line := range loglines {
		line.Parse()

		logline := sumo.LogLine{
			Image:         line.Image,
			Container:     line.ContainerId,
			RawMessage:    nil,
			Level:         getLevel(line),
			ParsedMessage: line.ParsedLine,
		}

		workChannel <- logline
	}
}
