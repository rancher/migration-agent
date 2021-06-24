package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/k3s/pkg/cli/cmds"
	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/pki"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	configDir        = "/etc/rancher/rke2/config.yaml.d"
	kubeProxyConfig  = "kubeproxy.kubeconfig"
	rkeClusterConfig = "rke2-cluster-config"
)

var (
	kubeconfigTemplate = template.Must(template.New("kubeconfig").Parse(`apiVersion: v1
clusters:
- cluster:
    server: {{.URL}}
    certificate-authority-data: {{.CACert}}
  name: local
contexts:
- context:
    cluster: local
    namespace: default
    user: user
  name: Default
current-context: Default
kind: Config
preferences: {}
users:
- name: user
  user:
    client-certificate-data: {{.ClientCert}}
    client-key-data: {{.ClientKey}}
`))
)

func ExportClusterConfiguration(ctx context.Context, fullState *cluster.FullState, nodeName string) error {
	args := getClusterConfig(fullState, nodeName)
	// _, err := configMap.Get("kube-system", rkeClusterConfig, metav1.GetOptions{})
	// if err == nil {
	// 	return fmt.Errorf("configMap %s already exist", rkeClusterConfig)
	// }
	// create a configmap with the cluster config args
	data, err := json.Marshal(args)
	if err != nil {
		return err
	}
	// cfgMap := &corev1.ConfigMap{
	// 	TypeMeta: metav1.TypeMeta{
	// 		Kind:       "ConfigMap",
	// 		APIVersion: "v1",
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      rkeClusterConfig,
	// 		Namespace: "kube-system",
	// 	},
	// 	Data: map[string]string{
	// 		rkeClusterConfig: string(data),
	// 	},
	// }

	// _, err = configMap.Create(cfgMap)
	// if err != nil {
	// 	if !apierrors.IsAlreadyExists(err) {
	// 		return err
	// 	}
	// }

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

func ExportKubeProxyConfig(fullState *cluster.FullState, dataDir string) error {
	kubeProxyCert := fullState.CurrentState.CertificatesBundle[pki.KubeProxyCertName]
	caCert := fullState.CurrentState.CertificatesBundle[pki.CACertName]
	config, err := clientcmd.BuildConfigFromFlags("", kubeProxyCert.ConfigPath)
	if err != nil {
		return err
	}
	proxyConfigPath := filepath.Join(dataDir, "agent", kubeProxyConfig)
	if _, err := os.Stat(proxyConfigPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(filepath.Join(dataDir, "agent"), 0700); err != nil {
			return err
		}
		data := struct {
			URL        string
			CACert     string
			ClientCert string
			ClientKey  string
		}{
			URL:        config.Host,
			CACert:     base64.URLEncoding.EncodeToString([]byte(caCert.CertificatePEM)),
			ClientCert: base64.URLEncoding.EncodeToString([]byte(kubeProxyCert.CertificatePEM)),
			ClientKey:  base64.URLEncoding.EncodeToString([]byte(kubeProxyCert.KeyPEM)),
		}

		output, err := os.Create(proxyConfigPath)
		if err != nil {
			return err
		}
		defer output.Close()

		return kubeconfigTemplate.Execute(output, &data)
	}

	return nil
}
