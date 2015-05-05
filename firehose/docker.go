package firehose

import (
	log "github.com/Sirupsen/logrus"

	"github.com/fsouza/go-dockerclient"

	"path/filepath"
	"strings"
)

// TODO: Document public methods

func LogLineStream(host, certpath string, rawLines chan<- *ContainerLine) error {
	client := getClient(host, certpath)

	if err := attachToRunningContainers(client, rawLines); err != nil {
		return err
	}

	dockerEvents := make(chan *docker.APIEvents, 5)
	attachToNewContainers(client, dockerEvents, rawLines)

	return nil
}

func attachToNewContainers(client *docker.Client, eventStream <-chan *docker.APIEvents, rawLines chan<- *ContainerLine) {
	for event := range eventStream {
		if event.Status == "start" || event.Status == "restart" {
			container, err := client.InspectContainer(event.ID)
			if err != nil {
				log.WithFields(log.Fields{
					"err":         err,
					"containerId": event.ID,
				}).Warn("could not retrieve information about starting container")
			} else {
				log.WithField("containerId", container.ID).Debug("container started, attaching")
				go lineCollector(client, container.ID, container.Image, rawLines)
			}
		}
	}
}

func attachToRunningContainers(client *docker.Client, rawLines chan<- *ContainerLine) error {
	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.WithField("err", err).Warn("error listing containers")
		return err
	}

	for _, container := range containers {
		log.WithField("containerId", container.ID).Debug("attaching to container")
		go lineCollector(client, container.ID, container.Image, rawLines)
	}

	return nil
}

func lineCollector(client *docker.Client, containerId string, imageName string, outputChan chan<- *ContainerLine) {
	outputBuffer := &channelStream{
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

type channelStream struct {
	OutputChannel chan<- *ContainerLine
	Image         string
	ContainerId   string
}

func (cs *channelStream) Write(p []byte) (n int, err error) {
	cs.OutputChannel <- &ContainerLine{
		RawLine:     p,
		Image:       cs.Image,
		ContainerId: cs.ContainerId,
	}

	return len(p), nil
}

type ContainerLine struct {
	Image       string
	ContainerId string
	RawLine     []byte
	ParsedLine  map[string]interface{}
}
