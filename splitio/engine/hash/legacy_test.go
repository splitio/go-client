package hash

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
)

func TestLegacyHashOnAlphanumericData(t *testing.T) {
	inFile, _ := os.Open("../../../testdata/sample-data.jsonl")
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		arr := make([]interface{}, 4)
		json.Unmarshal(scanner.Bytes(), &arr)
		seed := uint32(arr[0].(float64))
		str := []byte(arr[1].(string))
		digest := uint32(arr[2].(float64))

		calculated := Legacy(str, seed)
		if calculated != digest {
			t.Errorf("Legacy hash calculation failed for string %s. Should be %d and was %d", str, digest, calculated)
		}
	}
}

// func TestLegacyHashOnNonAlphanumericData(t *testing.T) {
// 	inFile, _ := os.Open("../../../testdata/sample-data-non-alpha-numeric.jsonl")
// 	defer inFile.Close()
//
// 	scanner := bufio.NewScanner(inFile)
// 	scanner.Split(bufio.ScanLines)
//
// 	i := 0
// 	for scanner.Scan() && i < 10 {
// 		i++
// 		arr := make([]interface{}, 4)
// 		json.Unmarshal(scanner.Bytes(), &arr)
// 		seed := int32(arr[0].(float64))
// 		str := arr[1].(string)
// 		digest := int32(arr[2].(float64))
//
// 		calculated := Legacy(str, seed)
// 		if calculated != digest {
// 			t.Errorf("Legacy hash calculation failed for string %s. Should be %d and was %d", str, digest, calculated)
// 		}
// 	}
// }
