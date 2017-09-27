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
		if index == 7 {
			if found {
				t.Error("Split Should have been removed and is present")
			}
		} else {
			if !found || split.Name != splitName || split.Algo != index {
				t.Error("Split should not have been removed or modified and it was")
			}
		}
	}
}

// func TestMMSegmentStorage(t *testing.T) {
// 	segments := [][]string{}
// 	segments[0] := []string{"1a", "1b", "1c"}
// 	segments[1] := []string{"2a", "2b", "2c"}
// 	segments[2] := []string{"3a", "3b", "3c"}
//
// 	segmentStorage := NewMMSegmentStorage()
// 	for index, segment := range segments {
// 	    segmentStorage.Put(fmt.Sprintf("segmentito_%d", index), segment)
// 	}
//
// 	for i := 0; i < 3, i++ {
// 	    segmentName := fmt.Sprintf("segmentito_%d", i)
// 	    segment, exists := segmentStorage.Get(segmentName)
// 	    if !exists {
// 		t.Error("%s should exist in storage and it doesn't.", segmentName)
// 	    }
//
//
//
// 	_, found := splitStorage.Get("nonexistant_split")
// 	if found {
// 		t.Error("Nil expected but split returned")
// 	}
//
// 	splitStorage.Remove("SomeSplit_7")
// 	for index := 0; index < 10; index++ {
// 		splitName := fmt.Sprintf("SomeSplit_%d", index)
// 		split, found := splitStorage.Get(splitName)
// 		if index == 7 {
// 			if found {
// 				t.Error("Split Should have been removed and is present")
// 			}
// 		} else {
// 			if !found || split.Name != splitName || split.Algo != index {
// 				t.Error("Split should not have been removed or modified and it was")
// 			}
// 		}
// 	}
// }
