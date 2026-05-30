package generator

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	// Only the LIGHT file cache driver from the (decoupled) framework: the
	// cache/drivers package carries memory + file stores and no heavy backends
	// (redis lives in the cache/redis leaf, which we deliberately do NOT pull).
	// This is the on-demand single-driver import the framework decouple was
	// built to enable - the installer caches the GitHub tags lookup to disk so
	// repeated `velocity new` runs skip the network call that motivated all of
	// this in the first place.
	"github.com/velocitykode/velocity/cache/drivers"
)

// tagCacheTTL bounds how long a resolved template tag is trusted on disk. An
// hour is long enough to spare the GitHub API call across a burst of project
// setups, short enough that a freshly published template release is picked up
// promptly.
const tagCacheTTL = time.Hour

var (
	tagDiskOnce  sync.Once
	tagDiskStore *drivers.FileStore
)

// diskTagCache lazily builds a file-backed cache store under the user cache
// dir (e.g. ~/Library/Caches/velocity-installer, ~/.cache/velocity-installer).
// Returns nil when the cache dir cannot be resolved or the store cannot be
// created; callers treat nil as "no persistent cache" and fall back to the
// network, so caching is strictly best-effort and never blocks a scaffold.
func diskTagCache() *drivers.FileStore {
	tagDiskOnce.Do(func() {
		dir, err := os.UserCacheDir()
		if err != nil {
			return
		}
		store, err := drivers.NewFileStore("template_tags", filepath.Join(dir, "velocity-installer"))
		if err != nil {
			return
		}
		tagDiskStore = store
	})
	return tagDiskStore
}

// tagCacheKey namespaces the per-repo cache entry.
func tagCacheKey(repo string) string { return "template_tag:" + repo }

// loadCachedTag returns a non-expired cached tag for repo, or ("", false).
func loadCachedTag(repo string) (string, bool) {
	store := diskTagCache()
	if store == nil {
		return "", false
	}
	if v, ok := store.GetStringCtx(context.Background(), tagCacheKey(repo)); ok && v != "" {
		return v, true
	}
	return "", false
}

// storeCachedTag persists a successfully resolved tag. Empty tags (API
// failures) are not cached so a transient outage does not pin "" for an hour.
func storeCachedTag(repo, tag string) {
	if tag == "" {
		return
	}
	if store := diskTagCache(); store != nil {
		_ = store.PutCtx(context.Background(), tagCacheKey(repo), tag, tagCacheTTL)
	}
}
