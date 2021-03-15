/*
Copyright 2021 Adevinta
*/

package common

import (
	"encoding/json"
)

func IsValidJSON(str *string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(*str), &js) == nil
}

func IsStringEmpty(str *string) bool {
	if str == nil {
		return true
	}
	if StringValue(str) == "" {
		return true
	}
	return false
}

// TODO: drop these in favor of aws's helpers?
// We already have aws as a dependency
func Bool(b bool) *bool       { return &b }
func String(s string) *string { return &s }
func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}
