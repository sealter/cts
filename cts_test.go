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
		So(sig, ShouldEqual, strategy.SigNone)

	})
}

func TestExec(t *testing.T) {
	Convey("should refresh balance cache unsuccessfully", t, func(c C) {
		gateio.Init("apikey", "secretkey")
		var err error
		err = exec(strategy.SigRise, "doge_usdt")
		So(err, ShouldNotBeNil)

		err = exec(strategy.SigFall, "doge_usdt")
		So(err, ShouldNotBeNil)

		err = exec(strategy.SigNone, "doge_usdt")
		So(err, ShouldBeNil)
	})
}
