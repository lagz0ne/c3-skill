package marketplace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Name    string    `yaml:"name"`
	URL     string    `yaml:"url"`
	Fetched time.Time `yaml:"fetched"`
}

type sourcesFile struct {
	Sources []Source `yaml:"sources"`
}

type Registry struct {
	baseDir string
}

func NewRegistry(baseDir string) *Registry {
	return &Registry{baseDir: baseDir}
}

func DefaultBaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".c3", "marketplace")
}

func (r *Registry) CacheDir(name string) string {
	return filepath.Join(r.baseDir, name)
}

func (r *Registry) sourcesPath() string {
	return filepath.Join(r.baseDir, "sources.yaml")
}

func (r *Registry) List() ([]Source, error) {
	data, err := os.ReadFile(r.sourcesPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var sf sourcesFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("parse sources.yaml: %w", err)
	}
	return sf.Sources, nil
}

func (r *Registry) Get(name string) (*Source, error) {
	sources, err := r.List()
	if err != nil {
		return nil, err
	}
	for _, s := range sources {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("source %q not found", name)
}

func (r *Registry) Add(src Source) error {
	if strings.TrimSpace(src.Name) == "" {
		return fmt.Errorf("source name is required")
	}
	sources, err := r.List()
	if err != nil {
		return err
	}
	for _, s := range sources {
		if s.Name == src.Name {
			return fmt.Errorf("source %q already exists (use remove + add to change URL)", src.Name)
		}
	}
	src.Fetched = time.Now().UTC()
	sources = append(sources, src)
	return r.save(sources)
}

func (r *Registry) Remove(name string) error {
	sources, err := r.List()
	if err != nil {
		return err
	}
	found := false
	var filtered []Source
	for _, s := range sources {
		if s.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, s)
	}
	if !found {
		return fmt.Errorf("source %q not found", name)
	}
	return r.save(filtered)
}

func (r *Registry) UpdateFetched(name string) error {
	sources, err := r.List()
	if err != nil {
		return err
	}
	for i, s := range sources {
		if s.Name == name {
			sources[i].Fetched = time.Now().UTC()
			return r.save(sources)
		}
	}
	return fmt.Errorf("source %q not found", name)
}

func (r *Registry) save(sources []Source) error {
	if err := os.MkdirAll(r.baseDir, 0755); err != nil {
		return err
	}
	sf := sourcesFile{Sources: sources}
	data, err := yaml.Marshal(&sf)
	if err != nil {
		return err
	}
	return os.WriteFile(r.sourcesPath(), data, 0644)
}
