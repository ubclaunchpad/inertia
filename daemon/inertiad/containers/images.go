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

type dockerHubImageTagDescription struct {
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

type dockerHubImageTagsResult struct {
	Count    int                            `json:"count"`
	Next     string                         `json:"next"`
	Previous interface{}                    `json:"previous"`
	Results  []dockerHubImageTagDescription `json:"results"`
}

func (res *dockerHubImageTagsResult) getVersions() versions {
	versions := []string{}
	for _, tag := range res.Results {
		versions = append(versions, tag.Name)
	}
	return versions
}

type versions []string

func (v versions) getLatest(min *semver.Version) (*semver.Version, error) {
	var latest semver.Version
	if min != nil {
		latest = *min
	}

	for _, version := range v {
		v, err := semver.Parse(strings.TrimPrefix(version, "v"))
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

// Image describes a Docker image. An empty registry defaults to DockerHub
type Image struct {
	Registry   string
	Repository string
}

// GetLatestImageTag retrieves the most recent valid semver tag of an image
func GetLatestImageTag(ctx context.Context, image Image, min *semver.Version) (*semver.Version, error) {
	var v versions
	switch image.Registry {
	case "ghcr.io":
		// ghcr.io has no REST API yet, so we just check releases
		target := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", image.Repository)
		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		var res struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return nil, err
		}
		v = []string{res.TagName}

	default:
		// assume dockerhub
		target := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags/", image.Repository)
		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		var res dockerHubImageTagsResult
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return nil, err
		}
		v = res.getVersions()
	}

	return v.getLatest(min)
}
