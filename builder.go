package occam

type Builder struct {
	BuilderName string      `json:"builder_name"`
	Trusted     bool        `json:"trusted"`
	Default     bool        `json:"default"`
	LocalInfo   BuilderInfo `json:"local_info"`
	RemoteInfo  BuilderInfo `json:"remote_info"`
}

type BuilderInfo struct {
	Description    string                      `json:"description"`
	CreatedBy      BuilderInfoCreatedBy        `json:"created_by"`
	Stack          BuilderInfoStack            `json:"stack"`
	Lifecycle      BuilderInfoLifecycle        `json:"lifecycle"`
	RunImages      []BuilderInfoRunImage       `json:"run_images"`
	Buildpacks     []BuilderInfoBuildpack      `json:"buildpacks"`
	DetectionOrder []BuilderInfoDetectionOrder `json:"detection_order"`
}

type BuilderInfoCreatedBy struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type BuilderInfoStack struct {
	ID string `json:"id"`
}

type BuilderInfoLifecycle struct {
	Version       string                   `json:"version"`
	BuildpackAPIs BuilderInfoLifecycleAPIs `json:"buildpack_apis"`
	PlatformAPIs  BuilderInfoLifecycleAPIs `json:"platform_apis"`
}

type BuilderInfoLifecycleAPIs struct {
	Deprecated []string `json:"deprecated"`
	Supported  []string `json:"supported"`
}

type BuilderInfoRunImage struct {
	Name string `json:"name"`
}

type BuilderInfoBuildpack struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Homepage string `json:"homepage"`
}

type BuilderInfoDetectionOrder struct {
	Buildpacks []BuilderInfoDetectionOrderBuildpack `json:"buildpacks"`
}

type BuilderInfoDetectionOrderBuildpack struct {
	ID         string                               `json:"id"`
	Version    string                               `json:"version"`
	Optional   bool                                 `json:"optional,omitempty"`
	Buildpacks []BuilderInfoDetectionOrderBuildpack `json:"buildpacks"`
}
