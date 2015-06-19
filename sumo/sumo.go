package sumo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"fmt"
	log "github.com/Sirupsen/logrus"
)

type Client struct {
	connection *http.Client
	endpoint   string
}

func NewClient(endpoint string) *Client {
	return &Client{
		connection: &http.Client{
			Timeout: time.Duration(10) * time.Second,
		},
	}
}

type LogLine struct {
	Image         string
	Container     string
	Level         string
	RawMessage    []byte
	ParsedMessage map[string]interface{}
}

func (ll *LogLine) Serialize() ([]byte, error) {
	sumoLine := map[string]interface{}{
		"_image":     ll.Image,
		"_container": ll.Container,
		"_loglevel":  ll.Level,
	}

	for key, value := range ll.ParsedMessage {
		sumoLine[key] = value
	}

	return json.Marshal(sumoLine)
}

func (c *Client) Log(line *LogLine) error {
	payload, err := line.Serialize()
	if err != nil {
		log.WithField("error", err).Warn("could not serialize payload")
		return err
	}
	buffer := bytes.NewBuffer(payload)

	response, err := c.connection.Post(c.endpoint, "application/json", buffer)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"endpoint": c.endpoint,
		}).Warn("could not deliver logline to sumologic")
		return err
	}

	log.WithFields(log.Fields{
		"statuscode": fmt.Sprintf("%d", response.StatusCode),
	}).Debug("logged to sumologic")

	return nil
}
