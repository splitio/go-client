package provisional

import (
	"fmt"
	"testing"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/provisional/int64cache"
)

func TestImpressionObserver(t *testing.T) {
	observer, err := NewImpressionObserver(5)
	if err != nil {
		t.Error("there should not be an error: ", err.Error())
	}

	imp := dtos.ImpressionRecord{
		BucketingKey: "someBuck",
		ChangeNumber: 123,
		KeyName:      "someKey",
		Label:        "someLabel",
		Time:         123456,
		Treatment:    "on",
	}

	res, err := observer.TestAndSet("someFeature", &imp)
	if err == nil {
		t.Error("there should be an error indicating that the result isn't cached yet: ", err.Error())
	}

	if _, ok := err.(*int64cache.Miss); !ok {
		t.Error("The error should be of type miss")
	}

	if res != 0 {
		t.Error("previous value should be default/empty/zero. Is ", res)
	}

	res, err = observer.TestAndSet("someFeature", &imp)
	if err != nil {
		t.Error("there should not have been an error. got: ", err.Error())
	}

	if res != 123456 {
		t.Error("previous value should be 123456. Is ", res)
	}

	// add 5 new impressions and validate that the first one returns 0, *Miss again
	for i := 0; i < 5; i++ {
		observer.TestAndSet("someFeature", &dtos.ImpressionRecord{
			KeyName:      fmt.Sprintf("key_%d", i),
			ChangeNumber: 123,
			Label:        "someLabel",
			Time:         1,
		})
	}

	res, err = observer.TestAndSet("someFeature", &imp)
	if err == nil {
		t.Error("there should be an error indicating that the result isn't cached yet: ", err.Error())
	}

	if _, ok := err.(*int64cache.Miss); !ok {
		t.Error("The error should be of type miss")
	}

	if res != 0 {
		t.Error("previous value should be default/empty/zero. Is ", res)
	}

}
