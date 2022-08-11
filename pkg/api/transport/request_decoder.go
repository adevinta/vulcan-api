/*
Copyright 2021 Adevinta
*/

package transport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"

	"github.com/adevinta/errors"
)

const tagName = "urlvar"
const headerTagName = "headervar"
const queryStringTagName = "urlquery"

func makeDecodeRequestFunc(req interface{}) kithttp.DecodeRequestFunc {
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		return setRequestStructFields(req, r)
	}
}

// setRequestStructFields will take a transport layer request struct and set
// the fields using parameters taken from the request route path. For example:
//
// Given the request struct for finding teams by a user_id
//
//	type FindTeamsByUserJSONRequest struct {
//		 UserID string `json:"user_id" urlvar:"user_id"`
//	}
//
// The UserID will be loaded from the request route, in this case:
// /v1/users/{user_id}/teams
//
// The link between the request struct and the route path is a custom tag on the
// struct field, indicating which path paremeter corresponds to this field.
func setRequestStructFields(req interface{}, r *http.Request) (interface{}, error) {
	requestType := reflect.TypeOf(req)
	requestObject := reflect.New(requestType).Interface()
	if r.ContentLength > 0 {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			if e := json.NewDecoder(r.Body).Decode(requestObject); e != nil {
				//TODO: log internal error
				log.Printf("e: %v", e)
				// If the error is already wrapped  don't wrap it again.
				if eStack, ok := e.(*errors.ErrorStack); ok {
					return nil, eStack
				}
				return nil, errors.Assertion("cannot unmarshal " + requestType.Name())
			}
		}
	}
	requestObject = loadParametersFromQueryString(requestObject, r.URL.Query())
	requestObject = loadParametersFromRequestPath(requestObject, mux.Vars(r))
	requestObject = loadParametersFromRequestHeaders(requestObject, r.Header)
	return requestObject, nil
}

func loadParametersFromQueryString(requestObject interface{}, v url.Values) interface{} {
	t := reflect.TypeOf(requestObject)
	obj := t.Elem()
	for i := 0; i < obj.NumField(); i++ {
		// If the field is a struct recurse to initialize the possible value of the fields
		// of that struct.
		field := obj.Field(i)
		if field.Type.Kind() == reflect.Struct {
			val := reflect.ValueOf(requestObject).Elem().Field(i)
			// This should not happen never, but still...
			if !val.CanAddr() {
				continue
			}
			loadParametersFromQueryString(val.Addr().Interface(), v)
			continue
		}
		tag := field.Tag.Get(queryStringTagName)

		// Skip if tag is not defined or ignored
		if tag == "" || tag == "-" {
			continue
		}
		tagValue := v.Get(tag)
		if tagValue != "" {
			switch reflect.ValueOf(requestObject).Elem().Field(i).Kind() {
			case reflect.String:
				reflect.ValueOf(requestObject).Elem().Field(i).SetString(tagValue)
			case reflect.Int:
				intValue, err := strconv.Atoi(tagValue)
				if err != nil {
					continue
				}
				reflect.ValueOf(requestObject).Elem().Field(i).SetInt(int64(intValue))
			case reflect.Float32:
				floatValue, err := strconv.ParseFloat(tagValue, 32)
				if err != nil {
					continue
				}
				reflect.ValueOf(requestObject).Elem().Field(i).SetFloat(floatValue)
			case reflect.Float64:
				floatValue, err := strconv.ParseFloat(tagValue, 64)
				if err != nil {
					continue
				}
				reflect.ValueOf(requestObject).Elem().Field(i).SetFloat(floatValue)
			}
		}
	}

	return requestObject
}

func loadParametersFromRequestPath(requestObject interface{}, vars map[string]string) interface{} {
	obj := reflect.TypeOf(requestObject).Elem()

	for i := 0; i < obj.NumField(); i++ {
		// If the field is a struct recurse to initialize the possible value of the fields
		// of that struct.
		field := obj.Field(i)
		if field.Type.Kind() == reflect.Struct {
			val := reflect.ValueOf(requestObject).Elem().Field(i)
			// This should not happen never, but still...
			if !val.CanAddr() {
				continue
			}
			loadParametersFromRequestPath(val.Addr().Interface(), vars)
			continue
		}
		tag := field.Tag.Get(tagName)

		// Skip if tag is not defined or ignored
		if tag == "" || tag == "-" {
			continue
		}
		if vars[tag] != "" {
			reflect.ValueOf(requestObject).Elem().Field(i).SetString(vars[tag])
		}
	}

	return requestObject
}

func loadParametersFromRequestHeaders(requestObject interface{}, headers http.Header) interface{} {
	obj := reflect.TypeOf(requestObject).Elem()

	for i := 0; i < obj.NumField(); i++ {
		field := obj.Field(i)
		if field.Type.Kind() == reflect.Struct {
			val := reflect.ValueOf(requestObject).Elem().Field(i)
			// This should not happen never, but still...
			if !val.CanAddr() {
				continue
			}
			loadParametersFromRequestHeaders(val.Addr().Interface(), headers)
			continue
		}
		tag := field.Tag.Get(headerTagName)

		// Skip if tag is not defined or ignored
		if tag == "" || tag == "-" {
			continue
		}

		if len(headers[tag]) > 0 {
			reflect.ValueOf(requestObject).Elem().Field(i).SetString(headers[tag][0])
		}
	}

	return requestObject
}
