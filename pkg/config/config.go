package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/k3s/pkg/cli/cmds"
	"github.com/rancher/rke/cluster"
	v1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	configDir        = "/etc/rancher/rke2/config.yaml.d"
	rkeClusterConfig = "rke2-cluster-config"
)

func ExportClusterConfiguration(ctx context.Context, fullState *cluster.FullState, configMap v1.ConfigMapController, nodeName string) error {
	args := getClusterConfig(fullState, nodeName)
	_, err := configMap.Get("kube-system", rkeClusterConfig, metav1.GetOptions{})
	if err == nil {
		return fmt.Errorf("configMap %s already exist", rkeClusterConfig)
	}
	// create a configmap with the cluster config args
	data, err := json.Marshal(args)
	if err != nil {
		return err
	}
	cfgMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      rkeClusterConfig,
			Namespace: "kube-system",
		},
		Data: map[string]string{
			rkeClusterConfig: string(data),
		},
	}

	_, err = configMap.Create(cfgMap)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	// create a config.d file to add cluster config
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "10-migration.yaml")
	return ioutil.WriteFile(configPath, data, 0644)
}

func getClusterConfig(fullState *cluster.FullState, nodeName string) map[string]string {
	services := fullState.CurrentState.RancherKubernetesEngineConfig.Services
	argsMap := map[string]string{
		cmds.ServiceCIDR.Name:          services.KubeAPI.ServiceClusterIPRange,
		cmds.ClusterCIDR.Name:          services.KubeController.ClusterCIDR,
		cmds.ServiceNodePortRange.Name: services.KubeAPI.ServiceNodePortRange,
		cmds.ClusterDomain.Name:        services.Kubelet.ClusterDomain,
		cmds.ClusterDNS.Name:           services.Kubelet.ClusterDNSServer,
		cmds.NodeNameFlag.Name:         nodeName,
	}
	if len(services.KubeAPI.ExtraArgs) > 0 {
		argsMap[cmds.ExtraAPIArgs.Name] = mapToString(services.KubeAPI.ExtraArgs)
	}
	if len(services.KubeController.ExtraArgs) > 0 {
		argsMap[cmds.ExtraControllerArgs.Name] = mapToString(services.KubeController.ExtraArgs)
	}
	if len(services.Scheduler.ExtraArgs) > 0 {
		argsMap[cmds.ExtraSchedulerArgs.Name] = mapToString(services.Scheduler.ExtraArgs)
	}
	if len(services.Kubelet.ExtraArgs) > 0 {
		argsMap[cmds.ExtraKubeletArgs.Name] = mapToString(services.Kubelet.ExtraArgs)
	}
	return argsMap
}

func mapToString(args map[string]string) string {
	argsJoined := ""
	for k, v := range args {
		if v == "" {
			argsJoined += k + ","
		} else {
			argsJoined += k + "=" + v + ","
		}
	}
	return strings.TrimSuffix(argsJoined, ",")
}
