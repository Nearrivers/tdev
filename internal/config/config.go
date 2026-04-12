// Package config
package config

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
)

var ErrProjectAlreadyExists = errors.New("project already exists. This error is for testing purposes")

type Project struct {
	Name  string `toml:"name"`
	Front string `toml:"front"`
	Back  string `toml:"back"`
}

type Config struct {
	Projects []Project `toml:"projects"`
}

type WriteFS interface {
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
}

type FileSystem interface {
	fs.ReadFileFS
	WriteFS
}

type osFS struct{}

func (osFS) Open(name string) (fs.File, error) { return os.Open(name) }

func (osFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

func (osFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (osFS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

type Store struct {
	fs       FileSystem
	filePath string
}

func New() (*Store, error) {
	base, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(base, ".config", "tdev", "projects.toml")
	return &Store{fs: osFS{}, filePath: path}, nil
}

func newWithFS(filesystem FileSystem, filePath string) *Store {
	return &Store{fs: filesystem, filePath: filePath}
}

func (s *Store) Load() (*Config, error) {
	data, err := s.fs.ReadFile(s.filePath)
	if errors.Is(err, fs.ErrNotExist) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Store) save(cfg *Config) error {
	dir := filepath.Dir(s.filePath)
	if err := s.fs.MkdirAll(dir, 0755); err != nil {
		return err
	}

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)

	go func() {
		enc := toml.NewEncoder(pw)
		errCh <- enc.Encode(cfg)
		pw.Close()
	}()

	data, err := io.ReadAll(pr)
	if err != nil {
		return err
	}
	if err := <-errCh; err != nil {
		return err
	}

	return s.fs.WriteFile(s.filePath, data, 0644)
}

func (s *Store) Add(c *Config, p Project) error {
	if slices.ContainsFunc(c.Projects, func(ep Project) bool {
		return ep.Name == p.Name
	}) {
		return ErrProjectAlreadyExists
	}

	c.Projects = append(c.Projects, p)
	return s.save(c)
}

func (s *Store) Remove(c *Config, name string) error {
	ps := slices.DeleteFunc(c.Projects, func(p Project) bool {
		return p.Name == name
	})

	c.Projects = ps

	return s.save(c)
}

func (s *Store) Get(c *Config, name string) (*Project, bool) {
	for _, project := range c.Projects {
		if project.Name == name {
			return &project, true
		}
	}

	return &Project{}, false
}
