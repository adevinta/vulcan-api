/*
Copyright 2021 Adevinta
*/

package api

type Role string

const (
	Owner  Role = "owner"
	Member Role = "member"
)

func (role Role) Valid() bool {
	switch role {
	case Member, Owner:
		return true
	default:
		return false
	}
}
