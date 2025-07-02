# `github.com/paketo-buildpacks/occam`

[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/paketo-buildpacks/occam)

`occam` is a Go library that provides an integration test framework that can be used to test Paketo Buildpacks

## Usage

```bash
go get github.com/paketo-buildpacks/occam
```

## Examples

### Package a buildpack

`occam` can be used to package a buildpack for use under test.

```go
var buildpack string

root, err = filepath.Abs("./..")
Expect(err).ToNot(HaveOccurred())

buildpackStore := occam.NewBuildpackStore().
    WithPackager(packagers.NewLibpak())

buildpack, err = buildpackStore.Get.
    WithVersion("1.2.3").
    Execute(root)
Expect(err).NotTo(HaveOccurred())
```

#### Important update: Libpak v2.0.0+

As of version `2.0.0` of [libpak](https://github.com/paketo-buildpacks/libpak), the implementation of `create-package` has changed and instead the binaries from [libpak tools](https://github.com/paketo-buildpacks/libpak-tools) are being used.
For backward compatibility and easier transition to libpak `2.0.0`, a new packager was introduced:

The new packager is named `NewLibpakTools` and you can use it in the same way as `NewLibpak` by calling it like this:

```go
buildpackStore := occam.NewBuildpackStore().
    WithPackager(packagers.NewLibpakTools())
```

In both cases the `libpak` need to be installed and available in your environemt. For `libpak` prior to `2.0.0` you can find the installation process on below url:

- https://github.com/paketo-buildpacks/github-config/blob/12ac77d11b435250bd0934c0d59c9c41eaa2ce01/implementation/scripts/.util/tools.sh#L259-L284

For versions `2.0.0` and above you can find the installation process on below url:

- https://github.com/paketo-buildpacks/github-config/blob/12ac77d11b435250bd0934c0d59c9c41eaa2ce01/implementation/scripts/.util/tools.sh#L201-L257

### Test a buildpack

Initialize helpers:

```go
pack := occam.NewPack().WithVerbose()
docker := occam.NewDocker()
venom := occam.NewVenom()
testContainers := occam.NewTestContainers()
```

Generate a random name for an image:

```go
imageName, err := occam.RandomName()
Expect(err).ToNot(HaveOccurred())
```

Use the pack helper to build a container image:

```go
var err error
var buildLogs fmt.Stringer
var image occam.Image

image, buildLogs, err = pack.WithNoColor().Build.
	WithBuildpacks(buildpack).
	WithEnv(map[string]string{
		"BP_JVM_VERSION": "11",
		"BP_JVM_TYPE": "jdk",
	}).
	WithBuilder("paketobuildpacks/builder:base").
	WithPullPolicy("if-not-present").
	WithClearCache().
	Execute(imageName, "/path/to/test/application")
Expect(err).ToNot(HaveOccurred())
```

Use the docker helper to run the container image:

```go
container, err = docker.Container.Run.
	WithEnv(map[string]string{"PORT": "8080"}).
	WithPublish("8080").
	Execute(image.ID)
Expect(err).NotTo(HaveOccurred())
```

Validate that the application returns the correct response:

```go
Eventually(container, time.Second*30).
	Should(Serve(ContainSubstring(`{"application_status":"UP"}`)).OnPort(8080))
```

### Test a container image with container structure tests

Initialize helpers:

```go
containerStructureTest := NewContainerStructureTest()
```

Call helper to verify the structure of the container

```go
_, err := containerStructureTest.Execute("test/my-image", "config.yaml")
Expect(err).NotTo(HaveOccurred())

```

Refer to [https://github.com/GoogleContainerTools/container-structure-test](https://github.com/GoogleContainerTools/container-structure-test) for available tests (e.g. [command tests](https://github.com/GoogleContainerTools/container-structure-test#command-tests), [file existence tests](https://github.com/GoogleContainerTools/container-structure-test#file-existence-tests), ...)

### Use venom to extent your integration test

Initialize helpers:

```go
venom = occam.NewVenom()
```

Call helper to run a testsuite against a container

```go
_, err := venom.WithPort("8080").Execute("testsuite.yaml")
Expect(err).NotTo(HaveOccurred())

```

Refer to [https://github.com/ovh/venom](https://github.com/ovh/venom) for details (e.g. [testsuites](https://github.com/ovh/venom#testsuites), [executors](https://github.com/ovh/venom#executors), ...)

## License

This library is released under version 2.0 of the [Apache License][a].

[a]: https://www.apache.org/licenses/LICENSE-2.0
