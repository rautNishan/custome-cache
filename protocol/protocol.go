package protocol

import (
	"errors"
	"log"
)

// Reference (https://redis.io/docs/latest/develop/reference/protocol-spec/)

func Decode(b []byte) (interface{}, error) {
	if len(b) == 0 {
		return nil, errors.New("no data") //For the first time
	}

	value, totalRead, err := decodeOne(b)

	if err != nil {
		log.Println("Error while calling functionj decodeOne: ", err)
	}

	if totalRead == 0 {
		return nil, nil //Indacating \r\n has not yet been read
	}

	return value, nil
}

func decodeOne(b []byte) (interface{}, int, error) {
	if len(b) == 0 {
		return nil, 0, nil //No data yet
	}

	switch b[0] {
	case '+':
		return readSimpleString(b)
	default:
		return readSimpleString(b)
	}
}

func readSimpleString(b []byte) (string, int, error) {
	//Do not read first byte(+)
	for i := 1; i < len(b)-1; i++ {
		if b[i] == '\r' && b[i+1] == '\n' {
			value := string(b[1:i])
			return value, i + 2, nil
		}
	}
	return "", 0, nil
}
