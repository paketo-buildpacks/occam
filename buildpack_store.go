package occam

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/freezer"
	"github.com/paketo-buildpacks/freezer/github"
	"github.com/paketo-buildpacks/occam/packagers"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/vacation"
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

//go:generate faux --interface RegistryBuildpackToLocal --output fakes/registry_buildpack_to_local.go
type RegistryBuildpackToLocal interface {
	Extract(ref string, destination string) (string, string, error)
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
	extractor := NewRegistryBuildpackImageExtractor(NewDocker())

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
			extractor:    extractor,
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

func (bs BuildpackStore) WithRegistryBuildpackExtractor(extractor RegistryBuildpackToLocal) BuildpackStore {
	bs.Get.extractor = extractor
	return bs
}

type BuildpackStoreGet struct {
	cacheManager CacheManager
	local        LocalFetcher
	remote       RemoteFetcher
	extractor    RegistryBuildpackToLocal

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

	info, err := os.Stat(url)

	switch {
	case err == nil && info.IsDir():
		buildpack := freezer.NewLocalBuildpack(url, filepath.Base(url)).
			WithOffline(g.offline).
			WithVersion(g.version)

		return g.local.Get(buildpack)
	case strings.HasPrefix(url, "github.com"):
		request := strings.SplitN(url, "/", 3)
		if len(request) < 3 {
			return "", fmt.Errorf("error incomplete github.com url: %q", url)
		}

		if g.platform == "" || g.arch == "" {
			g.platform = "linux"
			g.arch = "amd64"
		}

		buildpack := freezer.NewRemoteBuildpack(request[1], request[2], g.platform, g.arch).
			WithOffline(g.offline).
			WithVersion(g.version)

		return g.remote.Get(buildpack)
	default:
		tmpDir, err := os.MkdirTemp("", filepath.Base(url))
		if err != nil {
			return "", fmt.Errorf("failed to create temp dir: %w", err)
		}

		buildpackRootPath, version, err := g.extractor.Extract(url, tmpDir)
		if err != nil {
			return "", fmt.Errorf("failed to create local buildpack from registry image: %w", err)
		}

		buildpack := freezer.NewLocalBuildpack(buildpackRootPath, filepath.Base(url)).
			WithOffline(g.offline).
			WithVersion(version)

		return g.local.Get(buildpack)
	}
}

func (g BuildpackStoreGet) WithOfflineDependencies() BuildpackStoreGet {
	g.offline = true
	return g
}

func (g BuildpackStoreGet) WithVersion(version string) BuildpackStoreGet {
	g.version = version
	return g
}

type RegistryBuildpackImageExtractor struct {
	docker Docker
}

func NewRegistryBuildpackImageExtractor(docker Docker) RegistryBuildpackImageExtractor {
	return RegistryBuildpackImageExtractor{
		docker: docker,
	}
}

func (e RegistryBuildpackImageExtractor) Extract(ref string, destination string) (string, string, error) {
	err := e.docker.Pull.Execute(ref)
	if err != nil {
		return "", "", fmt.Errorf("failed to pull buildpack image: %s", err)
	}

	img, err := e.docker.Image.ExportToOCI.Execute(ref)
	if err != nil {
		return "", "", fmt.Errorf("failed get oci image: %s", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return "", "", fmt.Errorf("failed to get image layers: %s", err)
	}

	if len(layers) == 0 {
		return "", "", fmt.Errorf("no layers found in image")
	}
	layer0 := layers[0]

	reader, err := layer0.Uncompressed()
	if err != nil {
		return "", "", fmt.Errorf("failed to get layer: %w", err)
	}
	defer func() {
		if err2 := reader.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()

	err = vacation.NewArchive(reader).Decompress(destination)
	if err != nil {
		return "", "", fmt.Errorf("failed to decompress layer: %w", err)
	}

	buildpackRoot, version, err := e.GetBuildpackRootAndVersion(destination)
	if err != nil {
		return "", "", err
	}

	return buildpackRoot, version, nil
}

// Get buildpack root and version, and update buildpack toml so packager will work
func (e RegistryBuildpackImageExtractor) GetBuildpackRootAndVersion(path string) (string, string, error) {
	var buildpackTomlPath, buildpackRootDir string

	files := []string{}
	err := filepath.WalkDir(path, func(walkPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			files = append(files, walkPath)

			if filepath.Base(walkPath) == "buildpack.toml" {
				buildpackTomlPath = walkPath
				buildpackRootDir = filepath.Dir(walkPath)
			}
		}
		return nil
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to access extracted buildpack path: %w", err)
	}

	parser := cargo.NewBuildpackParser()
	config, err := parser.Parse(buildpackTomlPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse buildpack config: %w", err)
	}

	config.Metadata.PrePackage = ""
	config.Metadata.IncludeFiles = []string{}
	for _, filePath := range files {
		config.Metadata.IncludeFiles = append(config.Metadata.IncludeFiles, strings.Replace(filePath, buildpackRootDir+"/", "", 1))
	}

	file, err := os.OpenFile(buildpackTomlPath, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", fmt.Errorf("failed to open buildpack config: %w", err)
	}
	defer func() {
		if err2 := file.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()

	err = cargo.EncodeConfig(file, config)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode buildpack config: %w", err)
	}

	return buildpackTomlPath, config.Buildpack.Version, nil
}
