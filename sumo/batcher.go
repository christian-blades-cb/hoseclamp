package sumo

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/facebookgo/httpcontrol"
	"github.com/facebookgo/muster"
	"net/http"
	"time"
)

const (
	DEFAULTSUMOBATCHSIZE         = 256
	DEFAULTSUMOBATCHTIMEOUT      = 1 * time.Minute
	DEFAULTSUMOCONCURRENTBATCHES = 5
	DEFAULTSUMOPENDINGWORKCAP    = 50
)

type sumoBatch struct {
	payloadLines [][]byte
	index        uint
	httpClient   *http.Client
	endpoint     string
	maxLines     uint
}

func (s *sumoBatch) Add(item interface{}) {
	if s.index >= s.maxLines {
		log.Warn("Attempted to append to a full batch buffer. Data lost.")
		return
	}

	logline := item.(LogLine)

	if serialized, err := logline.Serialize(); err != nil {
		log.WithField("error", err).Warn("could not serialize logline")
	} else {
		s.payloadLines[s.index] = serialized
		s.index = s.index + 1
	}
}

func (s *sumoBatch) Fire(notifier muster.Notifier) {
	if s.index == 0 {
		log.Debug("empty batch, not delivering")
		notifier.Done()
		return
	}

	lastIndex := s.index + 1
	if lastIndex > uint(len(s.payloadLines)) {
		lastIndex = uint(len(s.payloadLines))
	}

	payload := bytes.Join(s.payloadLines[:lastIndex], []byte("\n"))
	buffer := bytes.NewBuffer(payload)

	if response, err := s.httpClient.Post(s.endpoint, "application/json; charset=utf-8", buffer); err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"endpoint": s.endpoint,
		}).Warn("could not deliver loglines to sumologic")
	} else {
		log.WithFields(log.Fields{
			"statuscode": fmt.Sprintf("%d", response.StatusCode),
			"payload":    string(payload),
		}).Debug("delivered to sumologic")
	}

	notifier.Done()
}

func newBatcher(batchSize uint, endpoint string, httpClient *http.Client) *sumoBatch {
	return &sumoBatch{
		payloadLines: make([][]byte, batchSize),
		index:        0,
		maxLines:     batchSize,
		endpoint:     endpoint,
		httpClient:   httpClient,
	}
}

func NewSumoClient(endpoint string) muster.Client {
	httpClient := &http.Client{
		Transport: &httpcontrol.Transport{
			RequestTimeout: 10 * time.Second,
		},
	}

	return muster.Client{
		MaxBatchSize:         DEFAULTSUMOBATCHSIZE,
		BatchTimeout:         DEFAULTSUMOBATCHTIMEOUT,
		MaxConcurrentBatches: DEFAULTSUMOCONCURRENTBATCHES,
		PendingWorkCapacity:  DEFAULTSUMOPENDINGWORKCAP,
		BatchMaker: func() muster.Batch {
			return newBatcher(DEFAULTSUMOBATCHSIZE, endpoint, httpClient)
		},
	}
}
