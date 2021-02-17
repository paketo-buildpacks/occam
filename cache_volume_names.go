package occam

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func CacheVolumeNames(name string) []string {
	refName := []byte(fmt.Sprintf("%s:latest", name))
	sum := sha256.Sum256(refName)

	parts := strings.SplitN(name, "/", 2)
	if len(parts) == 2 {
		name = parts[1]
	}

	var volumes []string
	for _, t := range []string{"build", "launch", "cache"} {
		volumes = append(volumes, fmt.Sprintf("pack-cache-%x.%s", sum[:6], t))
		volumes = append(volumes, fmt.Sprintf("pack-cache-%s_latest-%x.%s", name, sum[:6], t))
	}

	return volumes
}
