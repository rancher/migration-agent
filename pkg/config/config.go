package config

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rancher/k3s/pkg/cli/cmds"
	"github.com/rancher/rke/cluster"
	v1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	rkeClusterConfig = "rke2-cluster-config"
)

func ExportClusterConfiguration(ctx context.Context, fullState *cluster.FullState, configMap v1.ConfigMapController) error {
	args := getClusterConfig(fullState)
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
		return err
	}

	return nil
}

func getClusterConfig(fullState *cluster.FullState) map[string]string {
	services := fullState.CurrentState.RancherKubernetesEngineConfig.Services
	return map[string]string{
		cmds.ServiceCIDR.Name:          services.KubeAPI.ServiceClusterIPRange,
		cmds.ClusterCIDR.Name:          services.KubeController.ClusterCIDR,
		cmds.ServiceNodePortRange.Name: services.KubeAPI.ServiceNodePortRange,
		cmds.ClusterDomain.Name:        services.Kubelet.ClusterDomain,
		cmds.ClusterDNS.Name:           services.Kubelet.ClusterDNSServer,
		cmds.ExtraAPIArgs.Name:         mapToString(services.KubeAPI.ExtraArgs),
		cmds.ExtraControllerArgs.Name:  mapToString(services.KubeController.ExtraArgs),
		cmds.ExtraSchedulerArgs.Name:   mapToString(services.Scheduler.ExtraArgs),
		cmds.ExtraKubeletArgs.Name:     mapToString(services.Kubelet.ExtraArgs),
		cmds.ExtraKubeProxyArgs.Name:   mapToString(services.Kubeproxy.ExtraArgs),
	}
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
