package client

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/splitio/go-toolkit/logging"
)

// InputValidation struct is responsible for cheking any input of treatment and
// track methods.
type inputValidation struct {
	logger logging.LoggerInterface
}

func checkNotNull(value interface{}, operation string, name string) error {
	if value == nil {
		return errors.New(operation + ": " + name + " cannot be nil")
	}
	return nil
}

func checkIsNumeric(value interface{}) (string, error) {
	f, float := value.(float64)
	i, integer := value.(int)
	i32, integer32 := value.(int32)
	i64, integer64 := value.(int64)

	if float {
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	}
	if integer {
		return strconv.Itoa(i), nil
	}
	if integer32 {
		return strconv.FormatInt(int64(i32), 10), nil
	}
	if integer64 {
		return strconv.FormatInt(i64, 10), nil
	}
	return "", errors.New("Value is not of type numeric")
}

// ValidateTreatmentKey implements the validation for Treatment call
func (i *inputValidation) ValidateTreatmentKey(key interface{}) (string, *string, error) {
	if key == nil {
		return "", nil, errors.New("Treatment: key cannot be nil")
	}
	numberAsString, err := checkIsNumeric(key)
	if err == nil {
		i.logger.Warning(fmt.Sprintf("Treatment: matchingKey %s is not of type string, converting.", numberAsString))
		return numberAsString, nil, nil
	}
	sMatchingKey, ok := key.(string)
	if ok {
		return sMatchingKey, nil, nil
	}
	okey, ok := key.(*Key)
	if ok {
		return okey.MatchingKey, &okey.BucketingKey, nil
	}
	return "", nil, errors.New("Treatment: supplied key is neither a string or a Key struct")
}

// ValidateTrackInputs implements the validation for Track call
func ValidateTrackInputs(key string, trafficType string, eventType string, value interface{}) (string, string, string, interface{}, error) {
	var r = regexp.MustCompile(`[a-zA-Z0-9][-_\.a-zA-Z0-9]{0,62}`)
	if strings.TrimSpace(trafficType) == "" {
		return key, "", eventType, value, errors.New("Track: trafficType must not be an empty String")
	}
	if !r.MatchString(eventType) {
		return key, trafficType, "", value, errors.New("Track: eventName must adhere to the regular expression [a-zA-Z0-9][-_\\.a-zA-Z0-9]{0,62}")
	}
	_, err := checkIsNumeric(value)
	if value != nil && err != nil {
		return key, trafficType, eventType, 0, errors.New("Track: value must be a number")
	}
	return key, trafficType, eventType, value, nil
}
