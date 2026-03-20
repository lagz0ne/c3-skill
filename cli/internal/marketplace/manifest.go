package marketplace

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Name          string        `yaml:"name"`
	Description   string        `yaml:"description"`
	Tags          []string      `yaml:"tags"`
	Compatibility Compatibility `yaml:"compatibility"`
	Rules         []RuleEntry   `yaml:"rules"`
}

type Compatibility struct {
	Languages  []string `yaml:"languages"`
	Frameworks []string `yaml:"frameworks"`
}

type RuleEntry struct {
	ID       string   `yaml:"id"`
	Title    string   `yaml:"title"`
	Category string   `yaml:"category"`
	Tags     []string `yaml:"tags"`
	Summary  string   `yaml:"summary"`
}

func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse marketplace.yaml: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Manifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("marketplace.yaml: name is required")
	}
	if len(m.Rules) == 0 {
		return fmt.Errorf("marketplace.yaml: at least one rule is required")
	}
	for i, r := range m.Rules {
		if strings.TrimSpace(r.ID) == "" {
			return fmt.Errorf("marketplace.yaml: rule[%d]: id is required", i)
		}
		if !strings.HasPrefix(r.ID, "rule-") {
			return fmt.Errorf("marketplace.yaml: rule[%d]: id %q must start with \"rule-\"", i, r.ID)
		}
		if strings.TrimSpace(r.Summary) == "" {
			return fmt.Errorf("marketplace.yaml: rule[%d]: summary is required", i)
		}
	}
	return nil
}
