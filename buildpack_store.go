package occam

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/freezer"
	"github.com/paketo-buildpacks/freezer/github"
	"github.com/paketo-buildpacks/occam/packagers"
)

//go:generate faux --interface LocalFetcher --output fakes/local_fetcher.go
type LocalFetcher interface {
	WithPackager(packager freezer.Packager) freezer.LocalFetcher
	Get(freezer.LocalBuildpack) (string, error)
}

//go:generate faux --interface RemoteFetcher --output fakes/remote_fetcher.go
type RemoteFetcher interface {
	WithPackager(packager freezer.Packager) freezer.RemoteFetcher
	Get(freezer.RemoteBuildpack) (string, error)
}

//go:generate faux --interface CacheManager --output fakes/cache_manager.go
type CacheManager interface {
	Get(key string) (freezer.CacheEntry, bool, error)
	Set(key string, cachedEntry freezer.CacheEntry) error
	Dir() string
	Open() error
	Close() error
}

type BuildpackStore struct {
	Get BuildpackStoreGet
}

func NewBuildpackStore() BuildpackStore {
	gitToken := os.Getenv("GIT_TOKEN")
	cacheManager := freezer.NewCacheManager(filepath.Join(os.Getenv("HOME"), ".freezer-cache"))
	releaseService := github.NewReleaseService(github.NewConfig("https://api.github.com", gitToken))
	packager := packagers.NewJam()
	namer := freezer.NewNameGenerator()

	return BuildpackStore{
		Get: BuildpackStoreGet{
			local: freezer.NewLocalFetcher(
				&cacheManager,
				packager,
				namer,
			),
			remote: freezer.NewRemoteFetcher(
				&cacheManager,
				releaseService, packager,
			),
			cacheManager: &cacheManager,
		},
	}
}

func (bs BuildpackStore) WithLocalFetcher(fetcher LocalFetcher) BuildpackStore {
	bs.Get.local = fetcher
	return bs
}

func (bs BuildpackStore) WithRemoteFetcher(fetcher RemoteFetcher) BuildpackStore {
	bs.Get.remote = fetcher
	return bs
}

func (bs BuildpackStore) WithCacheManager(manager CacheManager) BuildpackStore {
	bs.Get.cacheManager = manager
	return bs
}

func (bs BuildpackStore) WithPackager(packager freezer.Packager) BuildpackStore {
	bs.Get.local = bs.Get.local.WithPackager(packager)
	bs.Get.remote = bs.Get.remote.WithPackager(packager)
	return bs
}

func (bs BuildpackStore) WithTarget(target string) BuildpackStore {
	targetExploded := strings.Split(target, "/")
	bs.Get.platform = targetExploded[0]
	bs.Get.arch = targetExploded[1]
	return bs
}

type BuildpackStoreGet struct {
	cacheManager CacheManager
	local        LocalFetcher
	remote       RemoteFetcher

	offline bool
	version string

	platform string
	arch     string
}

func (g BuildpackStoreGet) Execute(url string) (string, error) {
	err := g.cacheManager.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open cacheManager: %s", err)
	}
	defer g.cacheManager.Close()

	if strings.HasPrefix(url, "github.com") {
		request := strings.SplitN(url, "/", 3)
		if len(request) < 3 {
			return "", fmt.Errorf("error incomplete github.com url: %q", url)
		}

		if g.platform == "" || g.arch == "" {
			g.platform = "linux"
			g.arch = "amd64"
		}

		buildpack := freezer.NewRemoteBuildpack(request[1], request[2], g.platform, g.arch)
		buildpack.Offline = g.offline
		buildpack.Version = g.version

		return g.remote.Get(buildpack)
	}

	buildpack := freezer.NewLocalBuildpack(url, filepath.Base(url))
	buildpack.Offline = g.offline
	buildpack.Version = g.version

	return g.local.Get(buildpack)
}

func (g BuildpackStoreGet) WithOfflineDependencies() BuildpackStoreGet {
	g.offline = true
	return g
}

func (g BuildpackStoreGet) WithVersion(version string) BuildpackStoreGet {
	g.version = version
	return g
}
