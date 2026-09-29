package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const wallpaperPage = "https://dora-world.com/wallpaper"

var pageClient = &http.Client{Timeout: 30 * time.Second}

func getDesktopLinks(ctx context.Context) ([]string, error) {
	desktop, _, err := getWallpaperLinks(ctx)
	return desktop, err
}
func getMobileLinks(ctx context.Context) ([]string, error) {
	_, mobile, err := getWallpaperLinks(ctx)
	return mobile, err
}

func getWallpaperLinks(ctx context.Context) ([]string, []string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wallpaperPage, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := pageClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("wallpaper page returned %s", resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	base, _ := url.Parse(wallpaperPage)
	desktop := extractLinks(doc, "section.gn_box.wallpaper.nosp > .wallpaper_box", base)
	mobile := extractLinks(doc, "section.gn_box.wallpaper.nopc > .wallpaper_box", base)
	return desktop, mobile, nil
}

func extractLinks(doc *goquery.Document, selector string, base *url.URL) []string {
	links := []string{}
	seen := make(map[string]struct{})
	doc.Find(selector).Children().Each(func(_ int, child *goquery.Selection) {
		href, ok := child.ChildrenFiltered("a").Attr("href")
		if !ok || href == "" {
			return
		}
		parsed, err := url.Parse(href)
		if err != nil {
			return
		}
		resolved := base.ResolveReference(parsed).String()
		if parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
			return
		}
		if _, exists := seen[resolved]; exists {
			return
		}
		seen[resolved] = struct{}{}
		links = append(links, resolved)
	})
	return links
}
