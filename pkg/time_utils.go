package pkg

import "time"

// TimeNow returns the current time. This function exists to allow
// mocking time in tests if needed.
func TimeNow() time.Time {
	return time.Now()
}
