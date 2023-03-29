package new_db

import (
	"bytes"
	"encoding/json"
	"strings"
)

type StringArray []string

func (s *StringArray) FromDB(bts []byte) error {
	if len(bts) == 0 {
		return nil
	}

	str := string(bts)
	str = strings.TrimPrefix(str, "{")
	str = strings.TrimSuffix(str, "}")
	ia := strings.Split(str, ",")

	*s = ia
	return nil
}

func (s *StringArray) ToDB() ([]byte, error) {
	return serializeArray(*s, "{", "}"), nil
}

func (arr StringArray) MarshalJSON() ([]byte, error) {
	return serializeArrayAsString(arr, "[", "]"), nil
}

func (arr *StringArray) UnmarshalJSON(b []byte) error {
	var strarr []string

	err := json.Unmarshal(b, &strarr)
	if err != nil {
		return err
	}

	*arr = strarr
	return nil
}
func serializeArray(s []string, prefix string, suffix string) []byte {
	var buffer bytes.Buffer

	buffer.WriteString(prefix)

	for idx, val := range s {
		if idx > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(val)
	}

	buffer.WriteString(suffix)

	return buffer.Bytes()
}

func serializeArrayAsString(s []string, prefix string, suffix string) []byte {
	var buffer bytes.Buffer

	buffer.WriteString(prefix)

	for idx, val := range s {
		if idx > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString("\"")
		buffer.WriteString(val)
		buffer.WriteString("\"")
	}
	buffer.WriteString(suffix)
	return buffer.Bytes()
}
