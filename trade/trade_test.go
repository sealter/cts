package trade

import (
	"testing"

	"github.com/modood/cts/gateio"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPosition(t *testing.T) {
	Convey("should get position unsuccessfully", t, func(c C) {
		gateio.Init("key", "secret")
		_, _, err := Position()
		So(err, ShouldNotBeNil)
	})
}

func TestAllIn(t *testing.T) {
	Convey("should all in unsuccessfully", t, func(c C) {
		gateio.Init("key", "secret")
		err := AllIn("btc_usdt")
		So(err, ShouldNotBeNil)
	})
}
