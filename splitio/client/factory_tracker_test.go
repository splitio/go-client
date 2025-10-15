package client

import (
	"testing"

	"github.com/splitio/go-client/v6/splitio/conf"

	"github.com/splitio/go-toolkit/v5/logging/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFactoryTrackerMultipleInstantiation(t *testing.T) {
	logger := &mocks.LoggerMock{}
	logger.On("Info", mock.Anything).Return()
	logger.On("Debug", mock.Anything).Return()
	sdkConf := conf.Default()
	sdkConf.Logger = logger
	sdkConf.SplitFile = "../../testdata/splits.yaml"

	removeInstanceFromTracker(conf.Localhost)
	removeInstanceFromTracker("something")

	factory, _ := NewSplitFactory(conf.Localhost, sdkConf)
	client := factory.Client()
	assert.Equal(t, int64(1), factoryInstances[conf.Localhost])

	logger.On("Warning", []interface{}{"Factory Instantiation: You already have 1 factory with this SDK Key. We recommend keeping only one instance of the factory at all times (Singleton pattern) and reusing it throughout your application."}).Return().Once()
	factory2, _ := NewSplitFactory(conf.Localhost, sdkConf)
	_ = factory2.Client()
	assert.Equal(t, int64(2), factoryInstances[conf.Localhost])

	logger.On("Warning", []interface{}{"Factory Instantiation: You already have an instance of the Split factory. Make sure you definitely want this additional instance. We recommend keeping only one instance of the factory at all times (Singleton pattern) and reusing it throughout your application."}).Return().Once()
	factory4, _ := NewSplitFactory("asdadd", sdkConf)
	client2 := factory4.Client()
	client.Destroy()

	assert.Equal(t, int64(1), factoryInstances[conf.Localhost])
	assert.Equal(t, int64(1), factoryInstances["asdadd"])
	client.Destroy()

	assert.Equal(t, int64(1), factoryInstances[conf.Localhost])
	client2.Destroy()

	_, exist := factoryInstances["asdadd"]
	assert.False(t, exist)

	logger.On("Warning", []interface{}{"Factory Instantiation: You already have 1 factory with this SDK Key. We recommend keeping only one instance of the factory at all times (Singleton pattern) and reusing it throughout your application."}).Return().Once()
	factory3, _ := NewSplitFactory(conf.Localhost, sdkConf)
	_ = factory3.Client()
	assert.Equal(t, int64(2), factoryInstances[conf.Localhost])

	logger.On("Warning", []interface{}{"Factory Instantiation: You already have 2 factories with this SDK Key. We recommend keeping only one instance of the factory at all times (Singleton pattern) and reusing it throughout your application."}).Return().Once()
	factory5, _ := NewSplitFactory(conf.Localhost, sdkConf)
	_ = factory5.Client()
	assert.Equal(t, int64(3), factoryInstances[conf.Localhost])

	removeInstanceFromTracker(conf.Localhost)
	removeInstanceFromTracker("asdadd")
}
