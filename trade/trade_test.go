package trade

import (
	"testing"

	"github.com/modood/cts/gateio"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFlush(t *testing.T) {
	Convey("should refresh balance cache unsuccessfully", t, func(c C) {
		gateio.Init("key", "secret")
		err := Flush()
		So(err, ShouldNotBeNil)
	})
}

func TestPosition(t *testing.T) {
	Convey("should get position unsuccessfully", t, func(c C) {
		gateio.Init("key", "secret")
		_, err := Position()
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

func TestAllOut(t *testing.T) {
	Convey("should all out unsuccessfully", t, func(c C) {
		gateio.Init("key", "secret")
		err := AllOut("btc_usdt", "BTC")
		So(err, ShouldNotBeNil)
	})
}
