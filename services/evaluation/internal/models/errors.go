package models

import (
	"database/sql"
	"errors"
	"runtime"
	"strconv"
)

var ErrBadRequest400 = errors.New("bad request - Problem with the request")
var ErrUnauthorized401 = errors.New("unauthorized - Access token is missing or invalid")
var ErrForbidden403 = errors.New("forbidden")
var ErrNotFound404 = errors.New("not found - Requested entity is not found in database")
var ErrConflict409 = errors.New("conflict - UserDB already exists")
var ErrProjectAlreadyProcessing = errors.New("project is already being processed - cannot upload files while processing")
var ErrServerError500 = errors.New("internal server error - Request is valid but operation failed at server side")
var ErrServerError503 = errors.New("service unavailable")

func StacktraceError(errs ...error) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return errs[0]
	}
	return errors.Join(
		errors.Join(errs...),
		errors.New("				at "+file+":"+strconv.Itoa(line)),
	)
}

func CheckError(err error) (int, string) {
	if errors.Is(err, ErrBadRequest400) {
		return 400, ErrBadRequest400.Error()
	}

	if errors.Is(err, ErrUnauthorized401) {
		return 401, ErrUnauthorized401.Error()
	}

	if errors.Is(err, ErrForbidden403) {
		return 403, ErrForbidden403.Error()
	}

	if errors.Is(err, ErrNotFound404) ||
		errors.Is(err, sql.ErrNoRows) {
		return 404, ErrNotFound404.Error()
	}

	if errors.Is(err, ErrConflict409) {
		return 409, ErrConflict409.Error()
	}

	if errors.Is(err, ErrProjectAlreadyProcessing) {
		return 409, ErrProjectAlreadyProcessing.Error()
	}

	if errors.Is(err, ErrBadRequest400) {
		return 400, ErrBadRequest400.Error()
	}

	if errors.Is(err, ErrServerError503) {
		return 503, ErrServerError503.Error()
	}

	return 500, ErrServerError500.Error()
}
