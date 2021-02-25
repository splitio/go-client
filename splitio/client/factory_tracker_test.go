package client

import (
	"testing"

	"github.com/splitio/go-client/v6/splitio/conf"
	"github.com/splitio/go-toolkit/v4/logging"
)

func TestFactoryTrackerMultipleInstantiation(t *testing.T) {
	sdkConf := conf.Default()
	sdkConf.Logger = logging.NewLogger(&logging.LoggerOptions{
		LogLevel:      5,
		ErrorWriter:   &mW,
		WarningWriter: &mW,
		InfoWriter:    &mW,
		DebugWriter:   &mW,
		VerboseWriter: &mW,
	})
	sdkConf.SplitFile = "../../testdata/splits.yaml"

	delete(factoryInstances, conf.Localhost)
	delete(factoryInstances, "something")

	factory, _ := NewSplitFactory(conf.Localhost, sdkConf)
	client := factory.Client()

	if factoryInstances[conf.Localhost] != 1 {
		t.Error("It should be 1")
	}

	factory2, _ := NewSplitFactory(conf.Localhost, sdkConf)
	_ = factory2.Client()

	if factoryInstances[conf.Localhost] != 2 {
		t.Error("It should be 2")
	}
	expected := "Factory Instantiation: You already have 1 factory with this API Key. We recommend keeping only one " +
		"instance of the factory at all times (Singleton pattern) and reusing it throughout your application."
	if !mW.Matches(expected) {
		t.Error("Error is distinct from the expected one")
	}

	factory4, _ := NewSplitFactory("asdadd", sdkConf)
	client2 := factory4.Client()
	expected = "Factory Instantiation: You already have an instance of the Split factory. Make sure you definitely want " +
		"this additional instance. We recommend keeping only one instance of the factory at all times (Singleton pattern) and " +
		"reusing it throughout your application."
	if !mW.Matches(expected) {
		t.Error("Error is distinct from the expected one")
	}

	client.Destroy()

	if factoryInstances[conf.Localhost] != 1 {
		t.Error("It should be 1")
	}

	if factoryInstances["asdadd"] != 1 {
		t.Error("It should be 1")
	}

	client.Destroy()

	if factoryInstances[conf.Localhost] != 1 {
		t.Error("It should be 1")
	}

	client2.Destroy()

	_, exist := factoryInstances["asdadd"]
	if exist {
		t.Error("It should not exist")
	}

	factory3, _ := NewSplitFactory(conf.Localhost, sdkConf)
	_ = factory3.Client()
	expected = "Factory Instantiation: You already have 1 factory with this API Key. We recommend keeping only one " +
		"instance of the factory at all times (Singleton pattern) and reusing it throughout your application."
	if !mW.Matches(expected) {
		t.Error("Error is distinct from the expected one")
	}

	if factoryInstances[conf.Localhost] != 2 {
		t.Error("It should be 2")
	}

	factory5, _ := NewSplitFactory(conf.Localhost, sdkConf)
	_ = factory5.Client()
	expected = "Factory Instantiation: You already have 2 factories with this API Key. We recommend keeping only one " +
		"instance of the factory at all times (Singleton pattern) and reusing it throughout your application."
	if !mW.Matches(expected) {
		t.Error("Error is distinct from the expected one")
	}
	if factoryInstances[conf.Localhost] != 3 {
		t.Error("It should be 3")
	}

	delete(factoryInstances, conf.Localhost)
	delete(factoryInstances, "asdadd")
}
