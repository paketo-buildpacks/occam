package occam_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/freezer"

	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/fakes"

	. "github.com/onsi/gomega"
)

func testBuildpackStore(t *testing.T, when spec.G, it spec.S) {
	var (
		Expect            = NewWithT(t).Expect
		buildpackStore    occam.BuildpackStore
		fakeRemoteFetcher *fakes.RemoteFetcher
		fakeLocalFetcher  *fakes.LocalFetcher
		fakeCacheManager  *fakes.CacheManager
		fakeExtractor     *fakes.RegistryBuildpackToLocal
	)

	it.Before(func() {
		fakeRemoteFetcher = &fakes.RemoteFetcher{}
		fakeLocalFetcher = &fakes.LocalFetcher{}
		fakeCacheManager = &fakes.CacheManager{}
		fakeExtractor = &fakes.RegistryBuildpackToLocal{}

		buildpackStore = occam.NewBuildpackStore()
	})

	when("getting an online buildpack", func() {
		when("from a local uri", func() {
			var localDir string
			var name string
			it.Before(func() {
				localDir = t.TempDir()
				name = filepath.Base(localDir)
				fakeLocalFetcher.GetCall.Returns.String = "/path/to/cool-buildpack/"
				buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
					WithRemoteFetcher(fakeRemoteFetcher).
					WithCacheManager(fakeCacheManager)
			})
			it("returns a local filepath to a buildpack", func() {
				local_url, err := buildpackStore.Get.
					WithVersion("some-version").
					Execute(localDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(1))
				Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(1))

				Expect(local_url).To(Equal("/path/to/cool-buildpack/"))
				Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(0))
				Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(1))
				Expect(fakeLocalFetcher.GetCall.Receives.LocalBuildpack).To(Equal(freezer.LocalBuildpack{
					Path:        localDir,
					Name:        name,
					UncachedKey: name,
					CachedKey:   fmt.Sprintf("%s:cached", name),
					Offline:     false,
					Version:     "some-version",
				}))
			})
		})

		when("from a github uri", func() {
			it.Before(func() {
				fakeRemoteFetcher.GetCall.Returns.String = "/path/to/remote-buildpack/"
				buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
					WithRemoteFetcher(fakeRemoteFetcher).
					WithCacheManager(fakeCacheManager)
			})

			it("returns a local filepath to buildpack", func() {
				local_url, err := buildpackStore.Get.
					WithVersion("some-version").
					Execute("github.com/some-org/some-repo")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(1))
				Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(1))

				Expect(local_url).To(Equal("/path/to/remote-buildpack/"))
				Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(0))
				Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(1))
				Expect(fakeRemoteFetcher.GetCall.Receives.RemoteBuildpack).To(Equal(freezer.RemoteBuildpack{
					Org:         "some-org",
					Repo:        "some-repo",
					Platform:    "linux",
					Arch:        "amd64",
					UncachedKey: "some-org:some-repo:linux:amd64",
					CachedKey:   "some-org:some-repo:linux:amd64:cached",
					Offline:     false,
					Version:     "some-version",
				}))
			})
		})

		when("from a github uri with specific arch and platform", func() {
			it.Before(func() {
				fakeRemoteFetcher.GetCall.Returns.String = "/path/to/remote-buildpack/"
				buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
					WithRemoteFetcher(fakeRemoteFetcher).
					WithCacheManager(fakeCacheManager).
					WithTarget("linux/arm64")
			})

			it("returns a local filepath to buildpack", func() {
				local_url, err := buildpackStore.Get.
					WithVersion("some-version").
					Execute("github.com/some-org/some-repo")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(1))
				Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(1))

				Expect(local_url).To(Equal("/path/to/remote-buildpack/"))
				Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(0))
				Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(1))
				Expect(fakeRemoteFetcher.GetCall.Receives.RemoteBuildpack).To(Equal(freezer.RemoteBuildpack{
					Org:         "some-org",
					Repo:        "some-repo",
					Platform:    "linux",
					Arch:        "arm64",
					UncachedKey: "some-org:some-repo:linux:arm64",
					CachedKey:   "some-org:some-repo:linux:arm64:cached",
					Offline:     false,
					Version:     "some-version",
				}))
			})
		})

		when("from a registry uri", func() {
			url := "some-registry-url"
			localPath := "/some/local/path"
			it.Before(func() {
				fakeExtractor.ExtractCall.Returns.String_1 = localPath
				fakeExtractor.ExtractCall.Returns.String_2 = "some-version"

				fakeLocalFetcher.GetCall.Returns.String = "some-registry-path"
				buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
					WithRemoteFetcher(fakeRemoteFetcher).
					WithCacheManager(fakeCacheManager).
					WithRegistryBuildpackExtractor(fakeExtractor)
			})
			it("returns a local filepath to a buildpack", func() {

				local_url, err := buildpackStore.Get.
					WithVersion("some-version").
					Execute(url)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeExtractor.ExtractCall.CallCount).To(Equal(1))
				Expect(fakeExtractor.ExtractCall.Receives.Ref).To(Equal(url))
				Expect(fakeExtractor.ExtractCall.Receives.Destination).To(Not(BeEmpty()))
				Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(1))
				Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(1))

				Expect(local_url).To(Equal("some-registry-path"))
				Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(0))
				Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(1))
				Expect(fakeLocalFetcher.GetCall.Receives.LocalBuildpack).To(Equal(freezer.LocalBuildpack{
					Path:        "/some/local/path",
					Name:        url,
					UncachedKey: url,
					CachedKey:   fmt.Sprintf("%s:cached", url),
					Offline:     false,
					Version:     "some-version",
				}))
			})
		})

		when("Getting an offline buildpack", func() {
			when("from a local uri", func() {
				var localDir string
				var name string
				it.Before(func() {
					localDir = t.TempDir()
					name = filepath.Base(localDir)
					fakeLocalFetcher.GetCall.Returns.String = "/path/to/cool-buildpack/"
					buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
						WithRemoteFetcher(fakeRemoteFetcher).
						WithCacheManager(fakeCacheManager)
				})

				it("returns a local filepath to a buildpack", func() {
					local_url, err := buildpackStore.Get.
						WithOfflineDependencies().
						Execute(localDir)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(1))
					Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(1))

					Expect(local_url).To(Equal("/path/to/cool-buildpack/"))
					Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(0))
					Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(1))
					Expect(fakeLocalFetcher.GetCall.Receives.LocalBuildpack).To(Equal(freezer.LocalBuildpack{
						Path:        localDir,
						Name:        name,
						UncachedKey: name,
						CachedKey:   fmt.Sprintf("%s:cached", name),
						Offline:     true,
					}))
				})
			})

			when("from a github uri", func() {
				it.Before(func() {
					fakeRemoteFetcher.GetCall.Returns.String = "/path/to/remote-buildpack/"
					buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
						WithRemoteFetcher(fakeRemoteFetcher).
						WithCacheManager(fakeCacheManager)
				})

				it("returns a local filepath to a buildpack", func() {
					local_url, err := buildpackStore.Get.
						WithOfflineDependencies().
						WithVersion("some-version").
						Execute("github.com/some-org/some-repo")
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(1))
					Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(1))

					Expect(local_url).To(Equal("/path/to/remote-buildpack/"))
					Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(0))
					Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(1))
					Expect(fakeRemoteFetcher.GetCall.Receives.RemoteBuildpack).To(Equal(freezer.RemoteBuildpack{
						Org:         "some-org",
						Repo:        "some-repo",
						Platform:    "linux",
						Arch:        "amd64",
						UncachedKey: "some-org:some-repo:linux:amd64",
						CachedKey:   "some-org:some-repo:linux:amd64:cached",
						Offline:     true,
						Version:     "some-version",
					}))
				})
			})

			when("from a registry uri", func() {
				url := "some-registry-url"
				localPath := "/some/local/path"
				it.Before(func() {
					fakeExtractor.ExtractCall.Returns.String_1 = localPath
					fakeExtractor.ExtractCall.Returns.String_2 = "some-version"

					fakeLocalFetcher.GetCall.Returns.String = "some-registry-path"
					buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
						WithRemoteFetcher(fakeRemoteFetcher).
						WithCacheManager(fakeCacheManager).
						WithRegistryBuildpackExtractor(fakeExtractor)
				})
				it("returns a local filepath to a buildpack", func() {

					local_url, err := buildpackStore.Get.
						WithOfflineDependencies().
						WithVersion("some-version").
						Execute(url)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeExtractor.ExtractCall.CallCount).To(Equal(1))
					Expect(fakeExtractor.ExtractCall.Receives.Ref).To(Equal(url))
					Expect(fakeExtractor.ExtractCall.Receives.Destination).To(Not(BeEmpty()))
					Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(1))
					Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(1))

					Expect(local_url).To(Equal("some-registry-path"))
					Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(0))
					Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(1))
					Expect(fakeLocalFetcher.GetCall.Receives.LocalBuildpack).To(Equal(freezer.LocalBuildpack{
						Path:        "/some/local/path",
						Name:        url,
						UncachedKey: url,
						CachedKey:   fmt.Sprintf("%s:cached", url),
						Offline:     true,
						Version:     "some-version",
					}))
				})
			})
		})
	})

	when("failure cases", func() {
		when("unable to open cacheManager", func() {
			it.Before(func() {
				fakeCacheManager.OpenCall.Returns.Error = errors.New("bad bad error")
				buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
					WithRemoteFetcher(fakeRemoteFetcher).
					WithCacheManager(fakeCacheManager)
			})

			it("returns an error", func() {
				_, err := buildpackStore.Get.Execute("some-url")

				Expect(err).To(MatchError("failed to open cacheManager: bad bad error"))
			})
		})

		when("given an incomplete github.com url", func() {
			it("returns an error", func() {
				buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
					WithRemoteFetcher(fakeRemoteFetcher).
					WithCacheManager(fakeCacheManager)
				incompleteURL := "github.com/incomplete"
				_, err := buildpackStore.Get.Execute(incompleteURL)
				Expect(err).To(MatchError("error incomplete github.com url: \"github.com/incomplete\""))
			})
		})

		when("unable to extract registry uri to local path", func() {
			it.Before(func() {
				fakeExtractor.ExtractCall.Returns.Error = errors.New("bad bad error")
				buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
					WithRemoteFetcher(fakeRemoteFetcher).
					WithCacheManager(fakeCacheManager).
					WithRegistryBuildpackExtractor(fakeExtractor)
			})

			it("returns an error", func() {
				_, err := buildpackStore.Get.Execute("some-url")

				Expect(err).To(MatchError("failed to create local buildpack from registry image: bad bad error"))
			})
		})
	})
}
