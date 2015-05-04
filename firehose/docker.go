package firehose

import (
	log "github.com/christian-blades-cb/hoseclamp/_vendor/logrus"

	"github.com/fsouza/go-dockerclient"

	"path/filepath"
	"strings"
)

// TODO: Document public methods

func LogLineStream(host, certpath string, rawLines chan<- ContainerLine) error {
	client := getClient(host, certpath)
	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.WithField("err", err).Warn("Error listing containers")
		return err
	}

	for _, container := range containers {
		log.WithField("containerId", container.ID).Debug("Attaching to container")
		go lineCollector(client, container.ID, container.Image, rawLines)
	}

	return nil
}

func lineCollector(client *docker.Client, containerId string, imageName string, outputChan chan<- ContainerLine) {
	outputBuffer := &ChannelStream{
		OutputChannel: outputChan,
		ContainerId:   containerId,
		Image:         imageName,
	}

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

type ContainerLine struct {
	Image       string
	ContainerId string
	RawLine     []byte
	ParsedLine  map[string]interface{}
}

type ChannelStream struct {
	OutputChannel chan<- ContainerLine
	Image         string
	ContainerId   string
}

func (cs *ChannelStream) Write(p []byte) (n int, err error) {
	cs.OutputChannel <- ContainerLine{
		RawLine:     p,
		Image:       cs.Image,
		ContainerId: cs.ContainerId,
	}

	return len(p), nil
}
