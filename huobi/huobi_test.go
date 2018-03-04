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
		_, err := MarginAccount("btc_usdt")
		So(err, ShouldNotBeNil)
	})
}

func TestOrders(t *testing.T) {
	Convey("should return finished orders unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		_, err := Orders("btc_usdt")
		So(err, ShouldNotBeNil)
	})
}

func TestOpenOrders(t *testing.T) {
	Convey("should return pendding orders unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		_, err := OpenOrders("btc_usdt")
		So(err, ShouldNotBeNil)
	})
}

func TestBorrowOrders(t *testing.T) {
	Convey("should return borrow orders unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		_, err := BorrowOrders("btc_usdt")
		So(err, ShouldNotBeNil)
	})
}

func TestMarginTrade(t *testing.T) {
	Convey("should trade unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		_, err := MarginTrade("TEST", "btc_usdt", 1)
		So(err, ShouldNotBeNil)
	})
}

func TestBorrowAvailable(t *testing.T) {
	Convey("should return available amount unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		_, err := BorrowAvailable("btc_usdt", "usdt")
		So(err, ShouldNotBeNil)
	})
}

func TestBorrow(t *testing.T) {
	Convey("should return borrow unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		currency := "btc_usdt"
		symbol := "usdt"

		amount, err := BorrowAvailable(currency, symbol)
		So(err, ShouldNotBeNil)
		if amount != 0 {
			_, err := Borrow(currency, symbol, amount)
			So(err, ShouldBeNil)
		}
	})
}

func TestRepay(t *testing.T) {
	Convey("should return repay all debt unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		err := Repay("btc_usdt", "usdt")
		So(err, ShouldNotBeNil)
	})
}

func TestCancelAll(t *testing.T) {
	Convey("should return cancel all pendding orders unsuccessfully", t, func() {
		Init("apikey", "secretkey")
		err := CancelAll("btc_usdt")
		So(err, ShouldNotBeNil)
	})
}
