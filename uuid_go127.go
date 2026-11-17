//go:build go1.27

package squirrel

import "uuid"

// UUID is an array, but represents one SQL value. Keep it intact for the
// driver (database/sql's default converter handles it starting in Go 1.27).
func isStandardUUID(val any) bool {
	_, ok := val.(uuid.UUID)
	return ok
}
