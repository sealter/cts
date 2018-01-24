package strategy

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStrategies(t *testing.T) {
	Convey("should return all available strategy", t, func() {
		m := Strategies()

		ripdog, ok := m["ripdog"]
		So(ok, ShouldBeTrue)
		So(ripdog.Name(), ShouldEqual, "ripdog")
		So(ripdog.(*RippleDoge), ShouldHaveSameTypeAs, &RippleDoge{})
	})
}

func TestAvailable(t *testing.T) {
	Convey("should return all available strategy name", t, func() {
		s := Available()
		So(s, ShouldNotBeEmpty)
	})
}

func TestSignals(t *testing.T) {
	Convey("should return all available signal", t, func() {
		s := Signals()
		So(s, ShouldNotBeEmpty)
	})
}
