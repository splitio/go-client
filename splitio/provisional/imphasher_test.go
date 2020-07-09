package provisional

import (
	"github.com/splitio/go-client/splitio/service/dtos"
	"testing"
)

func TestHasher(t *testing.T) {
	hasher := ImpressionHasherImpl{}

	seen := map[int64]struct{}{}

	imp := dtos.ImpressionRecord{
		BucketingKey: "someBuck",
		ChangeNumber: 123,
		KeyName:      "someKey",
		Label:        "someLabel",
		Time:         123456,
		Treatment:    "on",
	}

	h, _ := hasher.Process("someFeature", &imp)
	seen[h] = struct{}{}
	if len(seen) != 1 {
		t.Error("should have one item")
	}

	h, _ = hasher.Process("someFeature", &imp)
	seen[h] = struct{}{}
	if len(seen) != 1 {
		t.Error("hashing the same impression twice should not change the size of the map")
	}

	imp.ChangeNumber = 1234
	h, _ = hasher.Process("someFeature", &imp)
	seen[h] = struct{}{}
	if len(seen) != 2 {
		t.Error("different hashes should increase the size of the map")
	}

	imp.KeyName = "someOtherKey"
	h, _ = hasher.Process("someFeature", &imp)
	seen[h] = struct{}{}
	if len(seen) != 3 {
		t.Error("different hashes should increase the size of the map")
	}

	imp.Label = "someOtherLabel"
	h, _ = hasher.Process("someFeature", &imp)
	seen[h] = struct{}{}
	if len(seen) != 4 {
		t.Error("different hashes should increase the size of the map")
	}

	imp.Treatment = "someOtherTreatment"
	h, _ = hasher.Process("someFeature", &imp)
	seen[h] = struct{}{}
	if len(seen) != 5 {
		t.Error("different hashes should increase the size of the map")
	}

	h, _ = hasher.Process("someOtherFeature", &imp)
	seen[h] = struct{}{}
	if len(seen) != 6 {
		t.Error("different hashes should increase the size of the map")
	}

	_, err := hasher.Process("someOtherFeature", nil)
	if err == nil {
		t.Error("passing nil should return an error")
	}
}
