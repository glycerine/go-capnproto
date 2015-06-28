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
	base.SetMaxSpeed(0.5)
	homes := air.NewAirportList(seg, 2)
	homes.Set(0, air.AIRPORT_JFK)
	homes.Set(1, air.AIRPORT_SFO)
	base.SetHomes(homes)
	b737.SetBase(base)
	airc.SetB737(b737)
	z.SetAircraft(airc)
	lit, err := z.MarshalCapLit()
	panicOn(err)
	json, err := z.MarshalJSON()
	panicOn(err)

	cv.Convey("Given the aircraftlib schema (and an Aircraft value), we should generate a MarshalCapLit() function that returns a literal representation in bytes for the given Aircraft value. And the MarshalJSON() should return the expected format too.", t, func() {
		cv.So(string(lit), cv.ShouldEqual, `(aircraft = (b737 = (base = (name = "helen", homes = [jfk, sfo], rating = 0, canFly = false, capacity = 0, maxSpeed = 0.5))))`)
		cv.So(string(json), cv.ShouldEqual, `{"aircraft":{"b737":{"base":{"name":"helen","homes":["jfk", "sfo"],"rating":0,"canFly":false,"capacity":0,"maxSpeed":0.5}}}}`)
	})

}

func panicOn(err error) {
	if err != nil {
		panic(err)
	}
}
