package storage

import (
	"fmt"
	"testing"

	"github.com/splitio/go-client/splitio/service/dtos"
)

func TestMMSplitStorage(t *testing.T) {
	splitStorage := NewMMSplitStorage()
	splits := make([]dtos.SplitDTO, 10)
	for index := 0; index < 10; index++ {
		splits = append(splits, dtos.SplitDTO{
			Name: fmt.Sprintf("SomeSplit_%d", index),
			Algo: index,
		})
	}

	splitStorage.PutMany(&splits)
	for index := 0; index < 10; index++ {
		splitName := fmt.Sprintf("SomeSplit_%d", index)
		split, found := splitStorage.Get(splitName)
		if !found || split.Name != splitName || split.Algo != index {
			t.Error("Split not returned as expected")
		}
	}

	_, found := splitStorage.Get("nonexistant_split")
	if found {
		t.Error("Nil expected but split returned")
	}

	splitStorage.Remove("SomeSplit_7")
	for index := 0; index < 10; index++ {
		splitName := fmt.Sprintf("SomeSplit_%d", index)
		split, found := splitStorage.Get(splitName)
		if index != 7 {
			if !found || split.Name != splitName || split.Algo != index {
				t.Error("Split not have been removed or modified")
			}
		} else {
			if found {
				t.Error("Split Should have been removed and is present")
			}
		}
	}
}
