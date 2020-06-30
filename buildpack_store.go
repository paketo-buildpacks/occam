package occam

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ForestEckhardt/freezer"
	"github.com/ForestEckhardt/freezer/github"
)

//go:generate faux --interface LocalFetcher --output fakes/local_fetcher.go
type LocalFetcher interface {
	Get(freezer.LocalBuildpack) (string, error)
}

//go:generate faux --interface RemoteFetcher --output fakes/remote_fetcher.go
type RemoteFetcher interface {
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
	packager := freezer.NewPackingTools()
	fileSystem := freezer.NewFileSystem(ioutil.TempDir)
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
				fileSystem,
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

type BuildpackStoreGet struct {
	cacheManager CacheManager
	local        LocalFetcher
	remote       RemoteFetcher

	offline bool
	version string
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

		buildpack := freezer.NewRemoteBuildpack(request[1], request[2])
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
