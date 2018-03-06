package huobi

import (
	"fmt"
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

func TestSymbols(t *testing.T) {
	Convey("should return all support symbol successfully", t, func() {
		r, err := Symbols()
		So(err, ShouldBeNil)
		So(r, ShouldNotBeEmpty)
	})
}

func TestNewSymbol(t *testing.T) {
	Convey("should return new symbol successfully", t, func() {
		_, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		_, err = NewSymbol("btcusdt")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errInvalidSymbol.Error())

		_, err = NewSymbol("abc_def")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errUnsupportedSymbol.Error())
	})
}

func TestLimit(t *testing.T) {
	Convey("should return trade limit of symbol successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		r, err := s.Limit()
		So(err, ShouldBeNil)
		So(r.BuyGT, ShouldBeGreaterThan, 0)
		So(r.BuyLT, ShouldBeGreaterThan, 0)
		So(r.SellGT, ShouldBeGreaterThan, 0)
		So(r.SellLT, ShouldBeGreaterThan, 0)
	})
}

func TestAccount(t *testing.T) {
	Convey("should return margin account successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		r, err := s.Account()
		So(err, ShouldBeNil)
		So(r.ID, ShouldBeGreaterThan, 0)
		So(r.Type, ShouldEqual, "margin")
		So(r.List, ShouldNotBeEmpty)
	})
}

func TestOrders(t *testing.T) {
	Convey("should return finished orders successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		_, err = s.Orders()
		So(err, ShouldBeNil)
	})
}

func TestOpenOrders(t *testing.T) {
	Convey("should return pendding orders successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		_, err = s.OpenOrders("")
		So(err, ShouldBeNil)
	})
}

func TestBorrowOrders(t *testing.T) {
	Convey("should return borrow orders successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		_, err = s.BorrowOrders("")
		So(err, ShouldBeNil)
	})
}

func TestBorrowAvailable(t *testing.T) {
	Convey("should return available amount successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		_, err = s.BorrowAvailable(s.BaseCurrency)
		So(err, ShouldBeNil)

		_, err = s.BorrowAvailable(s.QuoteCurrency)
		So(err, ShouldBeNil)

		_, err = s.BorrowAvailable("unknown")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errInvalidCurrency.Error())
	})
}

func TestBorrow(t *testing.T) {
	Convey("should borrow successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		amount, err := s.BorrowAvailable(s.BaseCurrency)
		So(err, ShouldBeNil)
		if amount != 0 {
			fmt.Println("\n", amount)
			err := s.Borrow(s.BaseCurrency, 0.001)
			So(err, ShouldBeNil)
		}
	})
}

func TestRepay(t *testing.T) {
	Convey("should repay all debt successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		err = s.Repay("btc")
		So(err, ShouldBeNil)
	})
}

func TestTrade(t *testing.T) {
	Convey("should trade unsuccessfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		err = s.Trade("FUCK", 1)
		So(err, ShouldNotBeNil)

		err = s.Trade("TESTBUY", 1)
		So(err, ShouldBeNil)

		err = s.Trade("TESTSELL", 0.001)
		So(err, ShouldBeNil)
	})
}

func TestCancelAll(t *testing.T) {
	Convey("should cancel all open orders successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		err = s.CancelAll()
		So(err, ShouldBeNil)
	})
}

func TestAllIn(t *testing.T) {
	Convey("should cancel all open orders successfully", t, func() {
		s, err := NewSymbol("btc_usdt")
		So(err, ShouldBeNil)

		err = s.AllIn("BUY", true)
		So(err, ShouldBeNil)

		err = s.AllIn("SELL", true)
		So(err, ShouldBeNil)
	})
}
