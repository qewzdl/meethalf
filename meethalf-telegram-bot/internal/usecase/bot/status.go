package bot

import (
	"errors"
	"net/http"
)

func isBannedError(err error) bool {
	if err == nil {
		return false
	}

	var status statusError
	if !errors.As(err, &status) {
		return false
	}

	return status.StatusCode() == http.StatusForbidden
}
