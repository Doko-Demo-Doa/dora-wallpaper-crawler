package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/asalih/go-box-sdk/box"
	"github.com/asalih/go-box-sdk/managers"
	"github.com/asalih/go-box-sdk/networking"
)

const RootFolderID = "0"
const DoraWallpaperFolderID = "432219983035"
const DoraWallpaperFolderName = "DoraWallpaper"

type NetworkClient struct {
	Uploader   *box.BoxClient
	Downloader *http.Client
}
type UploadedFile struct {
	Name string
	URL  string
}

func NewNetworkClient() *NetworkClient {
	var auth networking.Authentication

	if developerToken := os.Getenv("BOX_DEVELOPER_TOKEN"); developerToken != "" {
		auth = box.NewBoxDeveloperTokenAuth(developerToken, box.DeveloperTokenConfig{
			ClientID:     os.Getenv("BOX_CLIENT_ID"),
			ClientSecret: os.Getenv("BOX_CLIENT_SECRET"),
		})
	}

	uploader := box.NewBoxClient(auth, nil)
	downloader := &http.Client{
		Timeout: 5 * time.Minute,
	}

	return &NetworkClient{Uploader: uploader, Downloader: downloader}
}

// UploadFromURL downloads the remote file and uploads a copy into the
// DoraWallpaper folder in Box.
func (c *NetworkClient) UploadFromURL(ctx context.Context, source string) (UploadedFile, error) {
	parsed, err := url.Parse(source)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return UploadedFile{}, fmt.Errorf("Invalid source URL %q", source)
	}

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return UploadedFile{}, err
	}

	resp, err := c.Downloader.Do(getReq)
	if err != nil {
		return UploadedFile{}, err
	}
	defer resp.Body.Close()

	// Extract filename from path
	filename := path.Base(parsed.Path)

	boxUploadedFiles, err := c.Uploader.Uploads.UploadFile(ctx, &managers.UploadFileRequestBody{
		Attributes: managers.UploadFileAttributes{
			Name:   filename,
			Parent: managers.UploadFileAttributesParent{ID: DoraWallpaperFolderID},
		},
		File:         resp.Body,
		FileFileName: filename,
	}, nil)

	if err != nil {
		log.Printf("Upload failed for file %s: %v", filename, err)
		return UploadedFile{}, fmt.Errorf("Upload failed for file %s", filename)
	}

	if len(boxUploadedFiles.Entries) == 0 {
		return UploadedFile{}, fmt.Errorf("upload %s: Box returned no file entries", filename)
	}

	uploaded := boxUploadedFiles.Entries[0]
	log.Printf("Uploaded file to Box.net: %s (id %s)", uploaded.Name, uploaded.ID)

	return UploadedFile{
		Name: uploaded.Name,
		URL:  uploaded.ID,
	}, nil
}
