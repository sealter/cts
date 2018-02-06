package dingtalk

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInit(t *testing.T) {
	Convey("should init successfully", t, func() {
		So(token, ShouldEqual, "")
		Init("accesstoken")
		So(token, ShouldEqual, "accesstoken")
	})
}

func TestPush(t *testing.T) {
	Convey("should push unsuccessfully", t, func() {
		Init("accesstoken")
		err := Push("Hello robot")
		So(err, ShouldNotBeNil)
	})
}
