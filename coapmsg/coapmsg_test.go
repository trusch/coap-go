package coapmsg

// https://github.com/dustin/go-coap
import (
	"bytes"
	"encoding"
	"fmt"
	"reflect"
	"testing"
)

var (
	_ = encoding.BinaryMarshaler(&Message{})
	_ = encoding.BinaryUnmarshaler(&Message{})
)

// assertEqualMessages compares the e(xptected) message to the a(ctual) message
// and reports any diffs with t.Errorf.
func assertEqualMessages(t *testing.T, e, a Message) {
	if e.Type != a.Type {
		t.Errorf("Expected type %v, got %v", e.Type, a.Type)
	}
	if e.Code != a.Code {
		t.Errorf("Expected code %v, got %v", e.Code, a.Code)
	}
	if e.MessageID != a.MessageID {
		t.Errorf("Expected MessageID %v, got %v", e.MessageID, a.MessageID)
	}
	if !bytes.Equal(e.Token, a.Token) {
		t.Errorf("Expected token %#v, got %#v", e.Token, a.Token)
	}
	if !bytes.Equal(e.Payload, a.Payload) {
		t.Errorf("Expected payload %#v, got %#v", e.Payload, a.Payload)
	}

	if len(e.options) != len(a.options) {
		t.Errorf("Expected %v options, got %v", len(e.options), len(a.options))
	} else {
		for id, vals := range e.options {
			if len(e.options[id]) != len(a.options[id]) {
				t.Errorf("Expected option ID %v length to be equal, got %v != %v", id, len(e.options[id]), len(a.options[id]))
				continue
			}

			for i, val := range vals {
				expected := val
				actual := a.options[id][i]
				if !bytes.Equal(expected.AsBytes(), actual.AsBytes()) {
					t.Errorf("Expected Option ID %v value %v, got %v", id, expected, actual)
				}
			}
		}
	}
}

func TestCode(t *testing.T) {
	if Created.Class() != 2 {
		t.Error("Expected Created.Class to be 2 but is", Created.Class())
	}
	if Created.Detail() != 1 {
		t.Error("Expected Created.Detail to be 1 but is", Created.Class())
	}
	if NotFound.Class() != 4 {
		t.Error("Expected NotFound.Class to be 4 but is", Created.Class())
	}
	if NotFound.Detail() != 4 {
		t.Error("Expected NotFound.Detail to be 4 but is", Created.Class())
	}
}

func TestSetOptions(t *testing.T) {
	msg := Message{}

	msg.Options().Set(ContentFormat, AppJSON)
	msg.Options().Add(ContentFormat, AppXML)

	if len(msg.options) != 1 {
		t.Error("Expected 1 option but got", len(msg.options))
	} else {
		if len(msg.options[ContentFormat]) != 2 {
			t.Error("Expected 2 ContentFormat options but got", len(msg.options[ContentFormat]))
		} else {
			if MediaType(msg.options[ContentFormat][0].AsUInt16()) != AppJSON {
				t.Error("Expected option value", AppJSON, "but got", msg.options[ContentFormat][0])
			}
			if MediaType(msg.options[ContentFormat][1].AsUInt16()) != AppXML {
				t.Error("Expected option value", AppXML, "but got", msg.options[ContentFormat][0])
			}
		}
	}
}

func TestMediaTypes(t *testing.T) {
	types := []interface{}{TextPlain, AppLinkFormat, AppXML, AppOctets, AppExi, AppJSON}
	exp := "coapmsg.MediaType"
	for _, typ := range types {
		if got := fmt.Sprintf("%T", typ); got != exp {
			t.Errorf("Error on %#v, expected %q, was %q", typ, exp, got)
		}
	}
}

func TestOptionToBytes(t *testing.T) {
	tests := []struct {
		in  interface{}
		exp []byte
	}{
		{"", []byte{}},
		{[]byte{}, []byte{}},
		{"x", []byte{'x'}},
		{[]byte{'x'}, []byte{'x'}},
		{MediaType(3), []byte{0x3}},
		{3, []byte{0x3}},
		{838, []byte{0x3, 0x46}},
		{int32(838), []byte{0x3, 0x46}},
		{uint(838), []byte{0x3, 0x46}},
		{uint32(838), []byte{0x3, 0x46}},
	}

	for _, test := range tests {
		op := option{Value: test.in}
		got := op.ToBytes()
		if !bytes.Equal(test.exp, got) {
			t.Errorf("Error on %T(%v), got %#v, wanted %#v",
				test.in, test.in, got, test.exp)
		}
	}
}

func TestMessageConfirmable(t *testing.T) {
	tests := []struct {
		m   Message
		exp bool
	}{
		{Message{Type: Confirmable}, true},
		{Message{Type: NonConfirmable}, false},
	}

	for _, test := range tests {
		got := test.m.IsConfirmable()
		if got != test.exp {
			t.Errorf("Expected %v for %v", test.exp, test.m)
		}
	}
}

func TestMissingOption(t *testing.T) {
	gotEmpty := Message{}.options.Get(MaxAge)
	if gotEmpty.Len() != 0 {
		t.Errorf("Expected empty slice, got %v", gotEmpty)
	}

	gotNil := Message{}.options[MaxAge]
	if gotNil != nil {
		t.Errorf("Expected nil, got %v", gotNil)
	}
}

func TestOptionToBytesPanic(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Error("Expected panic. Didn't")
		}
	}()
	option{Value: 3.1415926535897}.ToBytes()
}

func TestTypeString(t *testing.T) {
	tests := map[COAPType]string{
		Confirmable:    "Confirmable",
		NonConfirmable: "NonConfirmable",
		255:            "Unknown (0xff)",
	}

	for code, exp := range tests {
		if code.String() != exp {
			t.Errorf("Error on %d, got %v, expected %v",
				code, code, exp)
		}
	}
}

func TestCodeString(t *testing.T) {
	tests := map[COAPCode]string{
		0:             "Empty",
		GET:           "GET",
		POST:          "POST",
		NotAcceptable: "NotAcceptable",
		255:           "Unknown (0xff)",
	}

	for code, exp := range tests {
		if code.String() != exp {
			t.Errorf("Error on %d, got %v, expected %v", code, code, exp)
		}
	}
}

func TestEncodeMessageWithoutOptionsAndPayload(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}

	data := req.MustMarshalBinary()

	// Inspected by hand.
	exp := []byte{0x40, 0x1, 0x30, 0x39}
	if !bytes.Equal(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestEncodeMessageSmall(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}

	req.Options().Add(ETag, []byte("weetag"))
	req.Options().Add(MaxAge, 3)

	data := req.MustMarshalBinary()

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
		0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestEncodeMessageSmallWithPayload(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Payload:   []byte("hi"),
	}

	req.Options().Add(ETag, []byte("weetag"))
	req.Options().Add(MaxAge, 3)

	data := req.MustMarshalBinary()

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
		0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
		0xff, 'h', 'i',
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestInvalidMessageParsing(t *testing.T) {
	var invalidPackets = [][]byte{
		nil,
		{0x40},
		{0x40, 0},
		{0x40, 0, 0},
		{0xff, 0, 0, 0, 0, 0},
		{0x4f, 0, 0, 0, 0, 0},
		{0x45, 0, 0, 0, 0, 0},                // TKL=5 but packet is truncated
		{0x40, 0x01, 0x30, 0x39, 0x4d},       // Extended word length but no extra length byte
		{0x40, 0x01, 0x30, 0x39, 0x4e, 0x01}, // Extended word length but no full extra length word
	}

	for _, data := range invalidPackets {
		msg, err := ParseMessage(data)
		if err == nil {
			t.Errorf("Unexpected success parsing short message (%#v): %v", data, msg)
		}
	}
}

func TestMessageWithEmptyPayloadButMarker(t *testing.T) {
	_, err := ParseMessage([]byte{0x40, 0x01, 0xab, 0xcd,
		0xff, // Payload marker
	})
	expected := "Message format error: Payload marker (0xFF) followed by zero-length payload"
	if err.Error() != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, err.Error())
	}
}

func TestOptionsWithIllegalLengthAreIgnoredDuringParsing(t *testing.T) {
	exp := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 0xabcd,
		Payload:   []byte{0xef},
	}
	msg, err := ParseMessage([]byte{0x40, 0x01, 0xab, 0xcd,
		0x73, // URI-Port option (id 7) (uint) with length 3 (valid lengths are 0-2)
		0x11, 0x22, 0x33, 0xff, 0xdd})
	expected := "Critical option with invalid length found"
	if err.Error() != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, err.Error())
	}
	//if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", msg) {
	//	t.Errorf("Expected\n%#v\ngot\n%#v", exp, msg)
	//}

	msg, err = ParseMessage([]byte{0x40, 0x01, 0xab, 0xcd,
		0xd5, 0x01, // Max-Age option (uint) with length 5 (valid lengths are 0-4)
		0x11, 0x22, 0x33, 0x44, 0x55, 0xff, 0xef})
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", msg) {
		t.Errorf("Expected\n%#v\ngot\n%#v", exp, msg)
	}
}

func TestDecodeMessageWithoutOptionsAndPayload(t *testing.T) {
	input := []byte{0x40, 0x1, 0x30, 0x39}
	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type != Confirmable {
		t.Errorf("Expected message type confirmable, got %v", msg.Type)
	}
	if msg.Code != GET {
		t.Errorf("Expected message code GET, got %v", msg.Code)
	}
	if msg.MessageID != 12345 {
		t.Errorf("Expected message ID 12345, got %v", msg.MessageID)
	}
	if len(msg.Token) != 0 {
		t.Errorf("Incorrect token: %q", msg.Token)
	}
	if len(msg.Payload) != 0 {
		t.Errorf("Incorrect payload: %q", msg.Payload)
	}
}

func TestDecodeMessageSmallWithPayload(t *testing.T) {
	input := []byte{
		0x40, 0x1, 0x30, 0x39, 0x21, 0x3,
		0x26, 0x77, 0x65, 0x65, 0x74, 0x61, 0x67,
		0xff, 'h', 'i',
	}

	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type != Confirmable {
		t.Errorf("Expected message type confirmable, got %v", msg.Type)
	}
	if msg.Code != GET {
		t.Errorf("Expected message code GET, got %v", msg.Code)
	}
	if msg.MessageID != 12345 {
		t.Errorf("Expected message ID 12345, got %v", msg.MessageID)
	}

	if !bytes.Equal(msg.Payload, []byte("hi")) {
		t.Errorf("Incorrect payload: %q", msg.Payload)
	}
}

func TestEncodeMessageVerySmall(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}
	req.SetPathString("x")

	data := req.MustMarshalBinary()

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xb1, 0x78,
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

// Same as above, but with a leading slash
func TestEncodeMessageVerySmall2(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}
	req.SetPathString("/x")

	data := req.MustMarshalBinary()

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xb1, 0x78,
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestEncodeSeveral(t *testing.T) {
	tests := map[string][]string{
		"a":   []string{"a"},
		"axe": []string{"axe"},
		"a/b/c/d/e/f/h/g/i/j": []string{"a", "b", "c", "d", "e",
			"f", "h", "g", "i", "j"},
	}
	for p, a := range tests {
		m := &Message{Type: Confirmable, Code: GET, MessageID: 12345}
		m.SetPathString(p)
		b := m.MustMarshalBinary()

		m2, err := ParseMessage(b)
		if err != nil {
			t.Fatalf("Can't parse my own message at %#v: %v", p, err)
		}

		if !reflect.DeepEqual(m2.Path(), a) {
			t.Errorf("Expected %#v, got %#v", a, m2.Path())
			t.Fail()
		}
	}
}

func TestPathAsOption(t *testing.T) {
	m := &Message{Type: Confirmable, Code: GET, MessageID: 12345}
	m.Options().Set(LocationPath, "a")
	m.Options().Add(LocationPath, "b")
	got := m.MustMarshalBinary()

	exp := []byte{0x40, 0x1, 0x30, 0x39, 0x81, 0x61, 0x1, 0x62}
	if !bytes.Equal(got, exp) {
		t.Errorf("Got \n%#v\nwanted \n%#v", got, exp)
	}
}

func TestEncodePath14(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}
	req.SetPathString("123456789ABCDE")

	data := req.MustMarshalBinary()

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xbd, 0x01, // extended option length
		'1', '2', '3', '4', '5', '6', '7', '8',
		'9', 'A', 'B', 'C', 'D', 'E',
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestEncodePath15(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}
	req.SetPathString("123456789ABCDEF")

	data := req.MustMarshalBinary()

	// Inspected by hand.
	exp := []byte{
		0x40, 0x1, 0x30, 0x39, 0xbd, 0x02, // extended option length
		'1', '2', '3', '4', '5', '6', '7', '8',
		'9', 'A', 'B', 'C', 'D', 'E', 'F',
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestEncodeLargePath(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
	}
	req.SetPathString("this_path_is_longer_than_fifteen_bytes")

	if req.PathString() != "this_path_is_longer_than_fifteen_bytes" {
		t.Fatalf("Didn't get back the same path I posted: %v",
			req.PathString())
	}

	data := req.MustMarshalBinary()

	// Inspected by hand.
	exp := []byte{
		// extended length           0x19 + 13 = 38
		0x40, 0x1, 0x30, 0x39, 0xbd, 0x19, 0x74, 0x68, 0x69,
		0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x5f, 0x69, 0x73,
		0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x65, 0x72, 0x5f, 0x74,
		0x68, 0x61, 0x6e, 0x5f, 0x66, 0x69, 0x66, 0x74, 0x65,
		0x65, 0x6e, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73,
	}
	if !reflect.DeepEqual(exp, data) {
		t.Fatalf("Expected\n%#v\ngot\n%#v", exp, data)
	}
}

func TestDecodeLargePath(t *testing.T) {
	data := []byte{
		0x40, 0x1, 0x30, 0x39, 0xbd, 0x19, 0x74, 0x68,
		0x69, 0x73, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x5f, 0x69, 0x73,
		0x5f, 0x6c, 0x6f, 0x6e, 0x67, 0x65, 0x72, 0x5f, 0x74, 0x68,
		0x61, 0x6e, 0x5f, 0x66, 0x69, 0x66, 0x74, 0x65, 0x65, 0x6e,
		0x5f, 0x62, 0x79, 0x74, 0x65, 0x73,
	}

	req, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("Error parsing request: %v", err)
	}

	path := "this_path_is_longer_than_fifteen_bytes"

	exp := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Payload:   []byte{},
	}

	exp.Options().Set(URIPath, path)

	if fmt.Sprintf("%#v", exp) != fmt.Sprintf("%#v", req) {
		b := exp.MustMarshalBinary()
		t.Fatalf("Expected\n%#v\ngot\n%#v\nfor %#v", exp, req, b)
	}
}

func TestDecodeMessageSmaller(t *testing.T) {
	data := []byte{
		0x40, 0x1, 0x30, 0x39, 0x46, 0x77,
		0x65, 0x65, 0x74, 0x61, 0x67, 0xa1, 0x3,
	}

	req, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("Error parsing request: %v", err)
	}

	exp := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Payload:   []byte{},
	}

	exp.Options().Set(ETag, []byte("weetag"))
	exp.Options().Set(MaxAge, uint32(3))

	expected := fmt.Sprintf("%#v", exp)
	actual := fmt.Sprintf("%#v", req)
	if expected != actual {
		t.Fatalf("Expected\n%s\ngot\n%s", expected, actual)
	}
}

func TestByteEncoding(t *testing.T) {
	tests := []struct {
		Value    uint32
		Expected []byte
	}{
		{0, nil},
		{13, []byte{13}},
		{1024, []byte{4, 0}},
		{984284, []byte{0x0f, 0x04, 0xdc}},
		{823958824, []byte{0x31, 0x1c, 0x9d, 0x28}},
	}

	for _, v := range tests {
		got := encodeInt(v.Value)
		if !reflect.DeepEqual(got, v.Expected) {
			t.Fatalf("Expected %#v, got %#v for %v",
				v.Expected, got, v.Value)
		}
	}
}

func TestByteDecoding(t *testing.T) {
	tests := []struct {
		Value uint32
		Bytes []byte
	}{
		{0, nil},
		{0, []byte{0}},
		{0, []byte{0, 0}},
		{0, []byte{0, 0, 0}},
		{0, []byte{0, 0, 0, 0}},
		{13, []byte{13}},
		{13, []byte{0, 13}},
		{13, []byte{0, 0, 13}},
		{13, []byte{0, 0, 0, 13}},
		{1024, []byte{4, 0}},
		{1024, []byte{4, 0}},
		{1024, []byte{0, 4, 0}},
		{1024, []byte{0, 0, 4, 0}},
		{984284, []byte{0x0f, 0x04, 0xdc}},
		{984284, []byte{0, 0x0f, 0x04, 0xdc}},
		{823958824, []byte{0x31, 0x1c, 0x9d, 0x28}},
	}

	for _, v := range tests {
		got := decodeInt(v.Bytes)
		if v.Value != got {
			t.Fatalf("Expected %v, got %v for %#v",
				v.Value, got, v.Bytes)
		}
	}
}

/*
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   | 1 | 0 |   0   |     GET=1     |          MID=0x7d34           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |  11   |  11   |      "temperature" (11 B) ...                 |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func TestExample1(t *testing.T) {
	input := append([]byte{0x40, 1, 0x7d, 0x34,
		(11 << 4) | 11}, []byte("temperature")...)

	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type != Confirmable {
		t.Errorf("Expected message type confirmable, got %v", msg.Type)
	}
	if msg.Code != GET {
		t.Errorf("Expected message code GET, got %v", msg.Code)
	}
	if msg.MessageID != 0x7d34 {
		t.Errorf("Expected message ID 0x7d34, got 0x%x", msg.MessageID)
	}

	if msg.Options().Get(URIPath).AsString() != "temperature" {
		t.Errorf("Incorrect uri path: %q", msg.Options().Get(URIPath).AsString())
	}

	if len(msg.Token) > 0 {
		t.Errorf("Incorrect token: %x", msg.Token)
	}
	if len(msg.Payload) > 0 {
		t.Errorf("Incorrect payload: %q", msg.Payload)
	}
}

/*
    0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   | 1 | 2 |   0   |    2.05=69    |          MID=0x7d34           |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |1 1 1 1 1 1 1 1|      "22.3 C" (6 B) ...
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func TestExample1Res(t *testing.T) {
	input := append([]byte{0x60, 69, 0x7d, 0x34, 0xff},
		[]byte("22.3 C")...)

	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if msg.Type != Acknowledgement {
		t.Errorf("Expected message type confirmable, got %v", msg.Type)
	}
	if msg.Code != Content {
		t.Errorf("Expected message code Content, got %v", msg.Code)
	}
	if msg.MessageID != 0x7d34 {
		t.Errorf("Expected message ID 0x7d34, got 0x%x", msg.MessageID)
	}

	if len(msg.Token) > 0 {
		t.Errorf("Incorrect token: %x", msg.Token)
	}
	if !bytes.Equal(msg.Payload, []byte("22.3 C")) {
		t.Errorf("Incorrect payload: %q", msg.Payload)
	}
}

func TestIssue15(t *testing.T) {

	input := []byte{0x53, 0x2, 0x7a,
		0x23, 0x1, 0x2, 0x3, 0xb1, 0x45, 0xd, 0xd, 0x73, 0x70, 0x61,
		0x72, 0x6b, 0x2f, 0x63, 0x63, 0x33, 0x30, 0x30, 0x30, 0x2d,
		0x70, 0x61, 0x74, 0x63, 0x68, 0x2d, 0x76, 0x65, 0x72, 0x73,
		0x69, 0x6f, 0x6e, 0xff, 0x31, 0x2e, 0x32, 0x38}
	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatalf("Error parsing message: %v", err)
	}

	if !bytes.Equal(msg.Token, []byte{1, 2, 3}) {
		t.Errorf("Expected token = [1, 2, 3], got %v", msg.Token)
	}

	if !bytes.Equal(msg.Payload, []byte{0x31, 0x2e, 0x32, 0x38}) {
		t.Errorf("Expected payload = {0x31, 0x2e, 0x32, 0x38}, got %v", msg.Payload)
	}

	pathExp := "E/spark/cc3000-patch-version"
	if got := msg.PathString(); got != pathExp {
		t.Errorf("Expected path %q, got %q", pathExp, got)
	}
}

func TestErrorOptionMarker(t *testing.T) {
	input := []byte{0x53, 0x2, 0x7a, 0x23,
		0x1, 0x2, 0x3, 0xbf, 0x01, 0x02, 0x03, 0x04, 0x05, 0x6, 0x7, 0x8, 0x9,
		0xa, 0xb, 0xc, 0xe, 0xf, 0x10}
	msg, err := ParseMessage(input)
	if err == nil {
		t.Errorf("Unexpected success parsing malformed option: %v", msg)
	}
}

func TestDecodeContentFormatOptionToMediaType(t *testing.T) {
	data := []byte{
		0x40, 0x1, 0x30, 0x39, 0xc1, 0x32, 0x51, 0x29,
	}

	parsedMsg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("Error parsing request: %v", err)
	}

	// NOTE: We do NOT treat content type special anymore. It's the user who must convert types
	// We could offer utils for that.
	//expected := "coapmsg.MediaType"
	expected := "coapmsg.OptionValue"
	actualContentFormatType := fmt.Sprintf("%T", parsedMsg.Options().Get(ContentFormat))
	if expected != actualContentFormatType {
		t.Fatalf("Expected %#v got %#v", expected, actualContentFormatType)
	}
	actualAcceptType := fmt.Sprintf("%T", parsedMsg.Options().Get(Accept))
	if expected != actualAcceptType {
		t.Fatalf("Expected %#v got %#v", expected, actualAcceptType)
	}
}

func TestEncodeMessageWithAllOptions(t *testing.T) {
	req := Message{
		Type:      Confirmable,
		Code:      GET,
		MessageID: 12345,
		Token:     []byte("TOKEN"),
		Payload:   []byte("PAYLOAD"),
	}

	req.Options().Add(IfMatch, []byte("IFMATCH"))
	req.Options().Add(URIHost, "URIHOST")
	req.Options().Add(ETag, []byte("ETAG"))
	req.Options().Add(IfNoneMatch, []byte{})
	req.Options().Add(Observe, uint32(9999))
	req.Options().Add(URIPort, uint32(5683))
	req.Options().Add(LocationPath, "LOCATIONPATH")
	req.Options().Add(URIPath, "URIPATH")
	req.Options().Add(ContentFormat, TextPlain)
	req.Options().Add(MaxAge, uint32(9999))
	req.Options().Add(URIQuery, "URIQUERY")
	req.Options().Add(Accept, TextPlain)
	req.Options().Add(LocationQuery, "LOCATIONQUERY")
	req.Options().Add(ProxyURI, "PROXYURI")
	req.Options().Add(ProxyScheme, "PROXYSCHEME")
	req.Options().Add(Size1, uint32(9999))

	data := req.MustMarshalBinary()

	parsedMsg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("Error parsing binary packet: %v", err)
	}
	assertEqualMessages(t, req, parsedMsg)
}
