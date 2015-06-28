package capn_test

import (
	"testing"

	"github.com/glycerine/go-capnproto"
	air "github.com/glycerine/go-capnproto/aircraftlib"
	cv "github.com/glycerine/goconvey/convey"
)

func TestPrint(t *testing.T) {

	seg := capn.NewBuffer(nil)
	z := air.NewRootZ(seg)
	airc := air.AutoNewAircraft(seg)
	b737 := air.AutoNewB737(seg)
	base := air.AutoNewPlaneBase(seg)
	base.SetName("helen")
	b737.SetBase(base)
	airc.SetB737(b737)
	z.SetAircraft(airc)
	j, err := z.MarshalCapLit()
	panicOn(err)

	cv.Convey("Given the aircraftlib schema (and an Aircraft value), we should generate a MarshalCapLit() function that returns a literal representation in bytes for the given Aircraft value ", t, func() {
		cv.So(string(j), cv.ShouldEqual, `(aircraft = (b737 = (base = (name = "helen", homes = [], rating = 0, canFly = false, capacity = 0, maxSpeed = 0))))`)
	})

}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}
