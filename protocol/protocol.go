package protocol

// Reference (https://redis.io/docs/latest/develop/reference/protocol-spec/)
func DecodeOne(b []byte) (interface{}, int, error) {
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
