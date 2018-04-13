package to

import (
	"encoding/base64"
	"time"
)

// Strp return a string pointer from string
func Strp(s string) *string {
	return &s
}

func Strs(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func Timep(s time.Time) *time.Time {
	return &s
}

func Intp(s int) *int {
	return &s
}

func Int64p(s int64) *int64 {
	return &s
}

func Float64p(s float64) *float64 {
	return &s
}

func Boolp(s bool) *bool {
	return &s
}

func ABytep(s []byte) *[]byte {
	return &s
}

////////
// Base64
////////

func Base64(str *string) string {
	if str == nil {
		return base64.StdEncoding.EncodeToString([]byte(""))
	}
	return base64.StdEncoding.EncodeToString([]byte(*str))
}

func Base64p(str *string) *string {
	return Strp(Base64(str))
}
