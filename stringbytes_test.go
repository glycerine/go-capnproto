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
/*
// Confirmed: StringBytes lets us avoid all allcations:
go test -v -run 500 -benchmem -bench=String
PASS
BenchmarkStringBytes-4       	50000000	        29.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutStringBytes-4	20000000	        79.0 ns/op	      32 B/op	       1 allocs/op
ok  	github.com/glycerine/go-capnproto	3.255s
$
$ go test -v -run=BBBBBBBB -bench=. -benchmem
PASS
BenchmarkPopulateCapnp-4                	 1000000	      1571 ns/op	      48 B/op	       1 allocs/op
BenchmarkMarshalCapnp-4                 	20000000	       118 ns/op	3927.52 MB/s	      16 B/op	       1 allocs/op
BenchmarkUnmarshalCapnp-4               	 2000000	       605 ns/op	 766.08 MB/s	     256 B/op	       5 allocs/op
BenchmarkUnmarshalCapnpZeroCopy-4       	10000000	       182 ns/op	2539.81 MB/s	      88 B/op	       3 allocs/op
BenchmarkUnmarshalCapnpZeroCopyNoAlloc-4	100000000	        24.9 ns/op	18668.12 MB/s	       0 B/op	       0 allocs/op
BenchmarkCompressor-4                   	  200000	      6651 ns/op	 247.78 MB/s	      16 B/op	       1 allocs/op
BenchmarkDecompressor-4                 	  100000	     17557 ns/op	  57.81 MB/s	    3296 B/op	     206 allocs/op
BenchmarkStringBytes-4                  	50000000	        29.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutStringBytes-4           	20000000	        78.6 ns/op	      32 B/op	       1 allocs/op
BenchmarkTextMovementBetweenSegments-4  	     500	   3562898 ns/op	  208304 B/op	    3005 allocs/op
ok  	github.com/glycerine/go-capnproto	19.246s
$
*/
func Test500StringBytesWorksAndDoesNoAllocation(t *testing.T) {

	baseBytes := CapnpEncode(`(name = "An Airport base station")`, "PlaneBase")
	bagBytes := CapnpEncode(`(counter = (size = 9, wordlist = ["hello","bye"]))`, "Bag")

	cv.Convey("Given an capnp serialized data segment containing strings or vectors of strings", t, func() {
		cv.Convey("We should be able to use StringBytes() to avoid copying data, "+
			"instead just getting a []byte back", func() {

			// Base - for standalone string field.
			multiBase := capn.NewSingleSegmentMultiBuffer()
			var err error
			_, err = capn.ReadFromMemoryZeroCopyNoAlloc(baseBytes, multiBase)
			if err != nil {
				panic(err)
			}
			seg := multiBase.Segments[0]

			base := air.ReadRootPlaneBase(seg)

			fmt.Printf("\n base.Name() = '%s'\n", base.Name())
			fmt.Printf(" base.NameBytes() = '%s'\n", base.NameBytes())
			cv.So(string(base.NameBytes()), cv.ShouldResemble, base.Name())

			// Bag - for vector of string, aka TextList
			_, err = capn.ReadFromMemoryZeroCopyNoAlloc(bagBytes, multiBase)
			if err != nil {
				panic(err)
			}
			seg = multiBase.Segments[0]
			bag := air.ReadRootBag(seg)

			fmt.Printf("\n bag.Counter().Wordlist().AtAsBytes(0) = '%s'\n", string(bag.Counter().Wordlist().AtAsBytes(0)))
			fmt.Printf(" bag.Counter().Wordlist().At(0) = '%s'\n", bag.Counter().Wordlist().At(0))
			cv.So(string(bag.Counter().Wordlist().AtAsBytes(0)), cv.ShouldResemble, bag.Counter().Wordlist().At(0))
		})
	})
}

func BenchmarkStringBytes(b *testing.B) {

	baseBytes := CapnpEncode(`(name = "An Airport base station")`, "PlaneBase")
	multiBase := capn.NewSingleSegmentMultiBuffer()
	var err error
	_, err = capn.ReadFromMemoryZeroCopyNoAlloc(baseBytes, multiBase)
	if err != nil {
		panic(err)
	}
	seg := multiBase.Segments[0]
	base := air.ReadRootPlaneBase(seg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := base.NameBytes()
		_ = name
	}
}

func BenchmarkWithoutStringBytes(b *testing.B) {

	baseBytes := CapnpEncode(`(name = "An Airport base station")`, "PlaneBase")
	multiBase := capn.NewSingleSegmentMultiBuffer()
	var err error
	_, err = capn.ReadFromMemoryZeroCopyNoAlloc(baseBytes, multiBase)
	if err != nil {
		panic(err)
	}
	seg := multiBase.Segments[0]
	base := air.ReadRootPlaneBase(seg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := base.Name()
		_ = name
	}
}
