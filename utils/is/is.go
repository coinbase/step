package is

import "time"

func EmptyStr(v *string) bool {
	return v == nil || *v == ""
}

// UniqueStrp will check a list of string pointers for unique values
// It will return false if any element is nil
func UniqueStrp(strs []*string) bool {
	seen := map[string]bool{}
	for _, s := range strs {
		if s == nil {
			return false
		}
		if seen[*s] {
			return false
		}
		seen[*s] = true
	}
	return true
}

// WithinTimeFrame returns if a time is after and before time from now
func WithinTimeFrame(tt *time.Time, diff_back time.Duration, diff_forward time.Duration) bool {
	// -1 make it subtract
	ago := time.Now().Add(-1 * diff_back)

	ahead := time.Now().Add(diff_forward)

	return tt.After(ago) && tt.Before(ahead)
}
