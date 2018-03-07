package gateio

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTicker(t *testing.T) {
	Convey("should return pair ticker", t, func() {
		r, err := Ticker("btc_usdt")
		So(err, ShouldBeNil)
		So(r.Result, ShouldEqual, true)

		_, err = Ticker("btc_shit")
		So(err, ShouldNotBeNil)
	})
}

func TestTickers(t *testing.T) {
	Convey("should return pairs", t, func() {
		_, err := Tickers()
		So(err, ShouldBeNil)
	})
}

func TestRate(t *testing.T) {
	Convey("should return rate successfully", t, func() {
		r, err := Rate()
		So(err, ShouldBeNil)
		So(r, ShouldBeGreaterThan, 0)
	})
}

func TestTrend(t *testing.T) {
	Convey("should return trend successfully", t, func() {
		_, _, err := Trend()
		So(err, ShouldBeNil)
	})
}
