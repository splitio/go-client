package mutexqueue

import (
	"strconv"
	"testing"

	"github.com/splitio/go-client/splitio/storage"
	"github.com/splitio/go-toolkit/logging"
)

func TestMSImpressionsStorage(t *testing.T) {
	logger := logging.NewLogger(nil)

	i0 := storage.Impression{FeatureName: "feature0", BucketingKey: "123", ChangeNumber: 123, KeyName: "k0", Time: 123, Treatment: "i0"}
	i1 := storage.Impression{FeatureName: "feature1", BucketingKey: "123", ChangeNumber: 123, KeyName: "k1", Time: 123, Treatment: "i1"}
	i2 := storage.Impression{FeatureName: "feature2", BucketingKey: "123", ChangeNumber: 123, KeyName: "k2", Time: 123, Treatment: "i2"}
	i3 := storage.Impression{FeatureName: "feature3", BucketingKey: "123", ChangeNumber: 123, KeyName: "k3", Time: 123, Treatment: "i3"}
	i4 := storage.Impression{FeatureName: "feature4", BucketingKey: "123", ChangeNumber: 123, KeyName: "k4", Time: 123, Treatment: "i4"}
	i5 := storage.Impression{FeatureName: "feature5", BucketingKey: "123", ChangeNumber: 123, KeyName: "k5", Time: 123, Treatment: "i5"}
	i6 := storage.Impression{FeatureName: "feature6", BucketingKey: "123", ChangeNumber: 123, KeyName: "k6", Time: 123, Treatment: "i6"}
	i7 := storage.Impression{FeatureName: "feature7", BucketingKey: "123", ChangeNumber: 123, KeyName: "k7", Time: 123, Treatment: "i7"}
	i8 := storage.Impression{FeatureName: "feature8", BucketingKey: "123", ChangeNumber: 123, KeyName: "k8", Time: 123, Treatment: "i8"}
	i9 := storage.Impression{FeatureName: "feature9", BucketingKey: "123", ChangeNumber: 123, KeyName: "k9", Time: 123, Treatment: "i9"}

	isFull := make(chan string, 1)
	queueSize := 20
	queue := NewMQImpressionsStorage(queueSize, isFull, logger)

	if queue.Count() != 0 {
		t.Error("Queue count error")
	}
	if !queue.Empty() {
		t.Error("Queue empty error")
	}

	// LogImpressions from back to front
	queue.LogImpressions([]storage.Impression{i0, i1, i2, i3, i4})

	if queue.Count() != 5 {
		t.Error("Queue count error")
	}
	if queue.Empty() {
		t.Error("Queue empty error")
	}

	queue.LogImpressions([]storage.Impression{i5, i6, i7, i8, i9})

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
	logger := logging.NewLogger(nil)
	impression := storage.Impression{FeatureName: "feature0", BucketingKey: "123", ChangeNumber: 123, KeyName: "k0", Time: 123, Treatment: "i0"}

	isFull := make(chan string, 1)
	maxSize := 10
	queue := NewMQImpressionsStorage(maxSize, isFull, logger)

	select {
	case <-isFull:
		t.Error("Signal sent when it shouldn't have!")
	default:
	}

	for i := 0; i < maxSize; i++ {
		err := queue.LogImpressions([]storage.Impression{impression})
		if err != nil {
			t.Error("Error pushing element into queue")
		}
	}

	err := queue.LogImpressions([]storage.Impression{impression})
	if err != ErrorMaxSizeReached {
		t.Error("It should return error")
	}

	select {
	case <-isFull:
	default:
		t.Error("Signal sent when it shouldn't have!")
	}
}
