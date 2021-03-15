/*
Copyright 2021 Adevinta
*/

package jwt

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	type expect struct {
		hasErr bool
		token  string
	}
	data := map[string]interface{}{}
	testCases := []struct {
		given  string
		key    string
		data   map[string]interface{}
		expect expect
	}{
		{
			given: "Empty key",
			data:  data,
			expect: expect{
				hasErr: true,
				token:  "",
			},
		},
		{
			given: "Good key",
			key:   `12567567891234556781234556789000`,
			data:  data,
			expect: expect{
				hasErr: false,
				token:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.lWw0MFb_yuJJt5Hu1-R5YSYvkBPbQDAzETdBEskdt28",
			},
		},
		{
			given: "Good key with content",
			key:   `12567567891234556781234556789000`,
			data: map[string]interface{}{
				"foo": "bar",
			},
			expect: expect{
				hasErr: false,
				token:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIifQ.rmfyCHpsDxCG7FNL54-titqLr8eGi0v-U89LDx4oQNg",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.given, func(t *testing.T) {
			jwtx := NewJWTConfig(tt.key)
			token, err := jwtx.GenerateToken(tt.data)
			got := expect{
				hasErr: err != nil,
				token:  token,
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("given %s expected %v but got %v", tt.given, tt.expect, got)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	jwtx := NewJWTConfig("secret")
	t.Run("expired token", func(t *testing.T) {
		tokenGenTime, _ := time.Parse(time.RFC3339, "2016-02-05T16:45:04.398453671+01:00")
		tokenExpTime, _ := time.Parse(time.RFC3339, "2017-02-05T16:45:04.398453671+01:00")
		token, err := jwtx.GenerateToken(map[string]interface{}{
			"first_name": "firstname",
			"last_name":  "lastname",
			"email":      "email",
			"username":   "username",
			"iat":        tokenGenTime.Unix(),
			"exp":        tokenExpTime.Unix(),
			"sub":        "email",
		})
		if err != nil {
			t.Fatalf("expected no error generating token but got: %v", err)
		}

		if err := jwtx.ValidateToken(token); !errors.Is(err, ErrTokenInvalid) {
			t.Fatalf("expected error validating token to be %v but got: %v",
				ErrTokenInvalid, err)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		tokenGenTime, _ := time.Parse(time.RFC3339, "2016-02-05T16:45:04.398453671+01:00")
		tokenExpTime, _ := time.Parse(time.RFC3339, "2027-02-05T16:45:04.398453671+01:00")
		token, err := jwtx.GenerateToken(map[string]interface{}{
			"first_name": "firstname",
			"last_name":  "lastname",
			"email":      "email",
			"username":   "username",
			"iat":        tokenGenTime.Unix(),
			"exp":        tokenExpTime.Unix(),
			"sub":        "email",
		})
		if err != nil {
			t.Fatalf("expected no error generating token but got: %v", err)
		}

		if err := jwtx.ValidateToken(token); err != nil {
			t.Fatalf("expected no error validating token but got: %v", err)
		}
	})
}
