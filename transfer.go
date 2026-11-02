package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const RootFolderID = "0"
const DoraWallpaperFolderID = "432219983035"
const DoraWallpaperFolderName = "DoraWallpaper"

type NetworkClient struct {
	Uploader   *s3.Client
	Downloader *http.Client
}
type UploadedFile struct {
	Name string
	URL  string
}

const CloudFlareR2BucketName = "dora-wallpaper"

func NewNetworkClient(ctx context.Context) *NetworkClient {

	cloudFlareAccountID := os.Getenv(("CLOUDFLARE_ACCOUNT_ID"))
	cloudFlareAccessKeyID := os.Getenv("CLOUDFLARE_ACCESS_KEY_ID")
	cloudFlareAccessKeySecret := os.Getenv("CLOUDFLARE_ACCESS_KEY_SECRET")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cloudFlareAccessKeyID, cloudFlareAccessKeySecret, "")),
		config.WithRegion("auto"), // Required by SDK but not used by R2
	)

	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cloudFlareAccountID))
	})

	if client == nil {
		log.Fatalf("Failed to create S3 client")
	}

	uploader := client
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

	log.Printf("Downloading: %s", source)

	resp, err := c.Downloader.Do(getReq)
	if err != nil {
		log.Printf("Download error: %v", err)
		return UploadedFile{}, err
	}
	defer resp.Body.Close()

	// Extract filename from path
	filename := path.Base(parsed.Path)

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return UploadedFile{}, err
	}

	if len(fileData) == 0 {
		return UploadedFile{}, fmt.Errorf("downloaded file is empty for %s", source)
	}

	_, err = c.Uploader.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(CloudFlareR2BucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(fileData),
		ContentType: aws.String("image/jpeg"),
	})

	if err != nil {
		log.Printf("Upload failed for file %s: %v", filename, err)
		return UploadedFile{}, fmt.Errorf("Upload failed for file %s", filename)
	}

	log.Printf("Uploaded file to R2: %s", filename)

	return UploadedFile{
		Name: filename,
		URL:  fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", os.Getenv("CLOUDFLARE_ACCOUNT_ID"), filename),
	}, nil
}

func (c *NetworkClient) ListFilesInBucket(ctx context.Context) ([]types.Object, error) {
	listObjectsOutput, err := c.Uploader.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(CloudFlareR2BucketName),
	})

	if err != nil {
		return nil, err
	}

	for _, obj := range listObjectsOutput.Contents {
		log.Printf("Found object: %s", *obj.Key)
	}

	return listObjectsOutput.Contents, nil
}
