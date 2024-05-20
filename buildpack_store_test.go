package occam_test

import (
	"errors"
	"testing"

	"github.com/sclevine/spec"

	"github.com/ForestEckhardt/freezer"

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
		fakeDockerPull    *fakes.BPStoreDockerPull
	)

	it.Before(func() {
		fakeRemoteFetcher = &fakes.RemoteFetcher{}
		fakeLocalFetcher = &fakes.LocalFetcher{}
		fakeCacheManager = &fakes.CacheManager{}
		fakeDockerPull = &fakes.BPStoreDockerPull{}

		buildpackStore = occam.NewBuildpackStore()

		buildpackStore = buildpackStore.WithLocalFetcher(fakeLocalFetcher).
			WithRemoteFetcher(fakeRemoteFetcher).
			WithCacheManager(fakeCacheManager).
			WithOCIFetcher(fakeDockerPull)
	})

	when("getting an online buildpack", func() {
		when("from a docker uri", func() {
			it("returns the URI to the OCI image", func() {
				local_url, err := buildpackStore.Get.
					Execute("docker://some-image:tag")
				Expect(err).NotTo(HaveOccurred())

				Expect(local_url).To(Equal("docker://some-image:tag"))

				Expect(fakeDockerPull.ExecuteCall.CallCount).To(Equal(1))
				Expect(fakeDockerPull.ExecuteCall.Receives.String).To(Equal("docker://some-image:tag"))

				Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(0))
				Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(0))

				Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(0))
				Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(0))
			})

			it("ignores the online/offline flags", func() {
				local_url, err := buildpackStore.Get.
					WithOfflineDependencies().
					Execute("docker://some-image:tag")
				Expect(err).NotTo(HaveOccurred())

				Expect(local_url).To(Equal("docker://some-image:tag"))

				Expect(fakeDockerPull.ExecuteCall.CallCount).To(Equal(1))
				Expect(fakeDockerPull.ExecuteCall.Receives.String).To(Equal("docker://some-image:tag"))

				Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(0))
				Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(0))

				Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(0))
				Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(0))
			})

			it("ignores the version flag", func() {
				local_url, err := buildpackStore.Get.
					WithVersion("some-version").
					Execute("docker://some-image:tag")
				Expect(err).NotTo(HaveOccurred())

				Expect(local_url).To(Equal("docker://some-image:tag"))

				Expect(fakeDockerPull.ExecuteCall.CallCount).To(Equal(1))
				Expect(fakeDockerPull.ExecuteCall.Receives.String).To(Equal("docker://some-image:tag"))

				Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(0))
				Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(0))

				Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(0))
				Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(0))
			})
		})

		when("from a local uri", func() {
			it.Before(func() {
				fakeLocalFetcher.GetCall.Returns.String = "/path/to/cool-buildpack/"
			})
			it("returns a local filepath to a buildpack", func() {
				local_url, err := buildpackStore.Get.
					WithVersion("some-version").
					Execute("/some/local/path")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(1))
				Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(1))

				Expect(local_url).To(Equal("/path/to/cool-buildpack/"))
				Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(0))
				Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(1))
				Expect(fakeLocalFetcher.GetCall.Receives.LocalBuildpack).To(Equal(freezer.LocalBuildpack{
					Path:        "/some/local/path",
					Name:        "path",
					UncachedKey: "path",
					CachedKey:   "path:cached",
					Offline:     false,
					Version:     "some-version",
				}))
			})
		})

		when("from a github uri", func() {
			it.Before(func() {
				fakeRemoteFetcher.GetCall.Returns.String = "/path/to/remote-buildpack/"
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
					UncachedKey: "some-org:some-repo",
					CachedKey:   "some-org:some-repo:cached",
					Offline:     false,
					Version:     "some-version",
				}))
			})
		})

		when("Getting an offline buildpack", func() {
			when("from a local uri", func() {
				it.Before(func() {
					fakeLocalFetcher.GetCall.Returns.String = "/path/to/cool-buildpack/"
				})

				it("returns a local filepath to a buildpack", func() {
					local_url, err := buildpackStore.Get.
						WithOfflineDependencies().
						Execute("/some/local/path")
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeCacheManager.OpenCall.CallCount).To(Equal(1))
					Expect(fakeCacheManager.CloseCall.CallCount).To(Equal(1))

					Expect(local_url).To(Equal("/path/to/cool-buildpack/"))
					Expect(fakeRemoteFetcher.GetCall.CallCount).To(Equal(0))
					Expect(fakeLocalFetcher.GetCall.CallCount).To(Equal(1))
					Expect(fakeLocalFetcher.GetCall.Receives.LocalBuildpack).To(Equal(freezer.LocalBuildpack{
						Path:        "/some/local/path",
						Name:        "path",
						UncachedKey: "path",
						CachedKey:   "path:cached",
						Offline:     true,
					}))
				})
			})

			when("from a github uri", func() {
				it.Before(func() {
					fakeRemoteFetcher.GetCall.Returns.String = "/path/to/remote-buildpack/"
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
						UncachedKey: "some-org:some-repo",
						CachedKey:   "some-org:some-repo:cached",
						Offline:     true,
						Version:     "some-version",
					}))
				})
			})
		})
	})

	when("failure cases", func() {
		when("attempting to fetch OCI image without OCI fetcher", func() {
			it.Before(func() {
				buildpackStore = occam.NewBuildpackStore()
			})

			it("returns an error", func() {
				_, err := buildpackStore.Get.Execute("docker://some-image:tag")

				Expect(err).To(MatchError("must provide OCI fetcher to fetch OCI images"))
			})
		})

		when("unable to open cacheManager", func() {
			it.Before(func() {
				fakeCacheManager.OpenCall.Returns.Error = errors.New("bad bad error")
			})

			it("returns an error", func() {
				_, err := buildpackStore.Get.Execute("some-url")

				Expect(err).To(MatchError("failed to open cacheManager: bad bad error"))
			})
		})

		when("given an incomplete github.com url", func() {
			it("returns an error", func() {
				incompleteURL := "github.com/incomplete"
				_, err := buildpackStore.Get.Execute(incompleteURL)
				Expect(err).To(MatchError("error incomplete github.com url: \"github.com/incomplete\""))
			})
		})
	})
}
