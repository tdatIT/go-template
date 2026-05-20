package svcerr

import (
	"github.com/tdatIT/go-template/internal/domain/enums"
)

var (
	ErrInternalServer = &Error{
		Status:  500,
		Code:    enums.Internal,
		Message: "Internal server error",
	}

	ErrBadRequest = &Error{
		Status:  400,
		Code:    enums.InvalidArgument,
		Message: "Bad request",
	}

	ErrMissingIdParam = &Error{
		Status:  400,
		Code:    enums.InvalidArgument,
		Message: "Missing id param",
	}

	ErrInvalidIdParam = &Error{
		Status:  400,
		Code:    enums.InvalidArgument,
		Message: "Invalid id param",
	}

	ErrNotChange = &Error{
		Status:  200,
		Code:    enums.Ok,
		Message: "Not change",
	}

	ErrPermissionDenied = &Error{
		Status:  403,
		Code:    enums.PermissionDenied,
		Message: "Permission denied",
	}

	ErrNotFound = &Error{
		Status:  404,
		Code:    enums.NotFound,
		Message: "Not found",
	}

	ErrForbidden = &Error{
		Status:  403,
		Code:    enums.PermissionDenied,
		Message: "Forbidden",
	}

	ErrRequestTimeout = &Error{
		Status:  408,
		Code:    enums.DeadlineExceeded,
		Message: "Request timeout",
	}

	ErrAlreadyExists = &Error{
		Status:  409,
		Code:    enums.AlreadyExists,
		Message: "Already exists",
	}

	ErrUnauthenticated = &Error{
		Status:  401,
		Code:    enums.Unauthenticated,
		Message: "Unauthorized",
	}

	ErrRecordNotFound = &Error{
		Status:  404,
		Code:    enums.RecordNotFound,
		Message: "Record does not exist",
	}

	ErrInvalidParameters = &Error{
		Status:  400,
		Code:    enums.FailedPrecondition,
		Message: "Invalid parameters",
	}

	ErrTooManyRequest = &Error{
		Status:  429,
		Code:    enums.ResourceExhausted,
		Message: "Too Many Requests",
	}

	ErrInvalidFilter = &Error{
		Status:  400,
		Code:    enums.InvalidArgument,
		Message: "Invalid filters",
	}

	ErrMethodNotAllowed = &Error{
		Status:  405,
		Code:    enums.MethodNotAllowed,
		Message: "Method not allowed",
	}
)
