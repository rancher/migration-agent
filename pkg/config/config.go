package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/k3s/pkg/cli/cmds"
	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/pki"
	"github.com/rancher/wharfie/pkg/registries"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

const (
	configDir           = "/etc/rancher/rke2/config.yaml.d"
	privateRegistryPath = "/etc/rancher/rke2/registries.yaml"
	kubeProxyConfig     = "kubeproxy.kubeconfig"
	rkeClusterConfig    = "rke2-cluster-config"
	registryFlagParts   = 4

	calicoCNI = "calico"
	canalCNI  = "canal"
	weaveCNI  = "weave"
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
	data, err := json.Marshal(args)
	if err != nil {
		return err
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

	// copy the network cni plugin except for weave as its not yet supported by RKE2
	networkPlugin := fullState.CurrentState.RancherKubernetesEngineConfig.Network.Plugin
	if networkPlugin != "" && networkPlugin != weaveCNI {
		argsMap["cni"] = networkPlugin
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

func ConfigurePrivateRegistries(ctx context.Context, fullState *cluster.FullState, registriesTLS []string) error {
	privateRegistryConfig := fullState.CurrentState.RancherKubernetesEngineConfig.PrivateRegistries
	if len(privateRegistryConfig) <= 0 {
		return nil
	}
	if _, err := os.Stat(privateRegistryPath); err != nil {
		if os.IsNotExist(err) {
			r := registries.Registry{}
			r.Configs = make(map[string]registries.RegistryConfig)
			for _, reg := range privateRegistryConfig {
				endpoint := reg.URL
				if endpoint == "" {
					endpoint = "docker.io"
				}
				if reg.User != "" && reg.Password != "" {
					r.Configs[endpoint] = registries.RegistryConfig{
						Auth: &registries.AuthConfig{
							Username: reg.User,
							Password: reg.Password,
						},
						TLS: getRegistryTLSConfig(endpoint, registriesTLS),
					}
				}
			}

			regBytes, err := yaml.Marshal(r)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(privateRegistryPath, regBytes, 0600); err != nil {
				return err
			}
		}
	}

	return nil
}

func getRegistryTLSConfig(endpoint string, registriesTLS []string) *registries.TLSConfig {
	if len(registriesTLS) <= 0 {
		return nil
	}
	var caCert, cert, key, u string
	for _, registryTLS := range registriesTLS {
		certs := strings.Split(registryTLS, ",")
		if len(certs) < registryFlagParts {
			continue
		}
		u = certs[0]
		if u != endpoint {
			continue
		}
		// validating registry url
		if _, err := url.ParseRequestURI(u); err != nil {
			logrus.Warnf("registry url %s is invalid", u)
			continue
		}
		caCert = certs[1]
		cert = certs[2]
		key = certs[3]
	}
	return &registries.TLSConfig{
		CAFile:   caCert,
		CertFile: cert,
		KeyFile:  key,
	}
}
