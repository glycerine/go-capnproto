package capn_test

import (
	"testing"

	"github.com/glycerine/go-capnproto"
	air "github.com/glycerine/go-capnproto/aircraftlib"
	cv "github.com/glycerine/goconvey/convey"
)

// For enums like Airport, their lists should have a set method
func TestEnumListHasSet(t *testing.T) {

	seg := capn.NewBuffer(nil)
	z := air.NewRootZ(seg)
	airc := air.AutoNewAircraft(seg)
	b737 := air.AutoNewB737(seg)
	base := air.AutoNewPlaneBase(seg)
	base.SetName("helen")
	homes := air.NewAirportList(seg, 2)

	// test is here!!
	// these next two lines should compile, because there should be a Set method.
	homes.Set(0, air.AIRPORT_JFK)
	homes.Set(1, air.AIRPORT_SFO)
	base.SetHomes(homes)
	b737.SetBase(base)
	airc.SetB737(b737)
	z.SetAircraft(airc)
	j, err := z.MarshalCapLit()
	panicOn(err)

	cv.Convey("To confirm that enum lists have a Set() method: Given the aircraftlib schema (and an Aircraft value), we should generate a MarshalCapLit() function that returns a literal representation in bytes for the given Aircraft value ", t, func() {
		cv.So(string(j), cv.ShouldEqual, `(aircraft = (b737 = (base = (name = "helen", homes = [jfk, sfo], rating = 0, canFly = false, capacity = 0, maxSpeed = 0))))`)
	})

}
