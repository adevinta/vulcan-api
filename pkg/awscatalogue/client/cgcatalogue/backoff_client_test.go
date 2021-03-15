/*
Copyright 2021 Adevinta
*/

package cgcatalogue

import (
	"errors"
	"testing"
)

func Test_execWithBackOff(t *testing.T) {
	t.Run("ResturnsWhenNoError", func(t *testing.T) {
		// Returns correctly when no errors.
		var times int
		err := execWithBackOff(1, 1, 0.1, func() error {
			times++
			return nil
		})
		if err != nil || times != 1 {
			t.Errorf("unexpected error or times values: %+v,%d", err, times)
		}
	})

	t.Run("ExecutesNoMoreThanNRetries", func(t *testing.T) {
		// Returns correctly when no errors.
		var times = 0
		err := execWithBackOff(1, 2, 0.1, func() error {
			times++
			return errors.New("an error")
		})
		if err == nil || times != 3 {
			t.Errorf("unexpected error or times values: %+v, %d", err, times)
		}
	})

	t.Run("ResturnsWhenUnexpectedStatusError", func(t *testing.T) {
		// Returns correctly when no errors.
		var times int
		err := execWithBackOff(1, 1, 0.1, func() error {
			times++
			return &unexpectedStatusError{}
		})
		var gotErr *unexpectedStatusError
		if err == nil || times != 1 || !errors.As(err, &gotErr) {
			t.Errorf("unexpected error or times values: %+v,%d", err, times)
		}
	})
}
