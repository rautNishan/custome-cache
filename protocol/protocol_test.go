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
		"+GET":          nil,
	}

	for k, v := range inputes {
		val, _, err := Decode([]byte(k))
		if err != nil {
			t.Fatalf("unexpected error for input '%s': %v", k, err)
		}

		if val != v {
			t.Fatalf("for input '%s' expected '%s', got '%v'", k, v, val)
		}
	}
}

// func TestDecodeOne(t *testing.T) {
// 	inputes := map[string]interface{}{
// 		"+SET\r\n": "SET",
// 		"+GET\r\n": "GET",
// 		"-SET\r\n": "SET",
// 	}

// 	for k, v := range inputes {
// 		val, n, err := decodeOne([]byte(k))
// 		if err != nil {
// 			t.Fatalf("unexpected error for input '%s': %v", k, err)
// 		}

// 		if s, ok := val.(string); ok && len(s)+3 != n {
// 			t.Fatalf("Bytes read are not correct for %s, %s", val, k)
// 		}

// 		if val != v {
// 			t.Fatalf("for input '%s' expected '%s', got '%v'", k, v, val)
// 		}
// 	}
// }

func TestDecodeArray(t *testing.T) {
	input := []byte(
		"*3\r\n" +
			"$3\r\nSET\r\n" +
			"$3\r\nkey\r\n" +
			"$5\r\nvalue\r\n",
	)

	val, _, err := Decode(input)
	if err != nil {
		t.Fatal(err)
	}

	arr := val.([]interface{})
	if arr[0].(string) != "SET" {
		t.Fatalf("expected SET, got %v", arr[0])
	}
	if arr[1].(string) != "key" {
		t.Fatalf("expected key, got %v", arr[1])
	}
	if arr[2].(string) != "value" {
		t.Fatalf("expected value, got %v", arr[2])
	}
}
