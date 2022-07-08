package utils

import (
	"bytes"
	"encoding/json"
)

// JsonEncodingCompact appends to dst the JSON-encoded src with insignificant space characters elided
func JsonEncodingCompact(src []byte) ([]byte, error) {
	compactedBuffer := new(bytes.Buffer)
	err := json.Compact(compactedBuffer, src)
	if err != nil {
		return compactedBuffer.Bytes(), err
	}
	return compactedBuffer.Bytes(), nil
}

// JsonEncoding parses and return the the JSON-encoded data
func JsonEncoding(src []byte) ([]byte, error) {
	var i interface{}
	err := json.Unmarshal(src, &i)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return body, nil
}
