package util

import (
	"encoding/json"
)

func JsonMarshal(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	return B2S(bytes)
}
