package main

import (
	"github.com/christian-blades-cb/gojsonexplode"
	log "github.com/christian-blades-cb/hoseclamp/_vendor/logrus"

	"fmt"
	"path/filepath"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/jessevdk/go-flags"

	"encoding/json"
	"github.com/kr/logfmt"
)

func main() {
	var opts struct {
		Host     string `long:"host" description:"Docker Host" required:"true" default:"unix:///var/run/docker.sock" env:"DOCKER_HOST"`
		CertPath string `long:"certpath" description:"Docker TLS Certificate path" env:"DOCKER_CERT_PATH"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		log.WithField("err", err).Fatal("Unable to parse command line args")
	}

	client := getClient(opts.Host, opts.CertPath)
	containers, _ := client.ListContainers(docker.ListContainersOptions{})
	for _, container := range containers {
		fmt.Printf("Name: %s\n", container.ID)
		fmt.Printf("Image: %s\n", container.Image)
	}

	outputChan := make(chan string)
	log.WithField("containerID", containers[0].ID).Info("Attaching")
	go lineCollector(client, containers[0].ID, outputChan)
	for line := range outputChan {
		outputLogline(line, containers[0].ID)
	}
}

func outputLogline(line string, containerID string) {
	var logline map[string]interface{}
	logline, err := unmarshalJson([]byte(line))
	if err != nil {
		logline = unmarshalLogfmt([]byte(line))
	}

	mm, _ := json.Marshal(logline)
	fmt.Printf("%s\n", mm)
}

func unmarshalJson(line []byte) (map[string]interface{}, error) {
	nestedMap := make(map[string]interface{})
	err := json.Unmarshal(line, &nestedMap)
	if err != nil {
		return nil, err
	}

	flattenedMap, err := gojsonexplode.ExplodeMap(nestedMap, "line", ".")
	if err != nil {
		return nil, err
	}

	return flattenedMap, nil
}

func unmarshalLogfmt(line []byte) map[string]interface{} {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("recovered")
		}
	}()

	logline := make(LogfmtMap)
	logfmt.Unmarshal(line, logline)
	return logline
}

type LogfmtMap map[string]interface{}

func (lm LogfmtMap) HandleLogfmt(key, val []byte) error {
	keystring := string(key[:])
	valstring := string(val[:])
	lm[keystring] = valstring

	return nil
}

func lineCollector(client *docker.Client, containerId string, outputChan chan<- string) {
	outputBuffer := &ChannelStream{OutputChannel: outputChan}
	client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    containerId,
		OutputStream: outputBuffer,
		ErrorStream:  outputBuffer,
		Logs:         true,
		Stream:       true,
		Stdout:       true,
		Stderr:       true,
	})
}

func getClient(host string, certpath string) *docker.Client {
	if strings.HasPrefix(host, "tcp") && certpath != "" {
		ca := filepath.Join(certpath, "ca.pem")
		cert := filepath.Join(certpath, "cert.pem")
		key := filepath.Join(certpath, "key.pem")
		client, err := docker.NewTLSClient(host, cert, key, ca)
		if err != nil {
			log.WithFields(log.Fields{
				"err":       err,
				"host":      host,
				"ca path":   ca,
				"cert path": cert,
				"key path":  key,
			}).Fatal("Could not connect to docker daemon")
		}
		return client
	} else {
		client, err := docker.NewClient(host)
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err,
				"host": host,
			}).Fatal("Could not connect to docker daemon")
		}
		return client
	}
}

type ChannelStream struct {
	OutputChannel chan<- string
}

func (cs *ChannelStream) Write(p []byte) (n int, err error) {
	cs.OutputChannel <- strings.TrimSpace(string(p[:]))
	return len(p), nil
}
