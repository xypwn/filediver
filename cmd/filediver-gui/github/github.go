package github

import (
	"fmt"
	"net/http"
	"regexp"
)

var reGHVersionedReleaseDLURL = regexp.MustCompile(`^https:\/\/github\.com\/[^\/]+\/[^\/]+\/releases\/download\/([^\/]+)\/([^\/]+)$`)

type Repository struct {
	User string
	Repo string
}

type ReleaseAsset struct {
	Repository
	Filename string
}

type ReleaseAssetInfo struct {
	ReleaseAsset
	ResolvedVersion string
	DownloadUrl     string
}

// Fetches github release artifact information.
// Leave version empty to automatically resolve latest version.
func (asset ReleaseAsset) FetchInfo(version string) (info ReleaseAssetInfo, err error) {
	info.ReleaseAsset = asset

	var url string
	if version == "" {
		url = fmt.Sprintf("https://github.com/%v/%v/releases/latest/download/%v", asset.User, asset.Repo, asset.Filename)
	} else {
		url = fmt.Sprintf("https://github.com/%v/%v/releases/download/%v/%v", asset.User, asset.Repo, version, asset.Filename)
	}

	versionFromURL := func(url string) (version string, ok bool) {
		m := reGHVersionedReleaseDLURL.FindStringSubmatch(url)
		if len(m) <= 1 {
			return "", false
		}
		return m[1], true
	}
	if ver, ok := versionFromURL(url); ok {
		info.ResolvedVersion = ver
		info.DownloadUrl = url
		return
	}

	resp, err := (&http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}).Get(url)
	if err != nil {
		return ReleaseAssetInfo{}, err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		return ReleaseAssetInfo{}, fmt.Errorf("invalid GitHub release asset URL \"%v\": expected status 302, but got %v", url, resp.Status)
	}

	loc, err := resp.Location()
	if err != nil {
		return ReleaseAssetInfo{}, err
	}
	if ver, ok := versionFromURL(loc.String()); ok {
		info.ResolvedVersion = ver
		info.DownloadUrl = loc.String()
		return
	} else {
		return ReleaseAssetInfo{}, fmt.Errorf("invalid GitHub release asset URL \"%v\": contains no version information", url)
	}
}
