package provisional

import (
	"fmt"
	"sync"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/splitio/go-toolkit/provisional/int64cache"
)

// ImpressionObserver is used to check wether an impression has been previously seen
type ImpressionObserver interface {
	TestAndSet(featureName string, keyImpression *dtos.ImpressionRecord) (int64, error)
}

// ImpressionObserverImpl is an implementation of the ImpressionObserver interface
type ImpressionObserverImpl struct {
	cache  int64cache.Int64Cache
	hasher ImpressionHasher
	mutex  sync.Mutex
}

// Atomically fetch cache data and update it
func (o *ImpressionObserverImpl) testAndSet(key int64, newValue int64) (int64, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	old, err := o.cache.Get(key)
	o.cache.Set(key, newValue)
	return old, err
}

// TestAndSet hashes the impression, updates the cache and returns the previous value
func (o *ImpressionObserverImpl) TestAndSet(featureName string, keyImpression *dtos.ImpressionRecord) (int64, error) {
	hash, err := o.hasher.Process(featureName, keyImpression)
	if err != nil {
		return 0, fmt.Errorf("error hashing impression: %s", err.Error())
	}

	return o.testAndSet(hash, keyImpression.Time)
}

// NewImpressionObserver constructs a new ImpressionObserver
func NewImpressionObserver(size int) (*ImpressionObserverImpl, error) {
	cache, err := int64cache.NewInt64Cache(size)
	if err != nil {
		return nil, fmt.Errorf("error building cache: %s", err.Error())
	}
	return &ImpressionObserverImpl{
		cache:  cache,
		hasher: &ImpressionHasherImpl{},
		mutex:  sync.Mutex{},
	}, nil
}

// ImpressionObserverNoOp is an implementation of the ImpressionObserver interface
type ImpressionObserverNoOp struct{}

// TestAndSet that does nothing
func (o *ImpressionObserverNoOp) TestAndSet(featureName string, keyImpression *dtos.ImpressionRecord) (int64, error) {
	return 0, nil
}
