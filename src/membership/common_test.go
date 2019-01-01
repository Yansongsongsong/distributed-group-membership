package membership

import (
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
