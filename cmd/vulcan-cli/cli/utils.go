/*
Copyright 2021 Adevinta
*/

package cli

import (
	"bufio"
	"os"
)

// ReadLines reads a whole file into memory
// and returns a slice of its lines.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func DereferenceString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func PtrString(s string) *string {
	return &s
}

func DereferenceBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
