package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	helmv1 "github.com/k3s-io/helm-controller/pkg/apis/helm.cattle.io/v1"
	"github.com/rancher/rke/cluster"
	yamlv3 "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	canalFlannelBackendType      = "canal_flannel_backend_type"
	canalFlannelFlexVolPluginDir = "canal_flex_volume_plugin_dir"
	canalFlannelInterface        = "canal_iface"
)

type canalConfig struct {
	Calico  map[string]string `json:"calico"`
	Flannel map[string]string `json:"flannel"`
}

type calicoConfig struct {
}

// MigrateCNIConfig should read the cni plugin specific configuration and copy it
// as a helm chart config to RKE2 and then save it to the manifest dir, this
// currently only works for canal installation because calico tigera operator
// doesnt contain a lot of customization
func MigrateCNIConfig(ctx context.Context, fullState *cluster.FullState, dataDir string) error {
	networkConfig := fullState.CurrentState.RancherKubernetesEngineConfig.Network
	if networkConfig.Plugin == "" {
		return nil
	}

	// migrate canal config to helm chart
	if networkConfig.Plugin == canalCNI {
		canalCfg := canalConfig{
			Calico: map[string]string{
				"vethuMTU":            strconv.Itoa(networkConfig.MTU),
				"flexVolumePluginDir": networkConfig.Options[canalFlannelFlexVolPluginDir],
			},
			Flannel: map[string]string{
				"backend": networkConfig.Options[canalFlannelBackendType],
				"iface":   networkConfig.Options[canalFlannelInterface],
			},
		}
		helmChartConfig, err := toHelmChartConfig("rke2-"+canalCNI, canalCfg)
		if err != nil {
			return err
		}
		manifestsDir := manifestsDir(dataDir)
		manifestFile := filepath.Join(manifestsDir, "rke2-canal-config.yaml")
		err = os.MkdirAll(manifestsDir, 0755)
		if err != nil {
			return err
		}

		// deploy manifest file
		return ioutil.WriteFile(manifestFile, helmChartConfig, 0600)
	}
	return nil
}

func toHelmChartConfig(helmChartName string, values canalConfig) ([]byte, error) {
	valuesYaml, err := yamlv3.Marshal(&values)
	if err != nil {
		return nil, err
	}
	hc := helmv1.HelmChartConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HelmChartConfig",
			APIVersion: "helm.cattle.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      helmChartName,
			Namespace: "kube-system",
		},
		Spec: helmv1.HelmChartConfigSpec{
			ValuesContent: string(valuesYaml),
		},
	}
	return yaml.Marshal(&hc)
}
