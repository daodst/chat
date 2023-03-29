package new_db

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

type Int64Array []int64

func (s *Int64Array) FromDB(bts []byte) error {
	if len(bts) == 0 {
		return nil
	}

	str := string(bts)
	if strings.HasPrefix(str, "{") {
		str = "[" + str[1:len(str)]
	}

	if strings.HasSuffix(str, "}") {
		str = str[0:len(str)-1] + "]"
	}

	var ia = &[]int64{}

	err := json.Unmarshal([]byte(str), ia)
	if err != nil {
		return err
	}

	*s = Int64Array(*ia)
	return nil
}

func (s *Int64Array) ToDB() ([]byte, error) {
	return serializeBigIntArray(*s, "{", "}"), nil
}

func (arr Int64Array) MarshalJSON() ([]byte, error) {
	return serializeBigIntArrayAsString(arr, "[", "]"), nil
}

func (arr *Int64Array) UnmarshalJSON(b []byte) error {
	var strarr []string
	var intarr []int64

	err := json.Unmarshal(b, &strarr)
	if err != nil {
		return err
	}

	for _, s := range strarr {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		intarr = append(intarr, i)
	}

	*arr = intarr
	return nil
}
func serializeBigIntArray(s []int64, prefix string, suffix string) []byte {
	var buffer bytes.Buffer

	buffer.WriteString(prefix)

	for idx, val := range s {
		if idx > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(strconv.FormatInt(val, 10))
	}

	buffer.WriteString(suffix)

	return buffer.Bytes()
}

func serializeBigIntArrayAsString(s []int64, prefix string, suffix string) []byte {
	var buffer bytes.Buffer

	buffer.WriteString(prefix)

	for idx, val := range s {
		if idx > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatInt(val, 10))
		buffer.WriteString("\"")
	}
	buffer.WriteString(suffix)
	return buffer.Bytes()
}
