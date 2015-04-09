package loghose

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnmarshalling(t *testing.T) {
	Convey("When I unmarshal a v0 JSON into a struct", t, func() {
		v0_json := `{"v":0,"id":"cce7ac6bdc36","image":"lcs:latest","name":"lcs","line":{"time":"2015-04-09T04:40:33Z","level":"info","msg":"req_served","app":"luceo-config-store","latency":"0.4984 ms","method":"HEAD","remote":"192.168.59.3:57210","req_id":"cce7ac6bdc36/U3rctIAYm5-000010","status":"404","uri":"/"}}`

		var v0_struct LoghoseLine

		err := json.Unmarshal([]byte(v0_json), &v0_struct)
		So(err, ShouldBeNil)

		Convey("Version is populated", func() {
			So(v0_struct.Version, ShouldEqual, 0)
		})

		Convey("ContainerId is populated", func() {
			So(v0_struct.ContainerId, ShouldEqual, "cce7ac6bdc36")
		})

		Convey("Image is populated", func() {
			So(v0_struct.Image, ShouldEqual, "lcs:latest")
		})

		Convey("ContainerName is populated", func() {
			So(v0_struct.ContainerName, ShouldEqual, "lcs")
		})

		Convey("Logline is populated", func() {
			So(v0_struct.Logline, ShouldNotBeNil)
		})
	})
}
