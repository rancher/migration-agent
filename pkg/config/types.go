package config

import (
	"github.com/rancher/rke/types"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type CanalConfig struct {
	Calico  map[string]string `json:"calico,omitempty"`
	Flannel map[string]string `json:"flannel,omitempty"`
}

type CalicoConfig struct {
	Installation CalicoInstallationSpec `json:"installation,omitempty"`
}

type CalicoInstallationSpec struct {
	// we only copy the mtu config from rke1 so we only need a map of string to integer
	CalicoNetwork            map[string]int    `json:"calicoNetwork,omitempty"`
	FlexVolumePath           string            `json:"flexVolumePath,omitempty"`
	ControlPlaneNodeSelector map[string]string `json:"controlPlaneNodeSelector,omitempty"`
}

type IngressConfig struct {
	ControllerConfig IngressControllerConfig `json:"controller,omitempty"`
	DefaultBackend   DefaultBackendConfig    `json:"defaultBackend,omitempty"`
}

type IngressControllerConfig struct {
	Config            map[string]string               `json:"config,omitempty"`
	NodeSelector      map[string]string               `json:"nodeSelector,omitempty"`
	ExtraArgs         map[string]string               `json:"extraArgs,omitempty"`
	ExtraEnvs         []types.ExtraEnv                `json:"extraEnvs,omitempty"`
	ExtraVolumes      []types.ExtraVolume             `json:"extraVolumes,omitempty"`
	ExtraVolumeMounts []types.ExtraVolumeMount        `json:"extraVolumeMounts,omitempty"`
	DNSPolicy         string                          `json:"dnsPolicy,omitempty"`
	UpdateStrategy    *appsv1.DaemonSetUpdateStrategy `json:"updateStrategy,omitempty"`
	HostPorts         IngressHostPorts                `json:"hostPort,omitempty"`
	HostNetwork       bool                            `json:"HostNetwork,omitempty"`
	Tolerations       []v1.Toleration                 `json:"tolerations,omitempty"`
	PriorityClassName string                          `json:"priorityClassName,omitempty"`
}

type DefaultBackendConfig struct {
	Enabled           bool   `json:"enabled,omitempty"`
	PriorityClassName string `json:"priorityClassName,omitempty"`
}

type IngressHostPorts struct {
	Ports IngressPorts `json:"ports,omitempty"`
}

type IngressPorts struct {
	HTTPPort  int `json:"http,omitempty"`
	HTTPSPort int `json:"https,omitempty"`
}
