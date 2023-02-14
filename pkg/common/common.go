/*
Copyright 2021 Adevinta
*/

package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/adevinta/errors"
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

func IsHttpStatusOk(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

func BuildQueryFilter(filters map[string]string) string {
	filterParts := []string{}
	for key, value := range filters {
		part := fmt.Sprintf("%s=%s", key, value)
		filterParts = append(filterParts, part)
	}
	return strings.Join(filterParts, "&")
}

func ParseHttpErr(statusCode int, mssg string) error {
	switch statusCode {
	case http.StatusBadRequest:
		return errors.Assertion(mssg)
	case http.StatusUnauthorized:
		return errors.Unauthorized(mssg)
	case http.StatusForbidden:
		return errors.Forbidden(mssg)
	case http.StatusNotFound:
		return errors.NotFound(mssg)
	case http.StatusMethodNotAllowed:
		return errors.MethodNotAllowed(mssg)
	case http.StatusUnprocessableEntity:
		return errors.Validation(mssg)
	default:
		return errors.Default(mssg)
	}
}
