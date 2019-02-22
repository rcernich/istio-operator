package main

import (
	"encoding/json"
	"fmt"
)

type MapType map[string]string

type EmbeddedField struct {
	Enabled *bool  `json:"enabled,omitempty"`
	FieldB  string `json:"fieldB,omitempty"`
}

func (ef EmbeddedField) String() string {
    var enabled bool
    if ef.Enabled == nil {
        enabled = false
    } else {
        enabled = *ef.Enabled
    }
	return fmt.Sprintf("EmbeddedField{Enabled=%t,FieldB=%s}", enabled, ef.FieldB)
}

// type CustomMapType struct {
// 	EmbeddedField `json:",inline"`
// 	MapType       `json:",inline"`
// }

type ChildType struct {
	FieldC string `json:"fieldC,omitempty"`
}

func (ct ChildType) String() string {
	return fmt.Sprintf("ChildType{FieldC=%s}", ct.FieldC)
}

type CustomMapType struct {
	EmbeddedField `json:",inline"`
	MapType       map[string]ChildType `json:",inline"`
}

func (cmt CustomMapType) String() string {
	return fmt.Sprintf("CustomMapType{%v,%v}", cmt.EmbeddedField, cmt.MapType)
}

func (cmt CustomMapType) MarshalJSON() ([]byte, error) {
	var data []byte
	embeddedData, err := json.Marshal(cmt.EmbeddedField)
	if err != nil {
		return []byte{}, err
	}
	if cmt.MapType != nil && len(cmt.MapType) > 0 {
		var err error
		data, err = json.Marshal(cmt.MapType)
		if err != nil {
			return []byte{}, err
		}
		data = append(append(embeddedData[:len(embeddedData)-1], byte(',')), data[1:]...)
	} else {
		data = embeddedData
	}
	return data, nil
}
func (cmt *CustomMapType) UnmarshalJSON(data []byte) error {
    foo := map[string]json.RawMessage{}
    err := json.Unmarshal(data, &foo)
    if err != nil {
        return err
    }
    if value, ok := foo["enabled"]; ok {
        err = json.Unmarshal(value, &cmt.Enabled)
        if err != nil {
            return err
        }
        delete(foo, "enabled")
    }
    if value, ok := foo["fieldB"]; ok {
        err = json.Unmarshal(value, &cmt.FieldB)
        if err != nil {
            return err
        }
        delete(foo, "fieldB")
    }
    cmt.MapType = map[string]ChildType{}
    for key, value := range foo {
        ct := ChildType{}
        err = json.Unmarshal(value, &ct)
        if err != nil {
            return err
        }
        cmt.MapType[key] = ct
    }
	return nil
}

type EmbeddedField2 struct {
	EmbeddedField `json:",inline"`
	FieldJ string
	FieldK string
}

type OuterType struct {
	EmbeddedField2 `json:",inline"`
}

func main() {
	trueVal := true
	ot := OuterType{
		EmbeddedField2: EmbeddedField2{EmbeddedField: EmbeddedField{Enabled: &trueVal, FieldB: "foo"}, FieldJ: "bar", FieldK: "baz"},
	}
	data, err := json.Marshal(ot)
	if err == nil {
		fmt.Printf("%s\n", data)
	} else {
		fmt.Printf("error: %v\n", err)
	}
	cmt := CustomMapType{}
	data, err = json.Marshal(cmt)
	if err == nil {
		fmt.Printf("%s\n", data)
	} else {
		fmt.Printf("error: %v\n", err)
	}
	cmt = CustomMapType{
		EmbeddedField: EmbeddedField{Enabled: &trueVal, FieldB: "boohoo"},
		MapType:       map[string]ChildType{"foo": ChildType{FieldC: "bar"}, "bar": ChildType{FieldC: "baz"}},
	}
	for key, value := range cmt.MapType {
		fmt.Printf("%s = %v\n", key, value)
	}
	data, err = json.Marshal(cmt)
	if err == nil {
		fmt.Printf("%s\n", data)
	} else {
		fmt.Printf("error: %v\n", err)
	}
	err = json.Unmarshal(data, &cmt)
	if err == nil {
		fmt.Printf("unmarshalled:\n%v", cmt)
	} else {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Println("Hello, playground")
}
