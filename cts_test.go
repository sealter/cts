package main

import (
	"testing"

	"github.com/modood/cts/gateio"
	"github.com/modood/cts/strategy"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignal(t *testing.T) {
	Convey("should refresh balance cache unsuccessfully", t, func(c C) {
		sl := strategy.Available()

		for _, v := range sl {
			sig, err := signal(v)
			So(err, ShouldBeNil)
			So(strategy.Signals(), ShouldContain, sig)
		}

		sig, err := signal("xxxxxx")
		So(err, ShouldNotBeNil)
		So(sig, ShouldEqual, strategy.SIG_NONE)

	})
}

func TestExec(t *testing.T) {
	Convey("should refresh balance cache unsuccessfully", t, func(c C) {
		gateio.Init("apikey", "secretkey")
		var err error
		err = exec(strategy.SIG_RISE, "doge_usdt")
		So(err, ShouldNotBeNil)

		err = exec(strategy.SIG_FALL, "doge_usdt")
		So(err, ShouldNotBeNil)

		err = exec(strategy.SIG_NONE, "doge_usdt")
		So(err, ShouldBeNil)
	})
}
