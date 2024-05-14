package client

import (
	"testing"

	"github.com/splitio/go-split-commons/v6/flagsets"
)

func TestPrintWarnings(t *testing.T) {

	flagSets, warnings := flagsets.SanitizeMany([]string{"set1", " set2"})
	if len(flagSets) != 2 {
		t.Error("flag set size should be 2")
	}
	printWarnings(getMockedLogger(), warnings)
	if !mW.Matches("Flag Set name  set2 has extra whitespace, trimming") {
		t.Error("Wrong message")
	}
	flagSets, warnings = flagsets.SanitizeMany([]string{"set1", "Set2"})
	if len(flagSets) != 2 {
		t.Error("flag set size should be 2")
	}
	printWarnings(getMockedLogger(), warnings)
	if !mW.Matches("Flag Set name Set2 should be all lowercase - converting string to lowercase") {
		t.Error("Wrong message")
	}
	flagSets, warnings = flagsets.SanitizeMany([]string{"set1", "@set4"})
	if len(flagSets) != 1 {
		t.Error("flag set size should be 1")
	}
	printWarnings(getMockedLogger(), warnings)
	if !mW.Matches("you passed @set4, Flag Set must adhere to the regular expressions ^[a-z0-9][_a-z0-9]{0,49}$. This means a Flag Set must start with a letter or number, be in lowercase, alphanumeric and have a max length of 50 characters. @set4 was discarded.") {
		t.Error("Wrong message")
	}
}
