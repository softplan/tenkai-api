package util

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

// Now creates a timestamp.Timestamp representing the current time.
func Now() *timestamp.Timestamp {
	return Timestamp(time.Now())
}

// Timestamp converts a time.Time to a protobuf *timestamp.Timestamp.
func Timestamp(t time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// Time converts a protobuf *timestamp.Timestamp to a time.Time.
func Time(ts *timestamp.Timestamp) time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

//FormatTimeStamp FormatTimeStamp
func FormatTimeStamp(ts *timestamp.Timestamp, layout string) string {
	return Time(ts).Format(layout)
}

//DateToString DateToString
func DateToString(ts *timestamp.Timestamp) string {
	return FormatTimeStamp(ts, time.ANSIC)
}
