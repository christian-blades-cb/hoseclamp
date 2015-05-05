package main

import (
	"strings"

	"github.com/jessevdk/go-flags"

	log "github.com/Sirupsen/logrus"
	"github.com/christian-blades-cb/hoseclamp/firehose"
	"github.com/christian-blades-cb/hoseclamp/logio"
)

func main() {
	var opts struct {
		LogioServer string `short:"s" long:"server" description:"logio server" default:"localhost:28777" env:"LOGIO_SERVER"`

		DockerHost     string `long:"host" description:"Docker Host" required:"true" default:"unix:///var/run/docker.sock" env:"DOCKER_HOST"`
		DockerCertPath string `long:"certpath" description:"Docker TLS Certificate path" env:"DOCKER_CERT_PATH"`
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

	rawLines := make(chan *firehose.ContainerLine, 20)
	go sendToLogio(client, rawLines)

	err = firehose.LogLineStream(opts.DockerHost, opts.DockerCertPath, rawLines)
	if err != nil {
		log.WithField("err", err).Warn("error on startup")
	}
}

func sendToLogio(client *logio.Client, loglines <-chan *firehose.ContainerLine) {
	for line := range loglines {
		firehose.Parse(line)

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

		log.Infoln(logline.Serialize())
		client.Log(logline)
	}
}
