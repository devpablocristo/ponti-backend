package dashboard

import (
	"time"
)

// Test utility functions shared across all test files

// timePtr creates a pointer to a time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}

// stringPtr creates a pointer to a string value
func stringPtr(s string) *string {
	return &s
}

// int64Ptr creates a pointer to an int64 value
func int64Ptr(i int64) *int64 {
	return &i
}
