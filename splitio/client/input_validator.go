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

// ValidateTreatmentKey implements the validation for Treatment call
func (i *inputValidation) ValidateTreatmentKey(key interface{}) (string, *string, error) {
	if key == nil {
		return "", nil, errors.New("Treatment: key cannot be nil")
	}
	iMatchingKey, ok := key.(int)
	if ok {
		convertedMatchingKey := strconv.Itoa(iMatchingKey)
		i.logger.Warning(fmt.Sprintf("Treatment: matchingKey %s is not of type string, converting.", convertedMatchingKey))
		return convertedMatchingKey, nil, nil
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
	if r.MatchString(eventType) == false {
		return key, trafficType, "", value, errors.New("Track: eventName must adhere to the regular expression [a-zA-Z0-9][-_\\.a-zA-Z0-9]{0,62}")
	}
	if value == nil {
		return key, trafficType, eventType, 0, errors.New("Track: value must be a number")
	}
	_, float := value.(float64)
	_, integer := value.(int)
	_, integer32 := value.(int32)
	_, integer64 := value.(int64)
	if !float && !integer && !integer32 && !integer64 {
		return key, trafficType, eventType, 0, errors.New("Track: value must be a number")
	}
	return key, trafficType, eventType, value, nil
}
