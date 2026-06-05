package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryClient uploads files to Cloudinary and returns the secure CDN URL.
type CloudinaryClient struct {
	cld    *cloudinary.Cloudinary
	folder string
}

func NewCloudinary(cloudName, apiKey, apiSecret, folder string) (*CloudinaryClient, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("cloudinary: init: %w", err)
	}
	return &CloudinaryClient{cld: cld, folder: folder}, nil
}

func (c *CloudinaryClient) Upload(data io.Reader, filename, contentType string) (string, error) {
	ctx := context.Background()

	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filepath.Base(filename), ext)
	name = strings.ReplaceAll(name, " ", "_")

	resp, err := c.cld.Upload.Upload(ctx, data, uploader.UploadParams{
		Folder:         c.folder,
		PublicID:       name,
		UniqueFilename: api.Bool(true),
		Overwrite:      api.Bool(false),
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary: upload: %w", err)
	}
	return resp.SecureURL, nil
}

// Delete destroys the asset on Cloudinary by deriving the public_id from the URL.
// Cloudinary URL format: https://res.cloudinary.com/{cloud}/{type}/upload/{version}/{folder}/{public_id}.{ext}
func (c *CloudinaryClient) Delete(publicURL string) error {
	if publicURL == "" {
		return nil
	}
	publicID := extractPublicID(publicURL)
	if publicID == "" {
		return nil
	}

	ctx := context.Background()
	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("cloudinary: delete %s: %w", publicID, err)
	}
	return nil
}

// extractPublicID derives the Cloudinary public_id (including folder, without extension)
// from a secure URL like: https://res.cloudinary.com/cloud/image/upload/v123/folder/name.jpg
func extractPublicID(secureURL string) string {
	// Find "/upload/" marker
	marker := "/upload/"
	idx := strings.Index(secureURL, marker)
	if idx == -1 {
		return ""
	}
	after := secureURL[idx+len(marker):]

	// Strip version segment if present (e.g. "v1234567890/")
	if len(after) > 0 && after[0] == 'v' {
		if slash := strings.Index(after, "/"); slash != -1 {
			after = after[slash+1:]
		}
	}

	// Remove file extension
	if dot := strings.LastIndex(after, "."); dot != -1 {
		after = after[:dot]
	}
	return after
}
