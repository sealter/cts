package huobi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInit(t *testing.T) {
	Convey("should init successfully", t, func() {
		So(key, ShouldEqual, "")
		So(secret, ShouldEqual, "")
		Init("apikey", "secretkey")
		So(key, ShouldEqual, "apikey")
		So(secret, ShouldEqual, "secretkey")
	})
}

func TestMarginAccount(t *testing.T) {
	Convey("should return margin account unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		_, err := MarginAccount("btcusdt")
		So(err, ShouldNotBeNil)
	})
}

func TestReq(t *testing.T) {
	Convey("should transfer unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		bs, err := req("POST", "https://api.huobi.pro/v1/dw/transfer-out/margin", map[string]string{
			"symbol":   "btcusdt",
			"currency": "usdt",
			"amount":   "1",
		})
		So(err, ShouldBeNil)
		err = handle(bs, err)
		So(err, ShouldNotBeNil)
	})
}
