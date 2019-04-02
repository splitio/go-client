// Package api contains all functions and dtos Split APIs
package dtos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

var splitsMock, _ = ioutil.ReadFile("../../../testdata/splits_mock.json")
var splitMock, _ = ioutil.ReadFile("../../../testdata/split_mock.json")
var segmentMock, _ = ioutil.ReadFile("../../../testdata/segment_mock.json")

func TestSplitDTO(t *testing.T) {
	mockedData := fmt.Sprintf(string(splitsMock), splitMock)

	var splitChangesDtoFromMock SplitChangesDTO
	var splitChangesDtoFromMarshal SplitChangesDTO

	err := json.Unmarshal([]byte(mockedData), &splitChangesDtoFromMock)
	if err != nil {
		t.Error("Error parsing split changes JSON ", err)
	}

	if dataSerialize, err := splitChangesDtoFromMock.Splits[0].MarshalBinary(); err != nil {
		t.Error(err)
	} else {
		marshalData := fmt.Sprintf(string(splitsMock), dataSerialize)
		err2 := json.Unmarshal([]byte(marshalData), &splitChangesDtoFromMarshal)
		if err2 != nil {
			t.Error("Error parsing split changes JSON ", err)
		}

		if splitChangesDtoFromMarshal.Splits[0].ChangeNumber !=
			splitChangesDtoFromMock.Splits[0].ChangeNumber {
			t.Error("Marshal struct mal formed [ChangeNumber]")
		}

		if splitChangesDtoFromMarshal.Splits[0].Name !=
			splitChangesDtoFromMock.Splits[0].Name {
			t.Error("Marshal struct mal formed [Name]")
		}

		if splitChangesDtoFromMarshal.Splits[0].Killed !=
			splitChangesDtoFromMock.Splits[0].Killed {
			t.Error("Marshal struct mal formed [Killed]")
		}

		if splitChangesDtoFromMarshal.Splits[0].Configurations == nil {
			t.Error("Marshal struct mal formed [Configurations]")
		}

		if splitChangesDtoFromMarshal.Splits[0].Configurations["of"] !=
			splitChangesDtoFromMock.Splits[0].Configurations["of"] {
			t.Error("Marshal struct mal formed [Configurations]")
		}

		if splitChangesDtoFromMarshal.Splits[0].Configurations["on"] !=
			splitChangesDtoFromMock.Splits[0].Configurations["on"] {
			t.Error("Marshal struct mal formed [Configurations]")
		}
	}
}

func TestImpressionDTO(t *testing.T) {
	impressionTXT := `{"keyName":"some_key","treatment":"off","time":1234567890,"changeNumber":55555555,"label":"some label","bucketingKey":"some_bucket_key"}`
	impressionDTO := &ImpressionDTO{
		KeyName:      "some_key",
		Treatment:    "off",
		Time:         1234567890,
		ChangeNumber: 55555555,
		Label:        "some label",
		BucketingKey: "some_bucket_key"}

	marshalImpression, err := impressionDTO.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if string(marshalImpression) != impressionTXT {
		t.Error("Error marshaling impression")
	}

}
