/*
Copyright 2021 Adevinta
*/

package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/adevinta/errors"
)

// ROLFP stores the vector containing the dimensions we use to classify the
// impact of an asset.
type ROLFP struct {
	Reputation byte
	Operation  byte
	Legal      byte
	Financial  byte
	Personal   byte
	Scope      byte
	IsEmpty    bool
}

// String returns the representation of the ROLFP in the form:
// R:0/O:0/L:0/F:0/P:0+S:0
func (r ROLFP) String() string {
	if r.IsEmpty {
		return ""
	}
	return fmt.Sprintf("R:%d/O:%d/L:%d/F:%d/P:%d+S:%d", r.Reputation, r.Operation, r.Legal, r.Financial, r.Personal, r.Scope)
}

// MarshalJSON marshals a ROLFP to JSON.
func (r ROLFP) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// Level return the corresponding level of the ROLFP according to the following rules:
// Level 0: Accounts meeting none of the ROLFP criteria.
// Level 1: Accounts meeting 1 or 2 of the ROLFP criteria.
// Level 2: Accounts either:
//   Meeting 3 or more of the ROLFP criteria.
//   With unknown scope, that is scope 2
// If the rolfp is empty the level will be 2.
func (r ROLFP) Level() byte {
	if r.IsEmpty || r.Scope == 2 {
		return 2
	}
	var criteria byte
	criteria = r.Reputation + r.Operation + r.Legal + r.Financial + r.Personal
	if criteria == 0 {
		return 0
	}
	if criteria == 1 || criteria == 2 {
		return 1
	}
	return 2
}

// UnmarshalJSON unmarshals a ROLFP encoded in the form:
// R:0/O:0/L:0/F:0/P:0+S:0.
func (r *ROLFP) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return r.UnmarshalText([]byte(s))
}

// MarshalText marshals the receiver into its text representation.
func (r *ROLFP) MarshalText() (text []byte, err error) {
	text = []byte(r.String())
	return
}

// UnmarshalText unmarsharls the text representation of a ROLFP into the
// receiver. The function will override any value stored in the fields of the
// receiver with the values in the txt param.
func (r *ROLFP) UnmarshalText(txt []byte) error {
	s := string(txt)
	if s == "" {
		r.IsEmpty = true
		return nil
	}
	parts := strings.Split(s, "+")
	if len(parts) != 2 {
		return errors.Validation(ErrROLFPInvalidText)
	}

	// Pase the scope field.
	scope, err := parseROLFField(parts[1])
	if err != nil {
		return errors.Validation(err)
	}
	r.Scope = scope

	rest := strings.Split(parts[0], "/")
	if len(rest) != 5 {
		return errors.Validation(ErrROLFPInvalidText)
	}

	// Parse the rest of the field.
	val, err := parseROLFField(rest[0])
	if err != nil {
		return errors.Validation(err)
	}
	r.Reputation = val

	val, err = parseROLFField(rest[1])
	if err != nil {
		return errors.Validation(err)
	}
	r.Operation = val

	val, err = parseROLFField(rest[2])
	if err != nil {
		return errors.Validation(err)
	}
	r.Legal = val

	val, err = parseROLFField(rest[3])
	if err != nil {
		return errors.Validation(err)
	}
	r.Financial = val

	val, err = parseROLFField(rest[4])
	if err != nil {
		return errors.Validation(err)
	}
	r.Personal = val

	err = r.Validate()
	if err != nil {
		return errors.Validation(err)
	}
	return nil
}

func parseROLFField(f string) (byte, error) {
	// X:n
	p := strings.Split(f, ":")
	if len(p) != 2 {
		return 0, fmt.Errorf("invalid ROLFP field %s", f)
	}
	v, err := strconv.ParseUint(p[1], 10, 8)
	if err != nil {
		return 0, err
	}
	return byte(v), nil
}
