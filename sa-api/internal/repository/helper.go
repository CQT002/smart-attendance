package repository

import "strings"

// isDuplicateError kiểm tra lỗi có phải duplicate key violation không (PostgreSQL error code 23505)
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "23505") ||
		strings.Contains(err.Error(), "unique constraint")
}
