package firehose

import (
	"encoding/json"

	"github.com/christian-blades-cb/gojsonexplode"
	log "github.com/christian-blades-cb/hoseclamp/_vendor/logrus"

	"github.com/kr/logfmt"
)

func Parse(line ContainerLine) {
	line.ParsedLine = parseLine(line.RawLine)
}

func parseLine(line []byte) map[string]interface{} {
	logline, err := unmarshalJson(line)
	if err != nil {
		logline = unmarshalLogfmt(line)
	}

	return logline
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
