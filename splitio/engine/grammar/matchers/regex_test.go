package matchers

import (
	"bufio"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-toolkit/v5/logging"
)

func TestRegexMatcher(t *testing.T) {
	logger := logging.NewLogger(&logging.LoggerOptions{})
	attrName := "value"

	fileHandle, err := os.Open("../../../../testdata/regex.txt")
	if err != nil {
		t.Error("Cannot open test files")
		return
	}
	defer fileHandle.Close()

	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "#")

		regex := line[0]
		text := line[1]
		shouldMatch, err := strconv.ParseBool(line[2])
		if err != nil {
			t.Error("Error parsing boolean")
			return
		}

		dto := &dtos.MatcherDTO{
			MatcherType: "MATCHES_STRING",
			String:      &regex,
			KeySelector: &dtos.KeySelectorDTO{
				Attribute: &attrName,
			},
		}

		matcher, err := BuildMatcher(dto, nil, logger)
		if err != nil {
			t.Error("There should be no errors when building the matcher")
			t.Error(err)
		}

		matcherType := reflect.TypeOf(matcher).String()
		if matcherType != "*matchers.RegexMatcher" {
			t.Errorf("Incorrect matcher constructed. Should be *matchers.RegexMatcher and was %s", matcherType)
		}

		matches := matcher.Match("asd", map[string]interface{}{"value": text}, nil)
		if matches != shouldMatch {
			t.Errorf("Match for text %s and regex %s should be %t", text, regex, shouldMatch)
		}
	}
}
