package client

import (
	"testing"

	"github.com/splitio/go-split-commons/v8/flagsets"
	"github.com/splitio/go-toolkit/v5/logging/mocks"

	"github.com/stretchr/testify/assert"
)

func TestPrintWarnings(t *testing.T) {
	flagSets, warnings := flagsets.SanitizeMany([]string{"set1", " set2"})
	assert.Len(t, flagSets, 2)
	logger := &mocks.LoggerMock{}
	logger.On("Warning", []interface{}{"Flag Set name  set2 has extra whitespace, trimming"}).Return().Once()
	printWarnings(logger, warnings)

	flagSets, warnings = flagsets.SanitizeMany([]string{"set1", "Set2"})
	assert.Len(t, flagSets, 2)
	logger.On("Warning", []interface{}{"Flag Set name Set2 should be all lowercase - converting string to lowercase"}).Return().Once()
	printWarnings(logger, warnings)

	flagSets, warnings = flagsets.SanitizeMany([]string{"set1", "@set4"})
	assert.Len(t, flagSets, 1)
	logger.On("Warning", []interface{}{"you passed @set4, Flag Set must adhere to the regular expressions ^[a-z0-9][_a-z0-9]{0,49}$. This means a Flag Set must start with a letter or number, be in lowercase, alphanumeric and have a max length of 50 characters. @set4 was discarded."}).Return().Once()
	printWarnings(logger, warnings)

	logger.AssertExpectations(t)
}
