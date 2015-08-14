package firehose

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/christian-blades-cb/gojsonexplode"

	"github.com/kr/logfmt"
)

// ContainerLine represents a log line, along with some metadata which identifies which container it came from.
type ContainerLine struct {
	Image       string
	ContainerId string
	RawLine     []byte
	ParsedLine  map[string]interface{}
}

// Parse attempts to deserialize a JSON or Logfmt logline. The parsed line is stored in the ParsedLine field.
func (cl *ContainerLine) Parse() {
	var err error
	if cl.ParsedLine, err = unmarshalJson(cl.RawLine); err != nil {
		cl.ParsedLine = unmarshalLogfmt(cl.RawLine)
	}
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
			log.Warn("recovered from panic while unmarshalling logfmt")
		}
	}()

	logline := make(logfmtMap)
	logfmt.Unmarshal(line, logline)
	return logline
}

type logfmtMap map[string]interface{}

func (lm logfmtMap) HandleLogfmt(key, val []byte) error {
	keystring := string(key[:])
	valstring := string(val[:])
	lm[keystring] = valstring

	return nil
}
