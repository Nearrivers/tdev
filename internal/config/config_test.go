package config_test

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/Nearrivers/tdev/internal/config"
)

// ── memFS : filesystem en mémoire ────────────────────────────────────────────

type memFS struct {
	files    map[string][]byte
	mkdirErr error // simule une erreur sur MkdirAll si non nil
}

func newMemFS() *memFS {
	return &memFS{files: make(map[string][]byte)}
}

func (m *memFS) Open(name string) (fs.File, error) {
	data, ok := m.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	// fstest.MapFS gère correctement fs.File (Stat, Read, Close...)
	mapFS := fstest.MapFS{
		name: &fstest.MapFile{Data: data},
	}
	return mapFS.Open(name)
}

func (m *memFS) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	return data, nil
}

func (m *memFS) WriteFile(path string, data []byte, _ fs.FileMode) error {
	m.files[path] = data
	return nil
}

func (m *memFS) MkdirAll(_ string, _ fs.FileMode) error {
	return m.mkdirErr
}

const testPath = "/fake/tdev/tdev.toml"

func newTestStore(mfs *memFS) *config.Store {
	return config.NewWithFS(mfs, testPath)
}

func TestLoad_FileNotExist(t *testing.T) {
	store := newTestStore(newMemFS())

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("load gave unexpected error: %v", err)
	}
	if len(cfg.Projects) != 0 {
		t.Errorf("Load() = %d projets, attendu 0", len(cfg.Projects))
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	mfs := newMemFS()
	mfs.files[testPath] = []byte(`not = [valid toml`)

	_, err := newTestStore(mfs).Load()
	if err == nil {
		t.Fatal("Load() aurait dû retourner une erreur pour un TOML invalide")
	}
}

func TestLoad_WithProjects(t *testing.T) {
	mfs := newMemFS()
	mfs.files[testPath] = []byte(`
[[projects]]
name = "monapp"
front = "/home/user/monapp-front"
back = "/home/user/monapp-back"

[[projects]]
name = "autreapp"
front = "/home/user/autreapp-front"
back = "/home/user/autreapp-back"
`)

	cfg, err := newTestStore(mfs).Load()
	if err != nil {
		t.Fatalf("Got an error while trying to load config but didn't exepct one. Got %v", err)
	}

	got := len(cfg.Projects)
	want := 2
	if got != want {
		t.Fatalf("Wrong number of projects found. Got %v, want %v", got, want)
	}

	cases := []config.Project{
		{Name: "monapp", Front: "/home/user/monapp-front", Back: "/home/user/monapp-back"},
		{Name: "autreapp", Front: "/home/user/autreapp-front", Back: "/home/user/autreapp-back"},
	}

	for i, exp := range cases {
		p := cfg.Projects[i]

		if p.Name != exp.Name || p.Front != exp.Front || p.Back != exp.Back {
			t.Errorf("Wrong project found at index %d, got %+v, want %+v", i, p, exp)
		}
	}
}

func TestAdd(t *testing.T) {
	mfs := newMemFS()
	mfs.files[testPath] = []byte{}

	store := newTestStore(mfs)

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Got an error while trying to load config but didn't expect one, got %v", err)
	}

	project := config.Project{
		Name:  "test",
		Front: "/home/user/mon-app",
		Back:  "/home/user/mon-back",
	}

	err = store.Add(cfg, project)
	if err != nil {
		t.Fatalf("got an error but didn't expect one, got %v", err)
	}

	got := len(cfg.Projects)
	want := 1
	if got != want {
		t.Fatalf("wrong number of projects found, got %d, want %d", got, want)
	}

	gotProject := cfg.Projects[0]
	if gotProject.Name != project.Name || gotProject.Front != project.Front || gotProject.Back != project.Back {
		t.Fatalf("wrong project found, found %+v, want %+v", gotProject, project)
	}
}

func TestAdd_AlreadyExists(t *testing.T) {
	mfs := newMemFS()
	mfs.files[testPath] = []byte(`
[[projects]]
name = "monapp"
front = "/home/user/monapp-front"
back = "/home/user/monapp-back"
`)

	store := newTestStore(mfs)

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Got an error while trying to load config but didn't expect one, got %v", err)
	}

	project := config.Project{
		Name:  "monapp",
		Front: "/home/user/mon-app",
		Back:  "/home/user/mon-back",
	}

	err = store.Add(cfg, project)
	if err != nil && !errors.Is(err, config.ErrProjectAlreadyExists) {
		t.Fatalf("got an error but didn't expect one, got %v", err)
	}

	got := len(cfg.Projects)
	want := 1
	if got != want {
		t.Fatalf("wrong number of projects found, got %d, want %d", got, want)
	}

	gotProject := cfg.Projects[0]
	if gotProject.Name != project.Name {
		t.Fatalf("wrong project found, found %+v, want %+v", gotProject, project)
	}
}

func TestRemove(t *testing.T) {
	mfs := newMemFS()
	mfs.files[testPath] = []byte(`
[[projects]]
name = "monapp"
front = "/home/user/monapp-front"
back = "/home/user/monapp-back"
`)

	store := newTestStore(mfs)

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Got an error while trying to load config but didn't expect one, got %v", err)
	}

	projectToRemove := "monapp"

	err = store.Remove(cfg, projectToRemove)
	if err != nil {
		t.Fatalf("got an error while removing but didn't expect one: got %v", err)
	}

	got := len(cfg.Projects)
	want := 0
	if got != want {
		t.Errorf("wrong number of projects found, got %d, want %d", got, want)
	}
}

func TestRemove_Inexistant(t *testing.T) {
	mfs := newMemFS()
	mfs.files[testPath] = []byte(`
[[projects]]
name = "monapp"
front = "/home/user/monapp-front"
back = "/home/user/monapp-back"
`)

	store := newTestStore(mfs)

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Got an error while trying to load config but didn't expect one, got %v", err)
	}

	projectToRemove := "inexistant"

	err = store.Remove(cfg, projectToRemove)
	if err != nil {
		t.Fatalf("got an error while removing but didn't expect one: got %v", err)
	}

	got := len(cfg.Projects)
	want := 1
	if got != want {
		t.Errorf("wrong number of projects found, got %d, want %d", got, want)
	}
}
