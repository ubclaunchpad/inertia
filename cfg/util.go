package cfg

import (
	"fmt"
	"reflect"
	"strings"
)

// SetProperty takes a struct pointer and searches for its "toml" tag with a search key
// and set property value with the tag
func SetProperty(name string, value string, obj interface{}) error {
	var ref = reflect.ValueOf(obj)
	if ref.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot set '%s': %s (%+v) is not a pointer",
			name, ref.Kind().String(), ref)
	}
	if ref.IsNil() {
		return fmt.Errorf("cannot set '%s': %s (%+v) is nil",
			name, ref.Kind().String(), ref)
	}
	var val = ref.Elem()

	// query for property within TOML tags
	var parts = strings.Split(name, ".")
	for i := 0; i < val.NumField(); i++ {
		var typeVal = val.Type().Field(i)
		var fieldVal = val.Field(i)
		if typeVal.Tag.Get("toml") == parts[0] {
			if len(parts) > 1 {
				// recurse on nested property
				var fieldPtr = fieldVal.Elem().Addr()
				return SetProperty(strings.Join(parts[1:], "."), value, fieldPtr.Interface())
			}
			if fieldVal.IsValid() && fieldVal.CanSet() && fieldVal.Kind() == reflect.String {
				// attempt to set string
				fieldVal.SetString(value)
				return nil
			}
		}
	}

	return fmt.Errorf("could not set property '%s' on %s %+v",
		name, ref.Kind().String(), val)
}
