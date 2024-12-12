package config

import (
	"strings"
)

const (
	EtcdExposeMetrics = "etcd-expose-metrics"

	MirrorRegionCN           = "cn"
	aliSystemDefaultRegistry = "registry.cn-hangzhou.aliyuncs.com"
)

var (
	RuntimeRKE2    Runtime = "rke2"
	RuntimeK3S     Runtime = "k3s"
	RuntimeUnknown Runtime = "unknown"

	ClusterInitRole Role = "cluster-init"
	ServerRole      Role = "server"
	AgentRole       Role = "agent"
)

type Runtime string
type Role string

type RuntimeConfig struct {
	Role                  Role                   `json:"role,omitempty"`
	Server                string                 `json:"server,omitempty"`
	SANS                  []string               `json:"tlsSans,omitempty"`
	NodeName              string                 `json:"nodeName,omitempty"`
	Address               string                 `json:"address,omitempty"`
	InternalAddress       string                 `json:"internalAddress,omitempty"`
	Taints                []string               `json:"taints,omitempty"`
	Labels                []string               `json:"labels,omitempty"`
	Token                 string                 `json:"token,omitempty"`
	SystemDefaultRegistry string                 `json:"systemDefaultRegistry,omitempty"`
	Mirror                string                 `json:"mirror,omitempty"`
	ConfigValues          map[string]interface{} `json:"extraConfig,omitempty"`
}

func (cfg *RuntimeConfig) SetDefaults() {
	// Assign default labels
	if cfg.Labels == nil {
		cfg.Labels = []string{}
	}
	cfg.Labels = append(cfg.Labels, "llmos.ai/managed=true")

	if cfg.ConfigValues == nil {
		cfg.ConfigValues = map[string]interface{}{}
	}

	// Enable etcd metrics by default
	if cfg.ConfigValues[EtcdExposeMetrics] == nil {
		cfg.ConfigValues[EtcdExposeMetrics] = true
	}

	// Determine default role if not explicitly set
	if cfg.Role == "" && cfg.Server != "" && cfg.Token != "" {
		cfg.Role = AgentRole
	}

	// Set default system registry if mirror is set to use cn region
	if cfg.SystemDefaultRegistry == "" && cfg.Mirror == MirrorRegionCN {
		cfg.SystemDefaultRegistry = aliSystemDefaultRegistry
	}
}

func GetRuntime(kubernetesVersion string) Runtime {
	if isRKE2(kubernetesVersion) {
		return RuntimeRKE2
	}

	if isK3s(kubernetesVersion) {
		return RuntimeK3S
	}
	return RuntimeUnknown
}

func isRKE2(kubernetesVersion string) bool {
	return strings.Contains(kubernetesVersion, string(RuntimeRKE2))
}

func isK3s(kubernetesVersion string) bool {
	return strings.Contains(kubernetesVersion, string(RuntimeK3S))
}
