package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client saves uploaded files to a local directory and serves them via the backend URL.
type Client struct {
	uploadDir string
	baseURL   string
}

func New(uploadDir, baseURL string) *Client {
	return &Client{
		uploadDir: uploadDir,
		baseURL:   strings.TrimRight(baseURL, "/"),
	}
}

// Upload writes data to uploadDir and returns the public URL.
func (c *Client) Upload(data io.Reader, filename, contentType string) (string, error) {
	if err := os.MkdirAll(c.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("storage: mkdir: %w", err)
	}
	objectName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitize(filename))
	dst := filepath.Join(c.uploadDir, objectName)

	f, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("storage: create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, data); err != nil {
		return "", fmt.Errorf("storage: write file: %w", err)
	}
	return c.baseURL + "/uploads/" + objectName, nil
}

// Delete removes the file referenced by publicURL. Ignores missing files.
func (c *Client) Delete(publicURL string) error {
	if publicURL == "" {
		return nil
	}
	prefix := c.baseURL + "/uploads/"
	if !strings.HasPrefix(publicURL, prefix) {
		return nil
	}
	objectName := strings.TrimPrefix(publicURL, prefix)
	path := filepath.Join(c.uploadDir, filepath.Base(objectName))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: delete: %w", err)
	}
	return nil
}

func sanitize(filename string) string {
	base := filepath.Base(filename)
	return strings.ReplaceAll(base, " ", "_")
}
