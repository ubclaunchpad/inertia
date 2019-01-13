package util

import (
	"reflect"
)

// SetProperty takes a struct pointer and searches for its "toml" tag with a search key
// and set property value with the tag
func SetProperty(name string, value string, structObject interface{}) bool {
	var val = reflect.ValueOf(structObject)
	if val.Kind() != reflect.Ptr {
		return false
	}
	var structVal = val.Elem()
	for i := 0; i < structVal.NumField(); i++ {
		valueField := structVal.Field(i)
		typeField := structVal.Type().Field(i)
		if typeField.Tag.Get("toml") == name {
			if valueField.IsValid() && valueField.CanSet() && valueField.Kind() == reflect.String {
				valueField.SetString(value)
				return true
			}
		}
	}
	return false
}
