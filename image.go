package occam

type Image struct {
	ID         string
	Buildpacks []ImageBuildpackMetadata
}

type ImageBuildpackMetadata struct {
	Key    string
	Layers map[string]ImageBuildpackMetadataLayer
}

type ImageBuildpackMetadataLayer struct {
	SHA    string
	Build  bool
	Launch bool
	Cache  bool
}
