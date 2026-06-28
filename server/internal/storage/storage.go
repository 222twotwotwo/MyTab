package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	ManifestFile = "manifest.json"
	MaxFiles     = 30
	MaxFileBytes = 5 << 20
)

var (
	ErrNotFound       = errors.New("project not found")
	ErrFileNotFound   = errors.New("file not found")
	ErrInvalidUpload  = errors.New("invalid upload")
	allowedExtensions = map[string]string{
		".html":  "html",
		".htm":   "html",
		".css":   "css",
		".js":    "javascript",
		".mjs":   "javascript",
		".json":  "json",
		".txt":   "text",
		".png":   "image",
		".jpg":   "image",
		".jpeg":  "image",
		".gif":   "image",
		".webp":  "image",
		".svg":   "image",
		".ico":   "image",
		".woff":  "font",
		".woff2": "font",
		".ttf":   "font",
		".otf":   "font",
		".mp3":   "media",
		".mp4":   "media",
		".webm":  "media",
	}
)

type Store struct {
	root string
}

type ProjectManifest struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	CreatedAt time.Time    `json:"createdAt"`
	EntryFile string       `json:"entryFile"`
	Files     []StoredFile `json:"files"`
}

type StoredFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Kind string `json:"kind"`
}

func New(root string) (*Store, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("storage root is required")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &Store{root: root}, nil
}

func (s *Store) List() ([]ProjectManifest, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, err
	}

	projects := make([]ProjectManifest, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifest, err := s.readManifest(entry.Name())
		if err != nil {
			continue
		}
		projects = append(projects, *manifest)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].CreatedAt.After(projects[j].CreatedAt)
	})
	return projects, nil
}

func (s *Store) Get(id string) (*ProjectManifest, error) {
	if !isSafeID(id) {
		return nil, ErrNotFound
	}
	return s.readManifest(id)
}

func (s *Store) Create(name string, headers []*multipart.FileHeader, paths []string) (*ProjectManifest, error) {
	if len(headers) == 0 {
		return nil, fmt.Errorf("%w: upload at least one file", ErrInvalidUpload)
	}
	if len(headers) > MaxFiles {
		return nil, fmt.Errorf("%w: upload at most %d files", ErrInvalidUpload, MaxFiles)
	}

	id, err := newID()
	if err != nil {
		return nil, err
	}

	projectDir := filepath.Join(s.root, id)
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		return nil, err
	}

	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(projectDir)
		}
	}()

	seen := make(map[string]struct{}, len(headers))
	files := make([]StoredFile, 0, len(headers))
	entryFile := ""

	for index, header := range headers {
		originalName := header.Filename
		if index < len(paths) && strings.TrimSpace(paths[index]) != "" {
			originalName = paths[index]
		}

		fileName, kind, err := validateFileName(originalName)
		if err != nil {
			return nil, err
		}
		lookup := strings.ToLower(fileName)
		if _, exists := seen[lookup]; exists {
			return nil, fmt.Errorf("%w: duplicate file %q", ErrInvalidUpload, fileName)
		}
		seen[lookup] = struct{}{}

		if header.Size > MaxFileBytes {
			return nil, fmt.Errorf("%w: %q is larger than %d bytes", ErrInvalidUpload, fileName, MaxFileBytes)
		}

		size, err := saveMultipartFile(filepath.Join(projectDir, fileName), header)
		if err != nil {
			return nil, err
		}

		files = append(files, StoredFile{Name: fileName, Size: size, Kind: kind})
		if kind == "html" && (entryFile == "" || strings.EqualFold(fileName, "index.html") || strings.EqualFold(fileName, "index.htm")) {
			entryFile = fileName
		}
	}

	if entryFile == "" {
		return nil, fmt.Errorf("%w: include at least one html file", ErrInvalidUpload)
	}

	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = "Untitled Project"
	}

	manifest := &ProjectManifest{
		ID:        id,
		Name:      displayName,
		CreatedAt: time.Now().UTC(),
		EntryFile: entryFile,
		Files:     files,
	}

	if err := s.writeManifest(manifest); err != nil {
		return nil, err
	}

	success = true
	return manifest, nil
}

func (s *Store) ResolveFile(id string, requested string) (string, string, error) {
	manifest, err := s.Get(id)
	if err != nil {
		return "", "", err
	}

	fileName, err := cleanRelativePath(strings.TrimPrefix(requested, "/"))
	if err != nil {
		return "", "", ErrFileNotFound
	}
	if fileName == "" {
		fileName = manifest.EntryFile
	}

	for _, file := range manifest.Files {
		if file.Name == fileName {
			return filepath.Join(s.root, id, file.Name), file.Kind, nil
		}
	}
	return "", "", ErrFileNotFound
}

func (s *Store) readManifest(id string) (*ProjectManifest, error) {
	if !isSafeID(id) {
		return nil, ErrNotFound
	}

	file, err := os.Open(filepath.Join(s.root, id, ManifestFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var manifest ProjectManifest
	if err := json.NewDecoder(file).Decode(&manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func (s *Store) writeManifest(manifest *ProjectManifest) error {
	file, err := os.Create(filepath.Join(s.root, manifest.ID, ManifestFile))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

func saveMultipartFile(destination string, header *multipart.FileHeader) (int64, error) {
	src, err := header.Open()
	if err != nil {
		return 0, err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return 0, err
	}

	dst, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	size, err := io.CopyN(dst, src, MaxFileBytes+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}
	if size > MaxFileBytes {
		return 0, fmt.Errorf("%w: %q is larger than %d bytes", ErrInvalidUpload, header.Filename, MaxFileBytes)
	}
	return size, nil
}

func validateFileName(original string) (string, string, error) {
	name, err := cleanRelativePath(original)
	if err != nil {
		return "", "", fmt.Errorf("%w: invalid file name", ErrInvalidUpload)
	}
	if name == "" {
		return "", "", fmt.Errorf("%w: invalid file name", ErrInvalidUpload)
	}
	if strings.ContainsAny(name, `<>:"\|?*`) {
		return "", "", fmt.Errorf("%w: unsupported file name %q", ErrInvalidUpload, name)
	}

	kind, ok := allowedExtensions[strings.ToLower(path.Ext(name))]
	if !ok {
		return "", "", fmt.Errorf("%w: %q is not a supported static resource", ErrInvalidUpload, name)
	}
	return name, kind, nil
}

func cleanRelativePath(original string) (string, error) {
	name := strings.TrimSpace(strings.ReplaceAll(original, "\\", "/"))
	name = strings.TrimPrefix(name, "/")
	if name == "" {
		return "", nil
	}
	if strings.HasPrefix(name, "~") || strings.Contains(name, ":") {
		return "", ErrInvalidUpload
	}

	cleaned := path.Clean(name)
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return "", ErrInvalidUpload
	}
	if strings.Contains(cleaned, "/../") {
		return "", ErrInvalidUpload
	}
	for _, part := range strings.Split(cleaned, "/") {
		if part == "" || part == "." || part == ".." {
			return "", ErrInvalidUpload
		}
	}
	return cleaned, nil
}

func isSafeID(id string) bool {
	if len(id) != 24 {
		return false
	}
	for _, char := range id {
		if (char < 'a' || char > 'f') && (char < '0' || char > '9') {
			return false
		}
	}
	return true
}

func newID() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
