package occam

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func CacheVolumeNames(volumeName string) []string {
	name := volumeName
	refName := []byte(fmt.Sprintf("%s:latest", name))
	sum := sha256.Sum256(refName)

	parts := strings.SplitN(volumeName, "/", 2)
	if len(parts) == 2 {
		name = parts[1]
	}

	var volumes []string
	for _, t := range []string{"build", "launch", "cache"} {
		volumes = append(volumes, fmt.Sprintf("pack-cache-%x.%s", sum[:6], t))
		volumes = append(volumes, fmt.Sprintf("pack-cache-%s_latest-%x.%s", name, sum[:6], t))
	}

	kanikoRefName := []byte(fmt.Sprintf("%s%s-volume", refName, volumeName))
	kanikoSum := sha256.Sum256(kanikoRefName)
	volumes = append(volumes, fmt.Sprintf("pack-cache-%s_latest-%x.kaniko", name, kanikoSum[:6]))

	return volumes
}
