package icon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

func iconScore(icon Icon, terms []string) int {
	name, title := strings.ToLower(icon.Name), strings.ToLower(icon.Title)
	if len(terms) > 1 && (name == strings.Join(terms, "-") || title == strings.Join(terms, " ")) {
		return 1000
	}
	score := 0
	for _, term := range terms {
		switch {
		case name == term || title == term:
			score += 100
		case strings.HasPrefix(name, term) || strings.HasPrefix(title, term):
			score += 50
		case strings.Contains(name, term) || strings.Contains(title, term):
			score += 20
		}
		for _, tag := range icon.Tags {
			tag = strings.ToLower(tag)
			if tag == term {
				score += 10
				break
			}
			if strings.Contains(tag, term) {
				score += 5
				break
			}
		}
	}
	return score
}

const (
	lucideTagsURL      = "https://unpkg.com/lucide-static@latest/tags.json"
	simpleIconsDataURL = "https://unpkg.com/simple-icons@latest/data/simple-icons.json"
)

var iconsHTTPClient = &http.Client{Timeout: time.Minute}

// SeedIcons fills the icons collection when it is empty.
func SeedIcons(ctx context.Context) {
	count, err := CountIcons(ctx)
	if err != nil {
		log.Println("Icons seed: count failed:", err)
		return
	}
	if count > 0 {
		return
	}
	RefreshIcons(ctx)
}

// RefreshIcons re-downloads every icon source. A source that fails to download
// keeps its previously stored icons.
func RefreshIcons(ctx context.Context) {
	sources := map[string]func(context.Context) ([]Icon, error){
		IconSourceLucide:      fetchLucideIcons,
		IconSourceSimpleIcons: fetchSimpleIcons,
	}
	for source, fetch := range sources {
		icons, err := fetch(ctx)
		if err == nil && len(icons) == 0 {
			err = fmt.Errorf("no icons returned")
		}
		if err == nil {
			err = ReplaceIcons(ctx, source, icons)
		}
		if err != nil {
			log.Printf("Icons refresh: %s failed: %v", source, err)
			continue
		}
		log.Printf("Icons refresh: %s stored %d icons", source, len(icons))
	}
}

func fetchLucideIcons(ctx context.Context) ([]Icon, error) {
	var tags map[string][]string
	if err := fetchJSON(ctx, lucideTagsURL, &tags); err != nil {
		return nil, err
	}

	icons := make([]Icon, 0, len(tags))
	for name, iconTags := range tags {
		if iconTags == nil {
			iconTags = []string{}
		}
		icons = append(icons, Icon{
			Id:     IconSourceLucide + ":" + name,
			Name:   name,
			Title:  lucideTitle(name),
			Source: IconSourceLucide,
			Tags:   iconTags,
		})
	}
	return icons, nil
}

func fetchSimpleIcons(ctx context.Context) ([]Icon, error) {
	var data []struct {
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Hex     string `json:"hex"`
		Aliases struct {
			Aka []string `json:"aka"`
		} `json:"aliases"`
	}
	if err := fetchJSON(ctx, simpleIconsDataURL, &data); err != nil {
		return nil, err
	}

	icons := make([]Icon, 0, len(data))
	for _, entry := range data {
		if entry.Slug == "" {
			continue
		}
		tags := entry.Aliases.Aka
		if tags == nil {
			tags = []string{}
		}
		icons = append(icons, Icon{
			Id:     IconSourceSimpleIcons + ":" + entry.Slug,
			Name:   entry.Slug,
			Title:  entry.Title,
			Source: IconSourceSimpleIcons,
			Tags:   tags,
			Hex:    entry.Hex,
		})
	}
	return icons, nil
}

func fetchJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := iconsHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

// lucideTitle turns "a-arrow-down" into "A Arrow Down".
func lucideTitle(name string) string {
	words := strings.Split(name, "-")
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// ScheduleIconRefresh seeds the icons collection if it is empty, then refreshes it every Sunday at 00:00 UTC.
func ScheduleIconRefresh() {
	go SeedIcons(context.Background())

	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Week().Sunday().At("00:00").Do(func() {
		RefreshIcons(context.Background())
	})
	s.StartAsync()
}

func ListIcons(ctx context.Context, filter *IconFilter) ([]Icon, error) {
	return GetAllIcons(ctx, filter)
}

func FindIcons(ctx context.Context, search *IconSearch) (*IconSearchResult, error) {
	return SearchIcons(ctx, search)
}
