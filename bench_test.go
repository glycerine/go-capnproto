package capn_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"
	"unsafe"

	capn "github.com/glycerine/go-capnproto"
)

// Albert's benchmarks:

func newCapnpLog(record *Log) {
	record.SetTimestamp(time.Now().UnixNano())
	record.SetZoneId(123456)
	record.SetZonePlan(ZONEPLAN_FREE)

	http := NewHTTP(record.Segment)
	record.SetHttp(http)
	http.SetProtocol(HTTPPROTOCOL_HTTP11)
	http.SetStatus(200)
	http.SetHostStatus(503)
	http.SetUpStatus(520)
	http.SetMethod(HTTPMETHOD_GET)
	http.SetContentType("text/html")
	http.SetUserAgent(userAgent)
	http.SetReferer("http://www.w3.org/hypertext/DataSources/Overview.html")
	http.SetRequestURI("/a9pPJR1.jpg")

	origin := NewOrigin(record.Segment)
	record.SetOrigin(origin)
	origin.SetIp([]byte(net.IPv4(1, 2, 3, 4).To4()))
	origin.SetPort(8080)
	origin.SetHostname("www.cloudflare.com")
	origin.SetProtocol(ORIGINPROTOCOL_HTTPS)

	record.SetCacheStatus(CACHESTATUS_HIT)
	record.SetServerIp([]byte(net.IPv4(192, 168, 1, 1).To4()))
	record.SetServerName("metal.cloudflare.com")
	record.SetRemoteIp([]byte(net.IPv4(10, 1, 2, 3).To4()))
	record.SetBytesDlv(123456)
	record.SetRayId("10c73629cce30078-LAX")
}

func BenchmarkPopulateCapnp(b *testing.B) {
	buf := make([]byte, 1<<20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		segment := capn.NewBuffer(buf[:0])
		record := NewRootLog(segment)
		newCapnpLog(&record)
	}
}

func BenchmarkMarshalCapnp(b *testing.B) {
	segment := capn.NewBuffer(make([]byte, 0, 1<<20))
	record := NewRootLog(segment)
	newCapnpLog(&record)

	var buf bytes.Buffer
	_, err := segment.WriteTo(&buf)
	if err != nil {
		b.Fatalf("WriteTo: %v", err)
	}
	b.SetBytes(int64(len(buf.Bytes())))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_, err := segment.WriteTo(&buf)
		if err != nil {
			b.Fatalf("WriteTo: %v", err)
		}
	}
}

func BenchmarkUnmarshalCapnp(b *testing.B) {
	segment := capn.NewBuffer(make([]byte, 0, 1<<20))
	record := NewRootLog(segment)
	newCapnpLog(&record)

	var buf bytes.Buffer
	_, err := segment.WriteTo(&buf)
	if err != nil {
		b.Fatalf("WriteTo: %v", err)
	}
	b.SetBytes(int64(len(buf.Bytes())))

	segmentBuf := bytes.NewBuffer(make([]byte, 0, 1<<20))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewBuffer(buf.Bytes())
		seg, err := capn.ReadFromStream(r, segmentBuf)
		if err != nil {
			b.Fatalf("ReadFromStream: %v", err)
		}
		record := ReadRootLog(seg)
		_ = record
	}
}

func BenchmarkUnmarshalCapnpZeroCopy(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seg, _, err := capn.ReadFromMemoryZeroCopy(data)
		if err != nil {
			b.Fatalf("ReadFromStream: %v", err)
		}
		_ = seg
		//record := ReadRootLog(seg)
		//_ = record
	}
}

func BenchmarkUnmarshalCapnpZeroCopyNoAlloc(b *testing.B) {
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

const userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/33.0.1750.146 Safari/537.36"

type HTTP capn.Struct

func NewHTTP(s *capn.Segment) HTTP        { return HTTP(s.NewStruct(16, 4)) }
func NewRootHTTP(s *capn.Segment) HTTP    { return HTTP(s.NewRootStruct(16, 4)) }
func ReadRootHTTP(s *capn.Segment) HTTP   { return HTTP(s.Root(0).ToStruct()) }
func (s HTTP) Protocol() HTTPProtocol     { return HTTPProtocol(capn.Struct(s).Get16(0)) }
func (s HTTP) SetProtocol(v HTTPProtocol) { capn.Struct(s).Set16(0, uint16(v)) }
func (s HTTP) Status() uint16             { return capn.Struct(s).Get16(2) }
func (s HTTP) SetStatus(v uint16)         { capn.Struct(s).Set16(2, v) }
func (s HTTP) HostStatus() uint16         { return capn.Struct(s).Get16(4) }
func (s HTTP) SetHostStatus(v uint16)     { capn.Struct(s).Set16(4, v) }
func (s HTTP) UpStatus() uint16           { return capn.Struct(s).Get16(6) }
func (s HTTP) SetUpStatus(v uint16)       { capn.Struct(s).Set16(6, v) }
func (s HTTP) Method() HTTPMethod         { return HTTPMethod(capn.Struct(s).Get16(8)) }
func (s HTTP) SetMethod(v HTTPMethod)     { capn.Struct(s).Set16(8, uint16(v)) }
func (s HTTP) ContentType() string        { return capn.Struct(s).GetObject(0).ToText() }
func (s HTTP) SetContentType(v string)    { capn.Struct(s).SetObject(0, s.Segment.NewText(v)) }
func (s HTTP) UserAgent() string          { return capn.Struct(s).GetObject(1).ToText() }
func (s HTTP) SetUserAgent(v string)      { capn.Struct(s).SetObject(1, s.Segment.NewText(v)) }
func (s HTTP) Referer() string            { return capn.Struct(s).GetObject(2).ToText() }
func (s HTTP) SetReferer(v string)        { capn.Struct(s).SetObject(2, s.Segment.NewText(v)) }
func (s HTTP) RequestURI() string         { return capn.Struct(s).GetObject(3).ToText() }
func (s HTTP) SetRequestURI(v string)     { capn.Struct(s).SetObject(3, s.Segment.NewText(v)) }

type HTTP_List capn.PointerList

func NewHTTPList(s *capn.Segment, sz int) HTTP_List { return HTTP_List(s.NewCompositeList(16, 4, sz)) }
func (s HTTP_List) Len() int                        { return capn.PointerList(s).Len() }
func (s HTTP_List) At(i int) HTTP                   { return HTTP(capn.PointerList(s).At(i).ToStruct()) }
func (s HTTP_List) ToArray() []HTTP                 { return *(*[]HTTP)(unsafe.Pointer(capn.PointerList(s).ToArray())) }

type HTTPProtocol uint16

const (
	HTTPPROTOCOL_UNKNOWN HTTPProtocol = 0
	HTTPPROTOCOL_HTTP10               = 1
	HTTPPROTOCOL_HTTP11               = 2
	HTTPPROTOCOL_MAX                  = 3
)

func (c HTTPProtocol) String() string {
	switch c {
	case HTTPPROTOCOL_UNKNOWN:
		return "unknown"
	case HTTPPROTOCOL_HTTP10:
		return "http10"
	case HTTPPROTOCOL_HTTP11:
		return "http11"
	case HTTPPROTOCOL_MAX:
		return "max"
	default:
		return ""
	}
}

type HTTPProtocol_List capn.PointerList

func NewHTTPProtocolList(s *capn.Segment, sz int) HTTPProtocol_List {
	return HTTPProtocol_List(s.NewUInt16List(sz))
}
func (s HTTPProtocol_List) Len() int              { return capn.UInt16List(s).Len() }
func (s HTTPProtocol_List) At(i int) HTTPProtocol { return HTTPProtocol(capn.UInt16List(s).At(i)) }
func (s HTTPProtocol_List) ToArray() []HTTPProtocol {
	return *(*[]HTTPProtocol)(unsafe.Pointer(capn.UInt16List(s).ToEnumArray()))
}

type HTTPMethod uint16

const (
	HTTPMETHOD_UNKNOWN  HTTPMethod = 0
	HTTPMETHOD_GET                 = 1
	HTTPMETHOD_POST                = 2
	HTTPMETHOD_DELETE              = 3
	HTTPMETHOD_PUT                 = 4
	HTTPMETHOD_HEAD                = 5
	HTTPMETHOD_PURGE               = 6
	HTTPMETHOD_OPTIONS             = 7
	HTTPMETHOD_PROPFIND            = 8
	HTTPMETHOD_MKCOL               = 9
	HTTPMETHOD_PATCH               = 10
	HTTPMETHOD_MAX                 = 11
)

func (c HTTPMethod) String() string {
	switch c {
	case HTTPMETHOD_UNKNOWN:
		return "unknown"
	case HTTPMETHOD_GET:
		return "get"
	case HTTPMETHOD_POST:
		return "post"
	case HTTPMETHOD_DELETE:
		return "delete"
	case HTTPMETHOD_PUT:
		return "put"
	case HTTPMETHOD_HEAD:
		return "head"
	case HTTPMETHOD_PURGE:
		return "purge"
	case HTTPMETHOD_OPTIONS:
		return "options"
	case HTTPMETHOD_PROPFIND:
		return "propfind"
	case HTTPMETHOD_MKCOL:
		return "mkcol"
	case HTTPMETHOD_PATCH:
		return "patch"
	case HTTPMETHOD_MAX:
		return "max"
	default:
		return ""
	}
}

type HTTPMethod_List capn.PointerList

func NewHTTPMethodList(s *capn.Segment, sz int) HTTPMethod_List {
	return HTTPMethod_List(s.NewUInt16List(sz))
}
func (s HTTPMethod_List) Len() int            { return capn.UInt16List(s).Len() }
func (s HTTPMethod_List) At(i int) HTTPMethod { return HTTPMethod(capn.UInt16List(s).At(i)) }
func (s HTTPMethod_List) ToArray() []HTTPMethod {
	return *(*[]HTTPMethod)(unsafe.Pointer(capn.UInt16List(s).ToEnumArray()))
}

type CacheStatus uint16

const (
	CACHESTATUS_UNKNOWN CacheStatus = 0
	CACHESTATUS_MISS                = 1
	CACHESTATUS_EXPIRED             = 2
	CACHESTATUS_HIT                 = 3
	CACHESTATUS_MAX                 = 4
)

func (c CacheStatus) String() string {
	switch c {
	case CACHESTATUS_UNKNOWN:
		return "unknown"
	case CACHESTATUS_MISS:
		return "miss"
	case CACHESTATUS_EXPIRED:
		return "expired"
	case CACHESTATUS_HIT:
		return "hit"
	case CACHESTATUS_MAX:
		return "max"
	default:
		return ""
	}
}

type CacheStatus_List capn.PointerList

func NewCacheStatusList(s *capn.Segment, sz int) CacheStatus_List {
	return CacheStatus_List(s.NewUInt16List(sz))
}
func (s CacheStatus_List) Len() int             { return capn.UInt16List(s).Len() }
func (s CacheStatus_List) At(i int) CacheStatus { return CacheStatus(capn.UInt16List(s).At(i)) }
func (s CacheStatus_List) ToArray() []CacheStatus {
	return *(*[]CacheStatus)(unsafe.Pointer(capn.UInt16List(s).ToEnumArray()))
}

type Origin capn.Struct

func NewOrigin(s *capn.Segment) Origin        { return Origin(s.NewStruct(8, 2)) }
func NewRootOrigin(s *capn.Segment) Origin    { return Origin(s.NewRootStruct(8, 2)) }
func ReadRootOrigin(s *capn.Segment) Origin   { return Origin(s.Root(0).ToStruct()) }
func (s Origin) Ip() []byte                   { return capn.Struct(s).GetObject(0).ToData() }
func (s Origin) SetIp(v []byte)               { capn.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s Origin) Port() uint16                 { return capn.Struct(s).Get16(0) }
func (s Origin) SetPort(v uint16)             { capn.Struct(s).Set16(0, v) }
func (s Origin) Hostname() string             { return capn.Struct(s).GetObject(1).ToText() }
func (s Origin) SetHostname(v string)         { capn.Struct(s).SetObject(1, s.Segment.NewText(v)) }
func (s Origin) Protocol() OriginProtocol     { return OriginProtocol(capn.Struct(s).Get16(2)) }
func (s Origin) SetProtocol(v OriginProtocol) { capn.Struct(s).Set16(2, uint16(v)) }

type Origin_List capn.PointerList

func NewOriginList(s *capn.Segment, sz int) Origin_List {
	return Origin_List(s.NewCompositeList(8, 2, sz))
}
func (s Origin_List) Len() int        { return capn.PointerList(s).Len() }
func (s Origin_List) At(i int) Origin { return Origin(capn.PointerList(s).At(i).ToStruct()) }
func (s Origin_List) ToArray() []Origin {
	return *(*[]Origin)(unsafe.Pointer(capn.PointerList(s).ToArray()))
}

type OriginProtocol uint16

const (
	ORIGINPROTOCOL_UNKNOWN OriginProtocol = 0
	ORIGINPROTOCOL_HTTP                   = 1
	ORIGINPROTOCOL_HTTPS                  = 2
	ORIGINPROTOCOL_MAX                    = 3
)

func (c OriginProtocol) String() string {
	switch c {
	case ORIGINPROTOCOL_UNKNOWN:
		return "unknown"
	case ORIGINPROTOCOL_HTTP:
		return "http"
	case ORIGINPROTOCOL_HTTPS:
		return "https"
	case ORIGINPROTOCOL_MAX:
		return "max"
	default:
		return ""
	}
}

type OriginProtocol_List capn.PointerList

func NewOriginProtocolList(s *capn.Segment, sz int) OriginProtocol_List {
	return OriginProtocol_List(s.NewUInt16List(sz))
}
func (s OriginProtocol_List) Len() int                { return capn.UInt16List(s).Len() }
func (s OriginProtocol_List) At(i int) OriginProtocol { return OriginProtocol(capn.UInt16List(s).At(i)) }
func (s OriginProtocol_List) ToArray() []OriginProtocol {
	return *(*[]OriginProtocol)(unsafe.Pointer(capn.UInt16List(s).ToEnumArray()))
}

type ZonePlan uint16

const (
	ZONEPLAN_UNKNOWN ZonePlan = 0
	ZONEPLAN_FREE             = 1
	ZONEPLAN_PRO              = 2
	ZONEPLAN_BIZ              = 3
	ZONEPLAN_ENT              = 4
	ZONEPLAN_MAX              = 5
)

func (c ZonePlan) String() string {
	switch c {
	case ZONEPLAN_FREE:
		return "Free"
	case ZONEPLAN_PRO:
		return "Pro"
	case ZONEPLAN_BIZ:
		return "Business"
	case ZONEPLAN_ENT:
		return "Enterprise"
	case ZONEPLAN_MAX:
		return "max"
	default:
		return ""
	}
}

type ZonePlan_List capn.PointerList

func NewZonePlanList(s *capn.Segment, sz int) ZonePlan_List { return ZonePlan_List(s.NewUInt16List(sz)) }
func (s ZonePlan_List) Len() int                            { return capn.UInt16List(s).Len() }
func (s ZonePlan_List) At(i int) ZonePlan                   { return ZonePlan(capn.UInt16List(s).At(i)) }
func (s ZonePlan_List) ToArray() []ZonePlan {
	return *(*[]ZonePlan)(unsafe.Pointer(capn.UInt16List(s).ToEnumArray()))
}

type Log capn.Struct

func NewLog(s *capn.Segment) Log           { return Log(s.NewStruct(32, 6)) }
func NewRootLog(s *capn.Segment) Log       { return Log(s.NewRootStruct(32, 6)) }
func ReadRootLog(s *capn.Segment) Log      { return Log(s.Root(0).ToStruct()) }
func (s Log) Timestamp() int64             { return int64(capn.Struct(s).Get64(0)) }
func (s Log) SetTimestamp(v int64)         { capn.Struct(s).Set64(0, uint64(v)) }
func (s Log) ZoneId() uint32               { return capn.Struct(s).Get32(8) }
func (s Log) SetZoneId(v uint32)           { capn.Struct(s).Set32(8, v) }
func (s Log) ZonePlan() ZonePlan           { return ZonePlan(capn.Struct(s).Get16(12)) }
func (s Log) SetZonePlan(v ZonePlan)       { capn.Struct(s).Set16(12, uint16(v)) }
func (s Log) Http() HTTP                   { return HTTP(capn.Struct(s).GetObject(0).ToStruct()) }
func (s Log) SetHttp(v HTTP)               { capn.Struct(s).SetObject(0, capn.Object(v)) }
func (s Log) Origin() Origin               { return Origin(capn.Struct(s).GetObject(1).ToStruct()) }
func (s Log) SetOrigin(v Origin)           { capn.Struct(s).SetObject(1, capn.Object(v)) }
func (s Log) Country() Country             { return Country(capn.Struct(s).Get16(14)) }
func (s Log) SetCountry(v Country)         { capn.Struct(s).Set16(14, uint16(v)) }
func (s Log) CacheStatus() CacheStatus     { return CacheStatus(capn.Struct(s).Get16(16)) }
func (s Log) SetCacheStatus(v CacheStatus) { capn.Struct(s).Set16(16, uint16(v)) }
func (s Log) ServerIp() []byte             { return capn.Struct(s).GetObject(2).ToData() }
func (s Log) SetServerIp(v []byte)         { capn.Struct(s).SetObject(2, s.Segment.NewData(v)) }
func (s Log) ServerName() string           { return capn.Struct(s).GetObject(3).ToText() }
func (s Log) SetServerName(v string)       { capn.Struct(s).SetObject(3, s.Segment.NewText(v)) }
func (s Log) RemoteIp() []byte             { return capn.Struct(s).GetObject(4).ToData() }
func (s Log) SetRemoteIp(v []byte)         { capn.Struct(s).SetObject(4, s.Segment.NewData(v)) }
func (s Log) BytesDlv() uint64             { return capn.Struct(s).Get64(24) }
func (s Log) SetBytesDlv(v uint64)         { capn.Struct(s).Set64(24, v) }
func (s Log) RayId() string                { return capn.Struct(s).GetObject(5).ToText() }
func (s Log) SetRayId(v string)            { capn.Struct(s).SetObject(5, s.Segment.NewText(v)) }

type Log_List capn.PointerList

func NewLogList(s *capn.Segment, sz int) Log_List { return Log_List(s.NewCompositeList(32, 6, sz)) }
func (s Log_List) Len() int                       { return capn.PointerList(s).Len() }
func (s Log_List) At(i int) Log                   { return Log(capn.PointerList(s).At(i).ToStruct()) }
func (s Log_List) ToArray() []Log                 { return *(*[]Log)(unsafe.Pointer(capn.PointerList(s).ToArray())) }

type Country uint16

const (
	COUNTRY_UNKNOWN Country = 0
	COUNTRY_A1              = 1
	COUNTRY_A2              = 2
	COUNTRY_O1              = 3
	COUNTRY_AD              = 4
	COUNTRY_AE              = 5
	COUNTRY_AF              = 6
	COUNTRY_AG              = 7
	COUNTRY_AI              = 8
	COUNTRY_AL              = 9
	COUNTRY_AM              = 10
	COUNTRY_AO              = 11
	COUNTRY_AP              = 12
	COUNTRY_AQ              = 13
	COUNTRY_AR              = 14
	COUNTRY_AS              = 15
	COUNTRY_AT              = 16
	COUNTRY_AU              = 17
	COUNTRY_AW              = 18
	COUNTRY_AX              = 19
	COUNTRY_AZ              = 20
	COUNTRY_BA              = 21
	COUNTRY_BB              = 22
	COUNTRY_BD              = 23
	COUNTRY_BE              = 24
	COUNTRY_BF              = 25
	COUNTRY_BG              = 26
	COUNTRY_BH              = 27
	COUNTRY_BI              = 28
	COUNTRY_BJ              = 29
	COUNTRY_BL              = 30
	COUNTRY_BM              = 31
	COUNTRY_BN              = 32
	COUNTRY_BO              = 33
	COUNTRY_BQ              = 34
	COUNTRY_BR              = 35
	COUNTRY_BS              = 36
	COUNTRY_BT              = 37
	COUNTRY_BV              = 38
	COUNTRY_BW              = 39
	COUNTRY_BY              = 40
	COUNTRY_BZ              = 41
	COUNTRY_CA              = 42
	COUNTRY_CC              = 43
	COUNTRY_CD              = 44
	COUNTRY_CF              = 45
	COUNTRY_CG              = 46
	COUNTRY_CH              = 47
	COUNTRY_CI              = 48
	COUNTRY_CK              = 49
	COUNTRY_CL              = 50
	COUNTRY_CM              = 51
	COUNTRY_CN              = 52
	COUNTRY_CO              = 53
	COUNTRY_CR              = 54
	COUNTRY_CU              = 55
	COUNTRY_CV              = 56
	COUNTRY_CW              = 57
	COUNTRY_CX              = 58
	COUNTRY_CY              = 59
	COUNTRY_CZ              = 60
	COUNTRY_DE              = 61
	COUNTRY_DJ              = 62
	COUNTRY_DK              = 63
	COUNTRY_DM              = 64
	COUNTRY_DO              = 65
	COUNTRY_DZ              = 66
	COUNTRY_EC              = 67
	COUNTRY_EE              = 68
	COUNTRY_EG              = 69
	COUNTRY_EH              = 70
	COUNTRY_ER              = 71
	COUNTRY_ES              = 72
	COUNTRY_ET              = 73
	COUNTRY_EU              = 74
	COUNTRY_FI              = 75
	COUNTRY_FJ              = 76
	COUNTRY_FK              = 77
	COUNTRY_FM              = 78
	COUNTRY_FO              = 79
	COUNTRY_FR              = 80
	COUNTRY_GA              = 81
	COUNTRY_GB              = 82
	COUNTRY_GD              = 83
	COUNTRY_GE              = 84
	COUNTRY_GF              = 85
	COUNTRY_GG              = 86
	COUNTRY_GH              = 87
	COUNTRY_GI              = 88
	COUNTRY_GL              = 89
	COUNTRY_GM              = 90
	COUNTRY_GN              = 91
	COUNTRY_GP              = 92
	COUNTRY_GQ              = 93
	COUNTRY_GR              = 94
	COUNTRY_GS              = 95
	COUNTRY_GT              = 96
	COUNTRY_GU              = 97
	COUNTRY_GW              = 98
	COUNTRY_GY              = 99
	COUNTRY_HK              = 100
	COUNTRY_HM              = 101
	COUNTRY_HN              = 102
	COUNTRY_HR              = 103
	COUNTRY_HT              = 104
	COUNTRY_HU              = 105
	COUNTRY_ID              = 106
	COUNTRY_IE              = 107
	COUNTRY_IL              = 108
	COUNTRY_IM              = 109
	COUNTRY_IN              = 110
	COUNTRY_IO              = 111
	COUNTRY_IQ              = 112
	COUNTRY_IR              = 113
	COUNTRY_IS              = 114
	COUNTRY_IT              = 115
	COUNTRY_JE              = 116
	COUNTRY_JM              = 117
	COUNTRY_JO              = 118
	COUNTRY_JP              = 119
	COUNTRY_KE              = 120
	COUNTRY_KG              = 121
	COUNTRY_KH              = 122
	COUNTRY_KI              = 123
	COUNTRY_KM              = 124
	COUNTRY_KN              = 125
	COUNTRY_KP              = 126
	COUNTRY_KR              = 127
	COUNTRY_KW              = 128
	COUNTRY_KY              = 129
	COUNTRY_KZ              = 130
	COUNTRY_LA              = 131
	COUNTRY_LB              = 132
	COUNTRY_LC              = 133
	COUNTRY_LI              = 134
	COUNTRY_LK              = 135
	COUNTRY_LR              = 136
	COUNTRY_LS              = 137
	COUNTRY_LT              = 138
	COUNTRY_LU              = 139
	COUNTRY_LV              = 140
	COUNTRY_LY              = 141
	COUNTRY_MA              = 142
	COUNTRY_MC              = 143
	COUNTRY_MD              = 144
	COUNTRY_ME              = 145
	COUNTRY_MF              = 146
	COUNTRY_MG              = 147
	COUNTRY_MH              = 148
	COUNTRY_MK              = 149
	COUNTRY_ML              = 150
	COUNTRY_MM              = 151
	COUNTRY_MN              = 152
	COUNTRY_MO              = 153
	COUNTRY_MP              = 154
	COUNTRY_MQ              = 155
	COUNTRY_MR              = 156
	COUNTRY_MS              = 157
	COUNTRY_MT              = 158
	COUNTRY_MU              = 159
	COUNTRY_MV              = 160
	COUNTRY_MW              = 161
	COUNTRY_MX              = 162
	COUNTRY_MY              = 163
	COUNTRY_MZ              = 164
	COUNTRY_NA              = 165
	COUNTRY_NC              = 166
	COUNTRY_NE              = 167
	COUNTRY_NF              = 168
	COUNTRY_NG              = 169
	COUNTRY_NI              = 170
	COUNTRY_NL              = 171
	COUNTRY_NO              = 172
	COUNTRY_NP              = 173
	COUNTRY_NR              = 174
	COUNTRY_NU              = 175
	COUNTRY_NZ              = 176
	COUNTRY_OM              = 177
	COUNTRY_PA              = 178
	COUNTRY_PE              = 179
	COUNTRY_PF              = 180
	COUNTRY_PG              = 181
	COUNTRY_PH              = 182
	COUNTRY_PK              = 183
	COUNTRY_PL              = 184
	COUNTRY_PM              = 185
	COUNTRY_PN              = 186
	COUNTRY_PR              = 187
	COUNTRY_PS              = 188
	COUNTRY_PT              = 189
	COUNTRY_PW              = 190
	COUNTRY_PY              = 191
	COUNTRY_QA              = 192
	COUNTRY_RE              = 193
	COUNTRY_RO              = 194
	COUNTRY_RS              = 195
	COUNTRY_RU              = 196
	COUNTRY_RW              = 197
	COUNTRY_SA              = 198
	COUNTRY_SB              = 199
	COUNTRY_SC              = 200
	COUNTRY_SD              = 201
	COUNTRY_SE              = 202
	COUNTRY_SG              = 203
	COUNTRY_SH              = 204
	COUNTRY_SI              = 205
	COUNTRY_SJ              = 206
	COUNTRY_SK              = 207
	COUNTRY_SL              = 208
	COUNTRY_SM              = 209
	COUNTRY_SN              = 210
	COUNTRY_SO              = 211
	COUNTRY_SR              = 212
	COUNTRY_SS              = 213
	COUNTRY_ST              = 214
	COUNTRY_SV              = 215
	COUNTRY_SX              = 216
	COUNTRY_SY              = 217
	COUNTRY_SZ              = 218
	COUNTRY_TC              = 219
	COUNTRY_TD              = 220
	COUNTRY_TF              = 221
	COUNTRY_TG              = 222
	COUNTRY_TH              = 223
	COUNTRY_TJ              = 224
	COUNTRY_TK              = 225
	COUNTRY_TL              = 226
	COUNTRY_TM              = 227
	COUNTRY_TN              = 228
	COUNTRY_TO              = 229
	COUNTRY_TR              = 230
	COUNTRY_TT              = 231
	COUNTRY_TV              = 232
	COUNTRY_TW              = 233
	COUNTRY_TZ              = 234
	COUNTRY_UA              = 235
	COUNTRY_UG              = 236
	COUNTRY_UM              = 237
	COUNTRY_US              = 238
	COUNTRY_UY              = 239
	COUNTRY_UZ              = 240
	COUNTRY_VA              = 241
	COUNTRY_VC              = 242
	COUNTRY_VE              = 243
	COUNTRY_VG              = 244
	COUNTRY_VI              = 245
	COUNTRY_VN              = 246
	COUNTRY_VU              = 247
	COUNTRY_WF              = 248
	COUNTRY_WS              = 249
	COUNTRY_XX              = 250
	COUNTRY_YE              = 251
	COUNTRY_YT              = 252
	COUNTRY_ZA              = 253
	COUNTRY_ZM              = 254
	COUNTRY_ZW              = 255
	COUNTRY_MAX             = 256
)

func (c Country) String() string {
	switch c {
	case COUNTRY_UNKNOWN:
		return "unknown"
	case COUNTRY_A1:
		return "a1"
	case COUNTRY_A2:
		return "a2"
	case COUNTRY_O1:
		return "o1"
	case COUNTRY_AD:
		return "ad"
	case COUNTRY_AE:
		return "ae"
	case COUNTRY_AF:
		return "af"
	case COUNTRY_AG:
		return "ag"
	case COUNTRY_AI:
		return "ai"
	case COUNTRY_AL:
		return "al"
	case COUNTRY_AM:
		return "am"
	case COUNTRY_AO:
		return "ao"
	case COUNTRY_AP:
		return "ap"
	case COUNTRY_AQ:
		return "aq"
	case COUNTRY_AR:
		return "ar"
	case COUNTRY_AS:
		return "as"
	case COUNTRY_AT:
		return "at"
	case COUNTRY_AU:
		return "au"
	case COUNTRY_AW:
		return "aw"
	case COUNTRY_AX:
		return "ax"
	case COUNTRY_AZ:
		return "az"
	case COUNTRY_BA:
		return "ba"
	case COUNTRY_BB:
		return "bb"
	case COUNTRY_BD:
		return "bd"
	case COUNTRY_BE:
		return "be"
	case COUNTRY_BF:
		return "bf"
	case COUNTRY_BG:
		return "bg"
	case COUNTRY_BH:
		return "bh"
	case COUNTRY_BI:
		return "bi"
	case COUNTRY_BJ:
		return "bj"
	case COUNTRY_BL:
		return "bl"
	case COUNTRY_BM:
		return "bm"
	case COUNTRY_BN:
		return "bn"
	case COUNTRY_BO:
		return "bo"
	case COUNTRY_BQ:
		return "bq"
	case COUNTRY_BR:
		return "br"
	case COUNTRY_BS:
		return "bs"
	case COUNTRY_BT:
		return "bt"
	case COUNTRY_BV:
		return "bv"
	case COUNTRY_BW:
		return "bw"
	case COUNTRY_BY:
		return "by"
	case COUNTRY_BZ:
		return "bz"
	case COUNTRY_CA:
		return "ca"
	case COUNTRY_CC:
		return "cc"
	case COUNTRY_CD:
		return "cd"
	case COUNTRY_CF:
		return "cf"
	case COUNTRY_CG:
		return "cg"
	case COUNTRY_CH:
		return "ch"
	case COUNTRY_CI:
		return "ci"
	case COUNTRY_CK:
		return "ck"
	case COUNTRY_CL:
		return "cl"
	case COUNTRY_CM:
		return "cm"
	case COUNTRY_CN:
		return "cn"
	case COUNTRY_CO:
		return "co"
	case COUNTRY_CR:
		return "cr"
	case COUNTRY_CU:
		return "cu"
	case COUNTRY_CV:
		return "cv"
	case COUNTRY_CW:
		return "cw"
	case COUNTRY_CX:
		return "cx"
	case COUNTRY_CY:
		return "cy"
	case COUNTRY_CZ:
		return "cz"
	case COUNTRY_DE:
		return "de"
	case COUNTRY_DJ:
		return "dj"
	case COUNTRY_DK:
		return "dk"
	case COUNTRY_DM:
		return "dm"
	case COUNTRY_DO:
		return "do"
	case COUNTRY_DZ:
		return "dz"
	case COUNTRY_EC:
		return "ec"
	case COUNTRY_EE:
		return "ee"
	case COUNTRY_EG:
		return "eg"
	case COUNTRY_EH:
		return "eh"
	case COUNTRY_ER:
		return "er"
	case COUNTRY_ES:
		return "es"
	case COUNTRY_ET:
		return "et"
	case COUNTRY_EU:
		return "eu"
	case COUNTRY_FI:
		return "fi"
	case COUNTRY_FJ:
		return "fj"
	case COUNTRY_FK:
		return "fk"
	case COUNTRY_FM:
		return "fm"
	case COUNTRY_FO:
		return "fo"
	case COUNTRY_FR:
		return "fr"
	case COUNTRY_GA:
		return "ga"
	case COUNTRY_GB:
		return "gb"
	case COUNTRY_GD:
		return "gd"
	case COUNTRY_GE:
		return "ge"
	case COUNTRY_GF:
		return "gf"
	case COUNTRY_GG:
		return "gg"
	case COUNTRY_GH:
		return "gh"
	case COUNTRY_GI:
		return "gi"
	case COUNTRY_GL:
		return "gl"
	case COUNTRY_GM:
		return "gm"
	case COUNTRY_GN:
		return "gn"
	case COUNTRY_GP:
		return "gp"
	case COUNTRY_GQ:
		return "gq"
	case COUNTRY_GR:
		return "gr"
	case COUNTRY_GS:
		return "gs"
	case COUNTRY_GT:
		return "gt"
	case COUNTRY_GU:
		return "gu"
	case COUNTRY_GW:
		return "gw"
	case COUNTRY_GY:
		return "gy"
	case COUNTRY_HK:
		return "hk"
	case COUNTRY_HM:
		return "hm"
	case COUNTRY_HN:
		return "hn"
	case COUNTRY_HR:
		return "hr"
	case COUNTRY_HT:
		return "ht"
	case COUNTRY_HU:
		return "hu"
	case COUNTRY_ID:
		return "id"
	case COUNTRY_IE:
		return "ie"
	case COUNTRY_IL:
		return "il"
	case COUNTRY_IM:
		return "im"
	case COUNTRY_IN:
		return "in"
	case COUNTRY_IO:
		return "io"
	case COUNTRY_IQ:
		return "iq"
	case COUNTRY_IR:
		return "ir"
	case COUNTRY_IS:
		return "is"
	case COUNTRY_IT:
		return "it"
	case COUNTRY_JE:
		return "je"
	case COUNTRY_JM:
		return "jm"
	case COUNTRY_JO:
		return "jo"
	case COUNTRY_JP:
		return "jp"
	case COUNTRY_KE:
		return "ke"
	case COUNTRY_KG:
		return "kg"
	case COUNTRY_KH:
		return "kh"
	case COUNTRY_KI:
		return "ki"
	case COUNTRY_KM:
		return "km"
	case COUNTRY_KN:
		return "kn"
	case COUNTRY_KP:
		return "kp"
	case COUNTRY_KR:
		return "kr"
	case COUNTRY_KW:
		return "kw"
	case COUNTRY_KY:
		return "ky"
	case COUNTRY_KZ:
		return "kz"
	case COUNTRY_LA:
		return "la"
	case COUNTRY_LB:
		return "lb"
	case COUNTRY_LC:
		return "lc"
	case COUNTRY_LI:
		return "li"
	case COUNTRY_LK:
		return "lk"
	case COUNTRY_LR:
		return "lr"
	case COUNTRY_LS:
		return "ls"
	case COUNTRY_LT:
		return "lt"
	case COUNTRY_LU:
		return "lu"
	case COUNTRY_LV:
		return "lv"
	case COUNTRY_LY:
		return "ly"
	case COUNTRY_MA:
		return "ma"
	case COUNTRY_MC:
		return "mc"
	case COUNTRY_MD:
		return "md"
	case COUNTRY_ME:
		return "me"
	case COUNTRY_MF:
		return "mf"
	case COUNTRY_MG:
		return "mg"
	case COUNTRY_MH:
		return "mh"
	case COUNTRY_MK:
		return "mk"
	case COUNTRY_ML:
		return "ml"
	case COUNTRY_MM:
		return "mm"
	case COUNTRY_MN:
		return "mn"
	case COUNTRY_MO:
		return "mo"
	case COUNTRY_MP:
		return "mp"
	case COUNTRY_MQ:
		return "mq"
	case COUNTRY_MR:
		return "mr"
	case COUNTRY_MS:
		return "ms"
	case COUNTRY_MT:
		return "mt"
	case COUNTRY_MU:
		return "mu"
	case COUNTRY_MV:
		return "mv"
	case COUNTRY_MW:
		return "mw"
	case COUNTRY_MX:
		return "mx"
	case COUNTRY_MY:
		return "my"
	case COUNTRY_MZ:
		return "mz"
	case COUNTRY_NA:
		return "na"
	case COUNTRY_NC:
		return "nc"
	case COUNTRY_NE:
		return "ne"
	case COUNTRY_NF:
		return "nf"
	case COUNTRY_NG:
		return "ng"
	case COUNTRY_NI:
		return "ni"
	case COUNTRY_NL:
		return "nl"
	case COUNTRY_NO:
		return "no"
	case COUNTRY_NP:
		return "np"
	case COUNTRY_NR:
		return "nr"
	case COUNTRY_NU:
		return "nu"
	case COUNTRY_NZ:
		return "nz"
	case COUNTRY_OM:
		return "om"
	case COUNTRY_PA:
		return "pa"
	case COUNTRY_PE:
		return "pe"
	case COUNTRY_PF:
		return "pf"
	case COUNTRY_PG:
		return "pg"
	case COUNTRY_PH:
		return "ph"
	case COUNTRY_PK:
		return "pk"
	case COUNTRY_PL:
		return "pl"
	case COUNTRY_PM:
		return "pm"
	case COUNTRY_PN:
		return "pn"
	case COUNTRY_PR:
		return "pr"
	case COUNTRY_PS:
		return "ps"
	case COUNTRY_PT:
		return "pt"
	case COUNTRY_PW:
		return "pw"
	case COUNTRY_PY:
		return "py"
	case COUNTRY_QA:
		return "qa"
	case COUNTRY_RE:
		return "re"
	case COUNTRY_RO:
		return "ro"
	case COUNTRY_RS:
		return "rs"
	case COUNTRY_RU:
		return "ru"
	case COUNTRY_RW:
		return "rw"
	case COUNTRY_SA:
		return "sa"
	case COUNTRY_SB:
		return "sb"
	case COUNTRY_SC:
		return "sc"
	case COUNTRY_SD:
		return "sd"
	case COUNTRY_SE:
		return "se"
	case COUNTRY_SG:
		return "sg"
	case COUNTRY_SH:
		return "sh"
	case COUNTRY_SI:
		return "si"
	case COUNTRY_SJ:
		return "sj"
	case COUNTRY_SK:
		return "sk"
	case COUNTRY_SL:
		return "sl"
	case COUNTRY_SM:
		return "sm"
	case COUNTRY_SN:
		return "sn"
	case COUNTRY_SO:
		return "so"
	case COUNTRY_SR:
		return "sr"
	case COUNTRY_SS:
		return "ss"
	case COUNTRY_ST:
		return "st"
	case COUNTRY_SV:
		return "sv"
	case COUNTRY_SX:
		return "sx"
	case COUNTRY_SY:
		return "sy"
	case COUNTRY_SZ:
		return "sz"
	case COUNTRY_TC:
		return "tc"
	case COUNTRY_TD:
		return "td"
	case COUNTRY_TF:
		return "tf"
	case COUNTRY_TG:
		return "tg"
	case COUNTRY_TH:
		return "th"
	case COUNTRY_TJ:
		return "tj"
	case COUNTRY_TK:
		return "tk"
	case COUNTRY_TL:
		return "tl"
	case COUNTRY_TM:
		return "tm"
	case COUNTRY_TN:
		return "tn"
	case COUNTRY_TO:
		return "to"
	case COUNTRY_TR:
		return "tr"
	case COUNTRY_TT:
		return "tt"
	case COUNTRY_TV:
		return "tv"
	case COUNTRY_TW:
		return "tw"
	case COUNTRY_TZ:
		return "tz"
	case COUNTRY_UA:
		return "ua"
	case COUNTRY_UG:
		return "ug"
	case COUNTRY_UM:
		return "um"
	case COUNTRY_US:
		return "us"
	case COUNTRY_UY:
		return "uy"
	case COUNTRY_UZ:
		return "uz"
	case COUNTRY_VA:
		return "va"
	case COUNTRY_VC:
		return "vc"
	case COUNTRY_VE:
		return "ve"
	case COUNTRY_VG:
		return "vg"
	case COUNTRY_VI:
		return "vi"
	case COUNTRY_VN:
		return "vn"
	case COUNTRY_VU:
		return "vu"
	case COUNTRY_WF:
		return "wf"
	case COUNTRY_WS:
		return "ws"
	case COUNTRY_XX:
		return "xx"
	case COUNTRY_YE:
		return "ye"
	case COUNTRY_YT:
		return "yt"
	case COUNTRY_ZA:
		return "za"
	case COUNTRY_ZM:
		return "zm"
	case COUNTRY_ZW:
		return "zw"
	case COUNTRY_MAX:
		return "max"
	default:
		return ""
	}
}

type Country_List capn.PointerList

func NewCountryList(s *capn.Segment, sz int) Country_List { return Country_List(s.NewUInt16List(sz)) }
func (s Country_List) Len() int                           { return capn.UInt16List(s).Len() }
func (s Country_List) At(i int) Country                   { return Country(capn.UInt16List(s).At(i)) }
func (s Country_List) ToArray() []Country {
	return *(*[]Country)(unsafe.Pointer(capn.UInt16List(s).ToEnumArray()))
}
func (s Country) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	buf, err = json.Marshal(s.String())
	if err != nil {
		return err
	}
	_, err = b.Write(buf)
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s Country) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
