package codemap

import (
	"errors"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// CodeMap maps component IDs to their source file paths.
type CodeMap map[string][]string

// ParseCodeMap reads and parses a code-map.yaml file.
// Returns an empty map if the file does not exist.
func ParseCodeMap(path string) (CodeMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return CodeMap{}, nil
		}
		return nil, err
	}

	cm := CodeMap{}
	if len(data) == 0 {
		return cm, nil
	}

	if err := yaml.Unmarshal(data, &cm); err != nil {
		return nil, err
	}

	if cm == nil {
		return CodeMap{}, nil
	}

	return cm, nil
}
