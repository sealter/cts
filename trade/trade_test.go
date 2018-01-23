package trade

import (
	"testing"

	"github.com/modood/cts/gateio"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPosition(t *testing.T) {
	Convey("should get position unsuccessfully", t, func(c C) {
		gateio.Init("key", "secret")
		_, err := Position()
		So(err, ShouldNotBeNil)
	})
}
