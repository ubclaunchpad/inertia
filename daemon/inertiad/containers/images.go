package containers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/blang/semver"
)

type imageTagDescription struct {
	Creator int `json:"creator"`
	ID      int `json:"id"`
	Images  []struct {
		Architecture string `json:"architecture"`
		Features     string `json:"features"`
		Digest       string `json:"digest"`
		Os           string `json:"os"`
		OsFeatures   string `json:"os_features"`
		Size         int    `json:"size"`
	} `json:"images"`
	LastUpdated         time.Time `json:"last_updated"`
	LastUpdater         int       `json:"last_updater"`
	LastUpdaterUsername string    `json:"last_updater_username"`
	Name                string    `json:"name"`
	Repository          int       `json:"repository"`
	FullSize            int       `json:"full_size"`
	V2                  bool      `json:"v2"`
}

type imageTagsResult struct {
	Count    int                   `json:"count"`
	Next     string                `json:"next"`
	Previous interface{}           `json:"previous"`
	Results  []imageTagDescription `json:"results"`
}

func (res *imageTagsResult) getLatest(min *semver.Version) (*semver.Version, error) {
	var latest semver.Version
	if min != nil {
		latest = *min
	}

	for _, tag := range res.Results {
		v, err := semver.Parse(strings.TrimPrefix(tag.Name, "v"))
		// ignore invalid tags - these are probably previews
		if err == nil {
			if v.Pre == nil && v.GT(latest) {
				latest = v
			}
		}
	}

	if latest.String() == "0.0.0" {
		return nil, fmt.Errorf("no new versions found")
	}

	return &latest, latest.Validate()
}

// GetLatestImageTag retrieves the most recent valid semver tag of an image
func GetLatestImageTag(ctx context.Context, image string, min *semver.Version) (*semver.Version, error) {
	target := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags/", image)
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var res imageTagsResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.getLatest(min)
}
