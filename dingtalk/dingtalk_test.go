package ding

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInit(t *testing.T) {
	Convey("should init successfully", t, func() {
		So(token, ShouldEqual, "")
		Init("access token")
		So(token, ShouldEqual, "access token")
	})
}

func TestPush(t *testing.T) {
	Convey("should push unsuccessfully", t, func() {
		Init("access token")
		err := Push("Hello robot")
		So(err, ShouldBeNil)
	})
}
