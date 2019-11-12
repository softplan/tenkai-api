package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDateTime(t *testing.T) {
	dt := DateToString(Now())
	assert.NotNil(t, dt)

	dt = FormatTimeStamp(Now(), "2006-12-01")
	assert.NotNil(t, dt)

	dtx := Timestamp(time.Now())
	assert.NotNil(t, dtx)

	now := Now()
	assert.NotNil(t, now)
}
