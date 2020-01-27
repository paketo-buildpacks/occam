package occam

import (
	"encoding/json"
	"strings"
)

type Container struct {
	ID    string
	Ports map[string]string
}

func NewContainerFromInspectOutput(output []byte) (Container, error) {
	var inspect []struct {
		ID              string `json:"Id"`
		NetworkSettings struct {
			Ports map[string]struct {
				HostPort string `json:"HostPort"`
			} `json:"Ports"`
		} `json:"NetworkSettings"`
	}

	err := json.Unmarshal(output, &inspect)
	if err != nil {
		return Container{}, err
	}

	ports := make(map[string]string)
	for key, value := range inspect[0].NetworkSettings.Ports {
		ports[strings.TrimSuffix(key, "/tcp")] = value.HostPort
	}

	return Container{
		ID:    inspect[0].ID,
		Ports: ports,
	}, nil
}
