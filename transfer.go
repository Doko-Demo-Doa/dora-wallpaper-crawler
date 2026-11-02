package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/asalih/go-box-sdk/box"
	"github.com/asalih/go-box-sdk/managers"
)

const RootFolderID = "0"
const DoraWallpaperFolderID = "432219983035"
const DoraWallpaperFolderName = "DoraWallpaper"

const uploadThingAPI = "https://api.uploadthing.com"

type NetworkClient struct {
	Uploader   *box.BoxClient
	Downloader *http.Client
}
type UploadedFile struct {
	Name string
	URL  string
}

func NewNetworkClient() *NetworkClient {
	auth := box.NewBoxCcgAuth(box.CcgConfig{
		ClientID:     os.Getenv("BOX_CLIENT_ID"),
		ClientSecret: os.Getenv("BOX_CLIENT_SECRET"),
		EnterpriseID: "",
	})

	client := box.NewBoxClient(auth, nil)

	return &NetworkClient{Uploader: client}
}

// UploadFromURL downloads the remote file, then sends it through UploadThing's
// server side upload API. UploadThing remains responsible for storage/CDN.
func (c *NetworkClient) UploadFromURL(ctx context.Context, source string) (UploadedFile, error) {
	parsed, err := url.Parse(source)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return UploadedFile{}, fmt.Errorf("Invalid source URL %q", source)
	}

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return UploadedFile{}, nil
	}

	resp, err := c.Downloader.Do(getReq)
	if err != nil {
		return UploadedFile{}, nil
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
		return UploadedFile{}, fmt.Errorf("Upload failed for file %s", filename)
	}

	for uFile := range boxUploadedFiles.Entries {
		if uFile != 0 {
			fmt.Printf("Uploaded file to Box.net: %s", filename)
			return UploadedFile{
				Name: filename,
				URL:  source,
			}, nil
		}
	}

	return UploadedFile{}, nil

}
