package local

import (
	"fmt"
	"github.com/splitio/go-client/splitio/service/dtos"
	"io/ioutil"
	"strings"
)

const (
	// SplitFileFormatClassic represents the file format of the standard split definition file <feature treatment>
	SplitFileFormatClassic = iota
	// SplitFileFormatJSON represents the file format of a JSON representation of split dtos
	SplitFileFormatJSON
)

// FileSplitFetcher struct fetches splits from a file
type FileSplitFetcher struct {
	splitFile        string
	fileFormat       int
	lastChangeNumber int64
}

// NewFileSplitFetcher returns a new instance of LocalFileSplitFetcher
func NewFileSplitFetcher(splitFile string, fileFormat int) *FileSplitFetcher {
	return &FileSplitFetcher{
		splitFile:  splitFile,
		fileFormat: fileFormat,
	}
}

func parseSplitsClassic(data string) []dtos.SplitDTO {
	splits := make([]dtos.SplitDTO, 0)
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		words := strings.Fields(line)
		if len(words) < 2 || len(words[0]) < 1 || words[0][0] == '#' {
			// Skip the line if it has less than two words, the words are empty strings or
			// it begins with '#' character
			continue
		}
		splitName := words[0]
		treatment := words[1]
		splits = append(splits, dtos.SplitDTO{
			Name: splitName,
			Conditions: []dtos.ConditionDTO{
				dtos.ConditionDTO{
					Label: "LOCAL",
					MatcherGroup: dtos.MatcherGroupDTO{
						Combiner: "AND",
						Matchers: []dtos.MatcherDTO{
							dtos.MatcherDTO{
								MatcherType: "ALL_KEYS",
								Negate:      false,
							},
						},
					},
					Partitions: []dtos.PartitionDTO{
						dtos.PartitionDTO{
							Size:      100,
							Treatment: treatment,
						},
						dtos.PartitionDTO{
							Size:      0,
							Treatment: "_",
						},
					},
				},
			},
			Status:           "ACTIVE",
			DefaultTreatment: treatment,
		})
	}
	return splits
}

// Fetch parses the file and returns the appropriate structures
func (s *FileSplitFetcher) Fetch(changeNumber int64) (*dtos.SplitChangesDTO, error) {
	fileContents, err := ioutil.ReadFile(s.splitFile)
	if err != nil {
		return nil, err
	}

	var splits []dtos.SplitDTO
	var till int64
	since := s.lastChangeNumber
	if s.lastChangeNumber != 0 {
		//The first time we should return since == till
		till = since + 1
	}

	data := string(fileContents)
	switch s.fileFormat {
	case SplitFileFormatClassic:
		splits = parseSplitsClassic(data)
	case SplitFileFormatJSON:
		fallthrough
	default:
		return nil, fmt.Errorf("Unsupported file format")

	}

	s.lastChangeNumber++
	return &dtos.SplitChangesDTO{
		Splits: splits,
		Since:  since,
		Till:   till,
	}, nil
}
