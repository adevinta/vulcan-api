/*
Copyright 2021 Adevinta
*/

package service

import (
	"context"
)

func (s vulcanitoService) Healthcheck(ctx context.Context) error {
	return s.db.Healthcheck()
}
