package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const uploadThingAPI = "https://api.uploadthing.com"

type UploadThing struct {
	Token  string
	Client *http.Client
}
type UploadedFile struct {
	Name string
	URL  string
}

func NewUploadThing(token string) *UploadThing {
	return &UploadThing{Token: token, Client: &http.Client{}}
}

// UploadFromURL downloads the remote file, then sends it through UploadThing's
// server side upload API. UploadThing remains responsible for storage/CDN.
func (u *UploadThing) UploadFromURL(ctx context.Context, source string) (UploadedFile, error) {
	if strings.TrimSpace(u.Token) == "" {
		return UploadedFile{}, fmt.Errorf("UPLOADTHING_TOKEN is required to upload new wallpapers")
	}
	parsed, err := url.Parse(source)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return UploadedFile{}, fmt.Errorf("invalid source URL %q", source)
	}
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return UploadedFile{}, err
	}
	resp, err := u.Client.Do(getReq)
	if err != nil {
		return UploadedFile{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return UploadedFile{}, fmt.Errorf("source returned %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20))
	if err != nil {
		return UploadedFile{}, err
	}
	if len(data) == 50<<20 {
		return UploadedFile{}, fmt.Errorf("source file exceeds 50 MiB")
	}
	name := path.Base(parsed.Path)
	if name == "." || name == "/" || name == "" {
		name = "wallpaper"
	}
	contentType := resp.Header.Get("Content-Type")
	if parsedType, _, _ := mime.ParseMediaType(contentType); parsedType == "" {
		contentType = "application/octet-stream"
	}
	payload, err := json.Marshal(map[string]any{"files": []any{map[string]any{"name": name, "type": contentType, "data": "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data)}}})
	if err != nil {
		return UploadedFile{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadThingAPI+"/v6/uploadFiles", bytes.NewReader(payload))
	if err != nil {
		return UploadedFile{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-uploadthing-api-key", u.Token)
	uploadResp, err := u.Client.Do(request)
	if err != nil {
		return UploadedFile{}, err
	}
	defer uploadResp.Body.Close()
	if uploadResp.StatusCode < 200 || uploadResp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(uploadResp.Body, 4096))
		return UploadedFile{}, fmt.Errorf("UploadThing returned %s: %s", uploadResp.Status, strings.TrimSpace(string(body)))
	}
	var result []struct {
		Data struct {
			Name string `json:"name"`
			URL  string `json:"ufsUrl"`
		} `json:"data"`
		Error any `json:"error"`
	}
	if err := json.NewDecoder(uploadResp.Body).Decode(&result); err != nil {
		return UploadedFile{}, err
	}
	if len(result) == 0 || result[0].Error != nil || result[0].Data.URL == "" {
		return UploadedFile{}, fmt.Errorf("UploadThing did not return a successful uploaded file")
	}
	return UploadedFile{Name: result[0].Data.Name, URL: result[0].Data.URL}, nil
}
