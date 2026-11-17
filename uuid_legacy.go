//go:build !go1.27

package squirrel

func isStandardUUID(any) bool { return false }
