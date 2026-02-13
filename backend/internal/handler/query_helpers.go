package handler

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/kento/driver/backend/pkg/apperror"
)

// parseIntParam parses an integer query parameter. Returns the default if the
// parameter is absent. Writes a 400 error and returns false if present but
// not a valid integer.
func parseIntParam(w http.ResponseWriter, r *http.Request, name string, defaultVal int) (int, bool) {
	s := r.URL.Query().Get(name)
	if s == "" {
		return defaultVal, true
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", name+" must be a valid integer")
		return 0, false
	}
	return v, true
}

// parseTimeParam parses an RFC3339 time query parameter. Returns zero time if
// absent. Writes a 400 error and returns false if present but invalid.
func parseTimeParam(w http.ResponseWriter, r *http.Request, name string) (time.Time, bool) {
	s := r.URL.Query().Get(name)
	if s == "" {
		return time.Time{}, true
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", name+" must be a valid RFC3339 timestamp")
		return time.Time{}, false
	}
	return t, true
}

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

// isValidPhoneNumber checks that a phone number is 7-15 digits, optionally
// starting with +.
func isValidPhoneNumber(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// isValidGPSCoord checks whether lat/lng are in valid ranges.
// Note: (0, 0) is a valid coordinate (Gulf of Guinea).
func isValidGPSCoord(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
