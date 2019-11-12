package repository

import (
	"database/sql/driver"
	"time"
)

//AnyTime AnyTime
type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}
