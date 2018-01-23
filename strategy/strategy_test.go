package strategy

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRippleDoge(t *testing.T) {
	Convey("should check ripple and doge successfully", t, func(c C) {
		ch := make(chan uint8)

		go func() {
			sig, err := RippleDoge()
			c.So(err, ShouldBeNil)
			ch <- sig
			close(ch)
		}()

		c.So([]uint8{SIG_NONE, SIG_RISE, SIG_FALL}, ShouldContain, <-ch)
	})
}
