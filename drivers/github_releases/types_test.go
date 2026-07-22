package github_releases

import "testing"

func TestReleaseHelpersHandleEmptyRelease(t *testing.T) {
	if got := releaseToFiles("/repo", nil); got != nil {
		t.Fatalf("releaseToFiles(nil) = %#v, want nil", got)
	}
	if got := releaseSize(nil); got != 0 {
		t.Fatalf("releaseSize(nil) = %d, want 0", got)
	}
	if got := releaseAssetsByTag("/repo", "v1", nil); got != nil {
		t.Fatalf("releaseAssetsByTag on empty releases = %#v, want nil", got)
	}
}

func TestReleaseHelpersBuildStablePaths(t *testing.T) {
	release := Release{TagName: "v1.2.3", CreatedAt: "2026-01-01T00:00:00Z", PublishedAt: "2026-01-02T00:00:00Z", Assets: []Asset{{Name: "app.zip", Size: 42, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-02T00:00:00Z", BrowserDownloadUrl: "https://example/app.zip"}}}
	files := releaseToFiles("/repo", &release)
	if len(files) != 1 || files[0].GetPath() != "/repo/app.zip" || files[0].GetID() != "https://example/app.zip" {
		t.Fatalf("release files = %#v", files)
	}
	dirs := releasesToVersionDirs("/repo", []Release{release})
	if len(dirs) != 1 || dirs[0].GetPath() != "/repo/v1.2.3" || dirs[0].GetSize() != 42 {
		t.Fatalf("release dirs = %#v", dirs)
	}
}

func TestParseReposRejectsMalformedLine(t *testing.T) {
	d := &GithubReleases{}
	if _, err := d.ParseRepos("owner/repo:extra:too-many-parts"); err == nil {
		t.Fatal("ParseRepos accepted malformed repository line")
	}
}
