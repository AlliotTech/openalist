package errs

import "errors"

var (
	PermissionDenied = errors.New("permission denied")
	InvalidName      = errors.New("invalid file name")
)
