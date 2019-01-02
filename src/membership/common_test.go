package membership

import (
	"reflect"
	"testing"
)

func TestParseOneEndpoint(t *testing.T) {
	if addr, e := ParseOneEndpoint("asadasf"); e != nil {
		t.Log("yes: ", e)
	} else {
		t.Fatal("that should throw error, Addr is: ", addr)
	}

	if addr, e := ParseOneEndpoint("127.0.0.1:123:12"); e != nil {
		t.Log("yes: ", e)
	} else {
		t.Fatal("that should throw error, Addr is: ", addr)
	}

	if addr, e := ParseOneEndpoint("127.0.1.1.1:123"); e != nil {
		t.Log("yes: ", e)
	} else {
		t.Fatal("that should throw error, Addr is: ", addr)
	}

	if addr, e := ParseOneEndpoint("127.0.0.1:1233"); e == nil {
		t.Log("yes: ", addr)
	} else {
		t.Fatal("that should not throw error, error is: ", e)
	}

}

func TestEncodeMessage(t *testing.T) {
	msg := Message{"192.168.70.30:9981", "192.168.70.30:9981", "tar", 2019, Join}
	bf := EncodeMessage(msg)

	m := DecodeMessage(bf.Bytes())

	if !reflect.DeepEqual(*m, msg) {
		t.Fatalf("wrong! \nmsg: %#v \nm %#v", msg, m)
	}

}
