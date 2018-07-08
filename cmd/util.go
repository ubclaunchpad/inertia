package cmd

import (
	"reflect"

	"github.com/spf13/cobra"
)

// deepCopy is a helper function for deeply copying a Cobra command.
func deepCopy(cmd *cobra.Command) *cobra.Command {
	newCmd := &cobra.Command{}
	*newCmd = *cmd
	return newCmd
}

// setProperty takes a struct pointer and searches for its "toml" tag with a search key
// and set property value with the tag
func setProperty(name string, value string, structObject interface{}) bool {
	val := reflect.ValueOf(structObject)

	if val.Kind() != reflect.Ptr {
		return false
	}
	structVal := val.Elem()
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
