/*
Copyright 2021 Adevinta
*/

package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/adevinta/vulcan-api/pkg/jwt"
)

var (
	key   string
	usage = `usage: test-tools [-key SECRETKEY] <email_address>`
)

func init() {
	flag.StringVar(&key, "key", "SUPERSECRETSIGNKEY", "Key used to sign the JWT token")
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		errExit(usage)
	}

	email := args[0]

	jwtConfig := jwt.NewJWTConfig(key)
	tokenGenTime := time.Now()
	token, err := jwtConfig.GenerateToken(map[string]interface{}{
		"iat":  tokenGenTime.Unix(),
		"sub":  email,
		"type": "API",
	})
	if err != nil {
		errExit(err.Error())
	}

	fmt.Printf("Token: %v\n", token)
	fmt.Printf("Fingerprint to store in DB: %x\n", sha256.Sum256([]byte(token)))
}

func errExit(e string) {
	fmt.Fprintln(os.Stderr, e)
	os.Exit(1)
}
