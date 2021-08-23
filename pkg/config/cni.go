package config

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	helmv1 "github.com/k3s-io/helm-controller/pkg/apis/helm.cattle.io/v1"
	"github.com/rancher/rke/cluster"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	canalFlannelBackendType      = "canal_flannel_backend_type"
	canalFlannelFlexVolPluginDir = "canal_flex_volume_plugin_dir"
	canalFlannelInterface        = "canal_iface"

	calicoFlexVolumePluginDir = "calico_flex_volume_plugin_dir"
)

// MigrateCNIConfig should read the cni plugin specific configuration and copy it
// as a helm chart config to RKE2 and then save it to the manifest dir, this
// currently only works for canal installation because calico tigera operator
// doesnt contain a lot of customization
func MigrateCNIConfig(ctx context.Context, fullState *cluster.FullState, dataDir string) error {
	var (
		helmChartConfig []byte
		err             error
	)
	networkConfig := fullState.CurrentState.RancherKubernetesEngineConfig.Network
	if networkConfig.Plugin == "" {
		return nil
	}

	// migrate canal config to helm chart
	if networkConfig.Plugin == canalCNI {
		logrus.Info("Canal CNI plugin is used by RKE1, migrating config to RKE2")
		canalCfg := CanalConfig{
			Calico: map[string]string{
				"vethuMTU":            strconv.Itoa(networkConfig.MTU),
				"flexVolumePluginDir": networkConfig.Options[canalFlannelFlexVolPluginDir],
			},
			Flannel: map[string]string{
				"backend": networkConfig.Options[canalFlannelBackendType],
				"iface":   networkConfig.Options[canalFlannelInterface],
			},
		}
		helmChartConfig, err = toHelmChartConfig("rke2-"+networkConfig.Plugin, canalCfg)
		if err != nil {
			return err
		}
	} else if networkConfig.Plugin == calicoCNI {
		logrus.Info("Calico CNI plugin is used by RKE1, migrating config to RKE2")
		calicoCfg := CalicoConfig{
			Installation: CalicoInstallationSpec{
				FlexVolumePath: networkConfig.Options[calicoFlexVolumePluginDir],
				CalicoNetwork: map[string]int{
					"mtu": networkConfig.MTU,
				},
				ControlPlaneNodeSelector: networkConfig.NodeSelector,
			},
		}
		helmChartConfig, err = toHelmChartConfig("rke2-"+networkConfig.Plugin, calicoCfg)
		if err != nil {
			return err
		}
	} else {
		logrus.Infof("network plugin is not recognized as rke2 network plugin")
		return nil
	}

	manifestsDir := manifestsDir(dataDir)
	manifestFile := filepath.Join(manifestsDir, "rke2-"+networkConfig.Plugin+"-config.yaml")
	err = os.MkdirAll(manifestsDir, 0700)
	if err != nil {
		return err
	}

	// deploy manifest file
	return ioutil.WriteFile(manifestFile, helmChartConfig, 0600)
}

func toHelmChartConfig(helmChartName string, values interface{}) ([]byte, error) {
	var (
		valuesYaml []byte
		err        error
	)
	if helmChartName == "rke2-"+canalCNI {
		valuesConfig, ok := values.(CanalConfig)
		if !ok {
			return nil, errors.New("invalid RKE CanalConfig")
		}
		valuesYaml, err = yaml.Marshal(&valuesConfig)
		if err != nil {
			return nil, err
		}
	} else if helmChartName == "rke2-"+calicoCNI {
		valuesConfig, ok := values.(CalicoConfig)
		if !ok {
			return nil, errors.New("invalid RKE CalicoConfig")
		}
		valuesYaml, err = yaml.Marshal(&valuesConfig)
		if err != nil {
			return nil, err
		}
	} else if helmChartName == "rke2-"+nginxIngress {
		valuesConfig, ok := values.(IngressConfig)
		if !ok {
			return nil, errors.New("invalid RKE Ingress Config")
		}
		valuesYaml, err = yaml.Marshal(&valuesConfig)
		if err != nil {
			return nil, err
		}
	}

	if string(valuesYaml) == "" {
		return nil, nil
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
