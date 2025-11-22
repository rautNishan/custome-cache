package protocol

import (
	"testing"
)

func TestDecode(t *testing.T) {
	inputes := map[string]interface{}{
		"+SET\r\n":      "SET",
		"+GET\r\n":      "GET",
		"HEHE":          nil,
		"+HEHE\r\n":     "HEHE",
		"+HEHE\r\n\r\n": "HEHE",
		"+GET":          nil, //Does not know when the ending is
	}

	for k, v := range inputes {
		val, err := Decode([]byte(k))
		if err != nil {
			t.Fatalf("unexpected error for input '%s': %v", k, err)
		}

		if val != v {
			t.Fatalf("for input '%s' expected '%s', got '%v'", k, v, val)
		}
	}
}

func TestDecodeOne(t *testing.T) {
	inputes := map[string]interface{}{
		"+SET\r\n": "SET",
		"+GET\r\n": "GET",
		"-SET\r\n": "SET",
	}

	for k, v := range inputes {
		val, n, err := decodeOne([]byte(k))
		if err != nil {
			t.Fatalf("unexpected error for input '%s': %v", k, err)
		}

		if s, ok := val.(string); ok && len(s)+3 != n {
			t.Fatalf("Bytes read are not correct for %s, %s", val, k)
		}

		if val != v {
			t.Fatalf("for input '%s' expected '%s', got '%v'", k, v, val)
		}
	}
}
