package mutexqueue

import (
	"strconv"
	"testing"

	"github.com/splitio/go-client/splitio/service/dtos"
)

func TestMSImpressionsStorage(t *testing.T) {
	i0 := dtos.ImpressionDTO{FeatureName: "feature0", BucketingKey: "123", ChangeNumber: 123, KeyName: "k0", Time: 123, Treatment: "i0"}
	i1 := dtos.ImpressionDTO{FeatureName: "feature1", BucketingKey: "123", ChangeNumber: 123, KeyName: "k1", Time: 123, Treatment: "i1"}
	i2 := dtos.ImpressionDTO{FeatureName: "feature2", BucketingKey: "123", ChangeNumber: 123, KeyName: "k2", Time: 123, Treatment: "i2"}
	i3 := dtos.ImpressionDTO{FeatureName: "feature3", BucketingKey: "123", ChangeNumber: 123, KeyName: "k3", Time: 123, Treatment: "i3"}
	i4 := dtos.ImpressionDTO{FeatureName: "feature4", BucketingKey: "123", ChangeNumber: 123, KeyName: "k4", Time: 123, Treatment: "i4"}
	i5 := dtos.ImpressionDTO{FeatureName: "feature5", BucketingKey: "123", ChangeNumber: 123, KeyName: "k5", Time: 123, Treatment: "i5"}
	i6 := dtos.ImpressionDTO{FeatureName: "feature6", BucketingKey: "123", ChangeNumber: 123, KeyName: "k6", Time: 123, Treatment: "i6"}
	i7 := dtos.ImpressionDTO{FeatureName: "feature7", BucketingKey: "123", ChangeNumber: 123, KeyName: "k7", Time: 123, Treatment: "i7"}
	i8 := dtos.ImpressionDTO{FeatureName: "feature8", BucketingKey: "123", ChangeNumber: 123, KeyName: "k8", Time: 123, Treatment: "i8"}
	i9 := dtos.ImpressionDTO{FeatureName: "feature9", BucketingKey: "123", ChangeNumber: 123, KeyName: "k9", Time: 123, Treatment: "i9"}

	isFull := make(chan bool, 1)
	queueSize := 20
	queue := NewMQImpressionsStorage(queueSize, isFull)

	if queue.Count() != 0 {
		t.Error("Queue count error")
	}
	if !queue.Empty() {
		t.Error("Queue empty error")
	}

	// LogImpressions from back to front
	queue.LogImpressions([]dtos.ImpressionDTO{i0, i1, i2, i3, i4})

	if queue.Count() != 5 {
		t.Error("Queue count error")
	}
	if queue.Empty() {
		t.Error("Queue empty error")
	}

	queue.LogImpressions([]dtos.ImpressionDTO{i5, i6, i7, i8, i9})

	impressions, _ := queue.PopN(25)

	for i := 0; i < len(impressions); i++ {
		if impressions[i].KeyName != "k"+strconv.Itoa(i) {
			t.Error("KeyName error")
		}

		if impressions[i].FeatureName != "feature"+strconv.Itoa(i) {
			t.Error("FeatureName error")
		}

		if impressions[i].Treatment != "i"+strconv.Itoa(i) {
			t.Error("Treatment error")
		}
	}

}

func TestMSImpressionsStorageMaxSize(t *testing.T) {
	impression := dtos.ImpressionDTO{FeatureName: "feature0", BucketingKey: "123", ChangeNumber: 123, KeyName: "k0", Time: 123, Treatment: "i0"}

	isFull := make(chan bool, 1)
	maxSize := 10
	queue := NewMQImpressionsStorage(maxSize, isFull)

	for i := 0; i < maxSize+1; i++ {
		err := queue.LogImpressions([]dtos.ImpressionDTO{impression})
		if int64(i) < queue.Count() {
			if err != nil {
				t.Error("Error pushing element into queue")
			}
		} else {
			if err != ErrorMaxSizeReached {
				t.Error("Error reporting max size reached")
			}
		}

	}

}
