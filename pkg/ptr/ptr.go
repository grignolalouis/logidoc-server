// Package ptr provides helpers to create pointers to primitive values.
package ptr

func Float64(f float64) *float64 { return &f }
func Int(i int) *int             { return &i }
