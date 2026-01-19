package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// Reference (https://redis.io/docs/latest/develop/reference/protocol-spec/)
const CLRF = "\r\n"

func Decode(b []byte) (interface{}, int, error) {
	if len(b) == 0 {
		return nil, 0, nil
	}
	return decodeOne(b)
}

func decodeOne(b []byte) (interface{}, int, error) {
	if len(b) == 0 {
		return nil, 0, nil
	}

	switch b[0] {
	case '+':
		return readSimpleString(b)
	case '-':
		return readSimpleString(b)
	case ':':
		return readInteger(b)
	case '$':
		return readBulkString(b)
	case '*':
		return readArray(b)
	default:
		return nil, 0, errors.New("unknown RESP type")
	}
}

func readSimpleString(b []byte) (string, int, error) {
	i := bytes.Index(b, []byte(CLRF))
	if i == -1 {
		return "", 0, nil
	}
	return string(b[1:i]), i + 2, nil
}

func readInteger(b []byte) (int, int, error) {
	i := bytes.Index(b, []byte(CLRF))
	if i == -1 {
		return 0, 0, nil
	}
	val, err := strconv.Atoi(string(b[1:i]))
	if err != nil {
		return 0, 0, err
	}
	return val, i + 2, nil
}

func readBulkString(b []byte) (string, int, error) {
	i := bytes.Index(b, []byte(CLRF))
	if i == -1 {
		return "", 0, nil
	}

	length, err := strconv.Atoi(string(b[1:i]))
	if err != nil {
		return "", 0, err
	}

	if length == -1 {
		return "", i + 2, nil
	}

	start := i + 2
	end := start + length
	if len(b) < end+2 {
		return "", 0, nil
	}

	if b[end] != '\r' || b[end+1] != '\n' {
		return "", 0, errors.New("invalid bulk string termination")
	}

	return string(b[start:end]), end + 2, nil
}

func readArray(b []byte) ([]interface{}, int, error) {
	i := bytes.Index(b, []byte(CLRF))
	if i == -1 {
		return nil, 0, nil
	}
	count, err := strconv.Atoi(string(b[1:i]))
	if err != nil {
		return nil, 0, err
	}

	offset := i + 2
	arr := make([]interface{}, 0, count)

	for j := 0; j < count; j++ {
		val, read, err := decodeOne(b[offset:])
		if err != nil {
			return nil, 0, err
		}
		if read == 0 {
			return nil, 0, nil
		}
		arr = append(arr, val)
		offset += read
	}
	return arr, offset, nil
}

func Encode(v interface{}, simpleString bool) []byte {
	switch val := v.(type) {
	case string:
		if simpleString {
			return encodeSimpleString(val)
		}
		return encodeBulkString(val)
	case int:
		return encodeInt(val)
	case int64:
		return encodeInt64(val)
	case error:
		return []byte(fmt.Sprintf("-%s\r\n", val))
	}
	return encodeNil()
}

func encodeInt(val int) []byte {
	//First find how many bytes do i need
	b := getIntLen(val)
	buff := make([]byte, 0, b)
	buff = append(buff, ':')
	buff = strconv.AppendInt(buff, int64(val), 10)
	buff = append(buff, '\r', '\n')
	return buff
}

func encodeInt64(val int64) []byte {
	//First find how many bytes do i need
	b := getIntLen64(val)
	buff := make([]byte, 0, b)
	buff = append(buff, ':')
	buff = strconv.AppendInt(buff, int64(val), 10)
	buff = append(buff, '\r', '\n')
	return buff
}

func encodeSimpleString(val string) []byte {
	buff := make([]byte, 0, len(val)+3)
	buff = append(buff, '+')
	buff = append(buff, val...)
	buff = append(buff, '\r', '\n')
	return buff
}

func encodeBulkString(val string) []byte {
	if val == "" {
		return []byte("$0\r\n\r\n")
	}
	digits := 1
	valLen := len(val)
	digits += getIntLen(valLen)
	buf := make([]byte, 0, 1+digits+2+len(val)+2)
	buf = append(buf, '$')
	buf = strconv.AppendInt(buf, int64(len(val)), 10)
	buf = append(buf, '\r', '\n')
	buf = append(buf, val...)
	buf = append(buf, '\r', '\n')

	return buf
}

func encodeNil() []byte {
	buff := make([]byte, 0, 3)
	buff = append(buff, '_')
	buff = append(buff, '\r', '\n')
	return buff
}

func getIntLen(val int) int {
	totalLen := 0
	if val == 0 {
		return 1
	}

	if val < 0 {
		totalLen += 1 //for - sign
		val = -val    //make it positive
	}

	for {
		if val == 0 {
			break
		}
		val = val / 10
		totalLen += 1
	}
	return totalLen
}

func getIntLen64(val int64) int {
	totalLen := 0
	if val == 0 {
		return 1
	}

	if val < 0 {
		totalLen += 1 //for - sign
		val = -val    //make it positive
	}

	for {
		if val == 0 {
			break
		}
		val = val / 10
		totalLen += 1
	}
	return totalLen
}
