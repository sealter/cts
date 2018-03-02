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
