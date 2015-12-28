package capn_test

import (
	"fmt"
	"testing"

	capn "github.com/glycerine/go-capnproto"
	air "github.com/glycerine/go-capnproto/aircraftlib"
	cv "github.com/glycerine/goconvey/convey"
)

// StringBytes() should allow us to avoid the copying overhead of reading a string
// and making a copy.

func Test500StringBytesWorksAndDoesNoAllocation(t *testing.T) {

	baseBytes := CapnpEncode(`(name = "An Airport base station")`, "PlaneBase")
	//bagBytes := CapnpEncode(`(counter = (size = 9, wordlist = ["hello","bye"]))`, "Bag")

	cv.Convey("Given an capnp serialized data segment containing strings or vectors of strings", t, func() {
		cv.Convey("We should be able to use StringBytes() to avoid copying data, "+
			"instead just getting a []byte back", func() {

			multiBase := capn.NewSingleSegmentMultiBuffer()
			var err error
			_, err = capn.ReadFromMemoryZeroCopyNoAlloc(baseBytes, multiBase)
			seg := multiBase.Segments[0]

			base := air.ReadRootPlaneBase(seg)

			fmt.Printf("base.Name() = '%s'\n", base.Name())
			fmt.Printf("base.NameBytes() = '%s'\n", base.NameBytes())
			cv.So(string(base.NameBytes()), cv.ShouldResemble, base.Name())
			//cv.So(string(bag.wordlist.At(0).NameBytes()), cv.ShouldResemble, bag.wordlist.At(0).Name())
		})
	})
}

/*
func BenchmarkStringBytes(b *testing.B) {
	segment := capn.NewBuffer(make([]byte, 0, 1<<20))
	record := NewRootLog(segment)
	newCapnpLog(&record)

	var buf bytes.Buffer
	_, err := segment.WriteTo(&buf)
	if err != nil {
		b.Fatalf("WriteTo: %v", err)
	}
	b.SetBytes(int64(len(buf.Bytes())))

	data := buf.Bytes()

	multi := capn.NewSingleSegmentMultiBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = capn.ReadFromMemoryZeroCopyNoAlloc(data, multi)
		if err != nil {
			b.Fatalf("ReadFromStream: %v", err)
		}
		//record := ReadRootLog(seg)
		//_ = record
	}
}
*/
