package occam

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Container struct {
	ID          string
	Ports       map[string]string
	Env         map[string]string
	IPAddresses map[string]string
}

func NewContainerFromInspectOutput(output []byte) (Container, error) {
	var inspect []struct {
		ID     string `json:"Id"`
		Config struct {
			Env []string `json:"Env"`
		} `json:"Config"`
		NetworkSettings struct {
			Ports map[string][]struct {
				HostPort string `json:"HostPort"`
			} `json:"Ports"`
			Networks map[string]struct {
				IPAddress string `json:"IPAddress"`
			} `json:"Networks"`
		} `json:"NetworkSettings"`
	}

	err := json.Unmarshal(output, &inspect)
	if err != nil {
		return Container{}, err
	}

	container := Container{ID: inspect[0].ID}

	if len(inspect[0].NetworkSettings.Ports) > 0 {
		container.Ports = make(map[string]string)

		for key, value := range inspect[0].NetworkSettings.Ports {
			container.Ports[strings.TrimSuffix(key, "/tcp")] = value[0].HostPort
		}
	}

	if len(inspect[0].Config.Env) > 0 {
		container.Env = make(map[string]string)

		for _, e := range inspect[0].Config.Env {
			parts := strings.SplitN(e, "=", 2)
			container.Env[parts[0]] = parts[1]
		}
	}

	if len(inspect[0].NetworkSettings.Networks) > 0 {
		container.IPAddresses = make(map[string]string)
		for networkName, network := range inspect[0].NetworkSettings.Networks {
			container.IPAddresses[networkName] = network.IPAddress
		}
	}

	return container, nil
}

func (c Container) HostPort(value string) string {
	return c.Ports[value]
}

func (c Container) Host() string {
	val, ok := os.LookupEnv("DOCKER_HOST")
	if !ok || strings.HasPrefix(val, "unix://") {
		return "localhost"
	}

	url, err := url.Parse(val)
	if err != nil {
		return "localhost"
	}

	return url.Hostname()
}

func (c Container) IPAddressForNetwork(networkName string) (string, error) {
	ipAddress, ok := c.IPAddresses[networkName]
	if !ok {
		return "", fmt.Errorf("invalid network name: %s", networkName)
	}

	return ipAddress, nil
}
