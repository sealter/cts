package strategy

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRippleDoge(t *testing.T) {
	Convey("should work successfully(strategy: ripdog)", t, func(c C) {
		ch := make(chan uint8)

		go func() {
			s := RippleDoge{}

			name := s.Name()
			c.So(name, ShouldEqual, "ripdog")

			sig, err := s.Signal()
			c.So(err, ShouldBeNil)

			ch <- sig
			close(ch)
		}()

		c.So([]uint8{SigNone, SigRise, SigFall}, ShouldContain, <-ch)
	})
}
