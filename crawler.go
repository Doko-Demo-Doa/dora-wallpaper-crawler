package main

import (
	"context"
	"log"
	"time"
)

type Crawler struct {
	Store  *Store
	Upload *UploadThing
}

func (c *Crawler) RunHourly(ctx context.Context) {
	if err := c.Crawl(ctx); err != nil {
		log.Printf("initial crawl failed: %v", err)
	}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Crawl(ctx); err != nil {
				log.Printf("crawl failed: %v", err)
			}
		}
	}
}

func (c *Crawler) Crawl(ctx context.Context) error {
	log.Printf("Starting wallpaper crawl at %s", time.Now().UTC().Format(time.RFC1123))
	desktop, mobile, err := getWallpaperLinks(ctx)
	if err != nil {
		return err
	}
	all := make([]string, 0, len(desktop)+len(mobile))
	isMobile := make(map[string]int16, len(desktop)+len(mobile))
	for _, link := range desktop {
		if _, exists := isMobile[link]; !exists {
			all = append(all, link)
			isMobile[link] = 0
		}
	}
	for _, link := range mobile {
		if _, exists := isMobile[link]; !exists {
			all = append(all, link)
		}
		// Keep the existing behavior where a URL present in both sections is
		// classified as mobile.
		isMobile[link] = 1
	}
	known, err := c.Store.ExistingLinks(ctx, all)
	if err != nil {
		return err
	}
	for _, link := range all {
		if known[link] {
			continue
		}
		uploaded, err := c.Upload.UploadFromURL(ctx, link)
		if err != nil {
			log.Printf("upload %s: %v", link, err)
			continue
		}
		if err := c.Store.Insert(ctx, Wallpaper{FileName: uploaded.Name, URL: uploaded.URL, OriginalURL: link, IsMobile: isMobile[link], CreatedAt: time.Now()}); err != nil {
			log.Printf("save %s: %v", link, err)
		}
	}
	return nil
}

func difference(links []string, known map[string]bool) []string {
	result := []string{}
	for _, link := range links {
		if !known[link] {
			result = append(result, link)
		}
	}
	return result
}
