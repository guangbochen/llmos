package images

import (
	"fmt"
	"strings"

	"github.com/llmos-ai/llmos/pkg/bootstrap/config"
)

const (
	defaultRuntimeImagePrefix   = "rancher/system-agent-installer"
	defaultInstallerImagePrefix = "llmos-ai/system-installer"
	ghcrRegistry                = "ghcr.io"
)

func GetLLMOSInstallerImage(imageOverride, registry, operatorVersion string) string {
	if registry == "" {
		registry = ghcrRegistry
	}
	return getInstallerImage(imageOverride, registry, defaultInstallerImagePrefix, "llmos-operator", operatorVersion)
}

func GetRuntimeInstallerImage(imageOverride, registry, kubernetesVersion string) string {
	return getInstallerImage(imageOverride, registry, defaultRuntimeImagePrefix,
		string(config.GetRuntime(kubernetesVersion)), kubernetesVersion)
}

func getInstallerImage(imageOverride, registry, imagePrefix, component, version string) string {
	if imageOverride != "" {
		return imageOverride
	}

	if imagePrefix == "" {
		imagePrefix = defaultInstallerImagePrefix
	}

	tag := strings.ReplaceAll(version, "+", "-")
	if tag == "" {
		tag = "latest"
	}

	if registry == "" {
		return fmt.Sprintf("%s-%s:%s", imagePrefix, component, tag)
	}

	return fmt.Sprintf("%s/%s-%s:%s", registry, imagePrefix, component, tag)
}
