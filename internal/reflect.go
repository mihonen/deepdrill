package internal

import (
	"reflect"
	"strings"
)

type FieldInfo struct {
	Name    string
	Type    string 
	Hint    string
	Options []string
}

func ParseStruct(target interface{}) []FieldInfo {
	var infos []FieldInfo
	val := reflect.ValueOf(target).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// We look for our custom "dd" (deepdrill) tag
		tag := field.Tag.Get("dd") 
		if tag == "" { continue }

		info := FieldInfo{Name: field.Name}
		// Simple parser for: "type:link,hint:some text"
		parts := strings.Split(tag, ",")
		for _, p := range parts {
			kv := strings.Split(p, ":")
			if len(kv) != 2 { continue }
			key, value := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			
			switch key {
			case "type": info.Type = value
			case "hint": info.Hint = value
			}
		}
		infos = append(infos, info)
	}
	return infos
}


