package operator

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/sirupsen/logrus"
)

const logIdentifierLength = 10

func randomString(rand *rand.Rand, size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func newLogIdentifier(r *rand.Rand) logrus.Fields {
	return logrus.Fields{
		"logID": randomString(r, logIdentifierLength),
	}
}

func newRequestLogger(logger logrus.FieldLogger, r *http.Request, rand *rand.Rand) logrus.FieldLogger {
	return logger.WithFields(logrus.Fields{
		"method": r.Method,
		"url":    r.URL.String(),
	}).WithFields(newLogIdentifier(rand))
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeErrorResponse(logger logrus.FieldLogger, w http.ResponseWriter, r *http.Request, status int, message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	writeResponseAsJSON(logger, w, status, errorResponse{Error: msg})
}

// writeResponseAsJSON attempts to marshal an arbitrary thing to JSON then write
// it to the http.ResponseWriter
func writeResponseAsJSON(logger logrus.FieldLogger, w http.ResponseWriter, code int, resp interface{}) {
	enc, err := json.Marshal(resp)
	if err != nil {
		logger.WithError(err).Error("failed JSON-encoding HTTP response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err = w.Write(enc); err != nil {
		logger.WithError(err).Error("failed writing HTTP response")
	}
}

func validateHour(hour int64) error {
	if hour >= 0 && hour <= 23 {
		return nil
	}
	return fmt.Errorf("invalid hour: %d, must be between 0 and 23", hour)
}

func validateMinute(minute int64) error {
	if minute >= 0 && minute <= 59 {
		return nil
	}
	return fmt.Errorf("invalid minute: %d, must be between 0 and 59", minute)
}

func validateSecond(second int64) error {
	if second >= 0 && second <= 59 {
		return nil
	}
	return fmt.Errorf("invalid second: %d, must be between 0 and 59", second)
}

func validateDayOfMonth(dom int64) error {
	if dom >= 1 && dom <= 31 {
		return nil
	}
	return fmt.Errorf("invalid day of month: %d, must be between 1 and 31", dom)
}
