package domain

import (
	"encoding/json"
	"testing"
)

type testJSONStruct struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func TestJSONObjectUnmarshal(t *testing.T) {
	var params testJSONStruct
	err := JSONObject("{}").Unmarshal(&params)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("params=%+v", params)
	err = JSONObject("{\"foo\":\"baz\",\"bar\":42}").Unmarshal(&params)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("params=%+v", params)
}

func TestJSONObjectUnmarshalJSON(t *testing.T) {
	ts := &testJSONStruct{
		Foo: "mmm",
		Bar: 33,
	}
	bytes, err := json.Marshal(ts)
	if err != nil {
		t.Fatal(err)
	}
	js := JSONObject("{\"foo\":\"baz\",\"bar\":42}")
	err = js.UnmarshalJSON(bytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("js=%+v", string(js))
}
