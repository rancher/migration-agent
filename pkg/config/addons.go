package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/types"
	appsv1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const (
	corednsConfigMap  = "rke-coredns-addon"
	ingrerssConfigMap = "rke-ingress-controller"
	metricsConfigMap  = "rke-metrics-addon"
	networkConfigMap  = "rke-network-plugin"
)

func manifestsDir(dataDir string) string {
	return filepath.Join(dataDir, "server", "manifests")
}

func RemoveOldAddons(ctx context.Context, dataDir string) error {
	objs := []runtime.Object{}
	crb := roleBinding()
	removalJob := job()
	sa := serviceAccount()

	objs = append(objs, crb, sa, removalJob)
	yamlContent, err := objectsToYaml(objs)
	if err != nil {
		return err
	}
	manifestsDir := manifestsDir(dataDir)
	manifestFile := filepath.Join(manifestsDir, "migration-agent-addons-remove.yaml")
	err = os.MkdirAll(manifestsDir, 0755)
	if err != nil {
		return err
	}
	// deploy manifest file
	return ioutil.WriteFile(manifestFile, []byte(yamlContent), 0600)
}

func job() *batch.Job {
	job := &batch.Job{
		TypeMeta: meta.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      "migration-agent-addons-remove",
			Namespace: "kube-system",
		},
		Spec: batch.JobSpec{
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						{
							Name: "network-config",
							VolumeSource: core.VolumeSource{
								ConfigMap: &core.ConfigMapVolumeSource{
									LocalObjectReference: core.LocalObjectReference{
										Name: networkConfigMap,
									},
								},
							},
						},
						{
							Name: "metrics-config",
							VolumeSource: core.VolumeSource{
								ConfigMap: &core.ConfigMapVolumeSource{
									LocalObjectReference: core.LocalObjectReference{
										Name: metricsConfigMap,
									},
								},
							},
						},
						{
							Name: "coredns-config",
							VolumeSource: core.VolumeSource{
								ConfigMap: &core.ConfigMapVolumeSource{
									LocalObjectReference: core.LocalObjectReference{
										Name: corednsConfigMap,
									},
								},
							},
						},
						{
							Name: "ingress-config",
							VolumeSource: core.VolumeSource{
								ConfigMap: &core.ConfigMapVolumeSource{
									LocalObjectReference: core.LocalObjectReference{
										Name: ingrerssConfigMap,
									},
								},
							},
						},
					},
					RestartPolicy: core.RestartPolicyOnFailure,
					Containers: []core.Container{
						{
							Name:            "network-addons-remove",
							Image:           "bitnami/kubectl:latest",
							ImagePullPolicy: core.PullIfNotPresent,
							Command: []string{
								"sh",
								"-c",
								"kubectl delete -f /etc/rke_addon/network/" + networkConfigMap},
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "network-config",
									MountPath: "/etc/rke_addon/network",
								},
							},
						},

						{
							Name:            "dns-addons-remove",
							Image:           "bitnami/kubectl:latest",
							ImagePullPolicy: core.PullIfNotPresent,
							Command: []string{
								"sh",
								"-c",
								"kubectl delete -f /etc/rke_addon/coredns/" + corednsConfigMap},
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "coredns-config",
									MountPath: "/etc/rke_addon/coredns",
								},
							},
						},
						{
							Name:            "ingress-addons-remove",
							Image:           "bitnami/kubectl:latest",
							ImagePullPolicy: core.PullIfNotPresent,
							Command: []string{
								"sh",
								"-c",
								"kubectl delete -f /etc/rke_addon/ingress/" + ingrerssConfigMap},
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "ingress-config",
									MountPath: "/etc/rke_addon/ingress",
								},
							},
						},
						{
							Name:            "metrics-addons-remove",
							Image:           "bitnami/kubectl:latest",
							ImagePullPolicy: core.PullIfNotPresent,
							Command: []string{
								"sh",
								"-c",
								"kubectl delete -f /etc/rke_addon/metrics/" + metricsConfigMap},
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "metrics-config",
									MountPath: "/etc/rke_addon/metrics",
								},
							},
						},
					},
					ServiceAccountName: "migration-agent",
				},
			},
		},
	}

	job.Spec.Template.Spec.HostNetwork = true
	job.Spec.Template.Spec.Tolerations = []core.Toleration{
		{
			Key:    "node.kubernetes.io/not-ready",
			Effect: core.TaintEffectNoSchedule,
		},
		{
			Key:      "node.cloudprovider.kubernetes.io/uninitialized",
			Operator: core.TolerationOpEqual,
			Value:    "true",
			Effect:   core.TaintEffectNoSchedule,
		},
		{
			Key:      "CriticalAddonsOnly",
			Operator: core.TolerationOpExists,
		},
		{
			Key:      "node-role.kubernetes.io/etcd",
			Operator: core.TolerationOpExists,
			Effect:   core.TaintEffectNoExecute,
		},
		{
			Key:      "node-role.kubernetes.io/control-plane",
			Operator: core.TolerationOpExists,
			Effect:   core.TaintEffectNoSchedule,
		},
		{
			Key:      "node-role.kubernetes.io/controlplane",
			Operator: core.TolerationOpExists,
			Effect:   core.TaintEffectNoSchedule,
		},
	}
	job.Spec.Template.Spec.NodeSelector = make(map[string]string)
	job.Spec.Template.Spec.NodeSelector["node-role.kubernetes.io/control-plane"] = "true"

	return job
}

func roleBinding() *rbac.ClusterRoleBinding {
	return &rbac.ClusterRoleBinding{
		TypeMeta: meta.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "migration-agent",
		},
		RoleRef: rbac.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     "cluster-admin",
		},
		Subjects: []rbac.Subject{
			{
				Name:      "migration-agent",
				Kind:      "ServiceAccount",
				Namespace: "kube-system",
			},
		},
	}
}

func serviceAccount() *core.ServiceAccount {
	trueVal := true
	return &core.ServiceAccount{
		TypeMeta: meta.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      "migration-agent",
			Namespace: "kube-system",
		},
		AutomountServiceAccountToken: &trueVal,
	}
}

func objectsToYaml(objs []runtime.Object) (string, error) {
	var result string
	for _, obj := range objs {
		objYaml, err := toYaml(obj)
		if err != nil {
			return result, err
		}
		result += objYaml
	}

	return result, nil
}

func toYaml(obj runtime.Object) (string, error) {
	out, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	yamlString := fmt.Sprintf("%s\n---\n", string(out))
	return yamlString, nil
}

// MigrateAddonsConfig should read the addons configuration and copy it
// as a helm chart config to RKE2 and then save it to the manifest dir.
func MigrateAddonsConfig(ctx context.Context, fullState *cluster.FullState, dataDir string) error {
	coreDNSCfg := fullState.CurrentState.RancherKubernetesEngineConfig.DNS
	rbac := (fullState.CurrentState.RancherKubernetesEngineConfig.Authorization.Mode == "rbac")
	if err := doMigrateCoreDNSAddon(ctx, coreDNSCfg, dataDir, rbac); err != nil {
		return err
	}
	metricsServerCfg := fullState.CurrentState.RancherKubernetesEngineConfig.Monitoring
	if err := doMigrateMetricsServer(ctx, &metricsServerCfg, dataDir, rbac); err != nil {
		return err
	}
	ingressCfg := fullState.CurrentState.RancherKubernetesEngineConfig.Ingress
	return doMigrateNginxIngressAddon(ctx, ingressCfg, dataDir)

}

func doMigrateNginxIngressAddon(ctx context.Context, ingressCfg types.IngressConfig, dataDir string) error {
	if ingressCfg.Provider != "nginx" {
		return nil
	}
	ingressValues := IngressConfig{
		ControllerConfig: IngressControllerConfig{
			Config:            ingressCfg.Options,
			NodeSelector:      ingressCfg.NodeSelector,
			ExtraArgs:         ingressCfg.ExtraArgs,
			ExtraEnvs:         ingressCfg.ExtraEnvs,
			ExtraVolumes:      ingressCfg.ExtraVolumes,
			ExtraVolumeMounts: ingressCfg.ExtraVolumeMounts,
			Tolerations:       ingressCfg.Tolerations,
			DNSPolicy:         ingressCfg.DNSPolicy,
			HostPorts: IngressHostPorts{
				Ports: IngressPorts{
					HTTPPort:  ingressCfg.HTTPPort,
					HTTPSPort: ingressCfg.HTTPSPort,
				},
			},
			PriorityClassName: ingressCfg.NginxIngressControllerPriorityClassName,
		},
		DefaultBackend: DefaultBackendConfig{
			PriorityClassName: ingressCfg.DefaultHTTPBackendPriorityClassName,
			Enabled:           *ingressCfg.DefaultBackend,
		},
	}
	ingressValues.ControllerConfig.UpdateStrategy = &appsv1.DaemonSetUpdateStrategy{
		Type:          ingressCfg.UpdateStrategy.Strategy,
		RollingUpdate: ingressCfg.UpdateStrategy.RollingUpdate,
	}
	helmChartConfig, err := toHelmChartConfig("rke2-"+nginxIngress, ingressValues)
	if err != nil {
		return err
	}

	manifestsDir := manifestsDir(dataDir)
	manifestFile := filepath.Join(manifestsDir, "rke2-"+nginxIngress+"-config.yaml")
	err = os.MkdirAll(manifestsDir, 0700)
	if err != nil {
		return err
	}

	// deploy manifest file
	return ioutil.WriteFile(manifestFile, helmChartConfig, 0600)
}

func doMigrateCoreDNSAddon(ctx context.Context, corednsCfg *types.DNSConfig, dataDir string, rbac bool) error {
	if corednsCfg.Provider != "coredns" {
		return nil
	}
	dnsValues := CoreDNSConfig{
		PriorityClassName: corednsCfg.Options[cluster.CoreDNSPriorityClassNameKey],
		NodeSelector:      corednsCfg.NodeSelector,
		RollingUpdate:     corednsCfg.UpdateStrategy.RollingUpdate,
		Tolerations:       corednsCfg.Tolerations,
		AutoScalerConfig: AutoScalerConfig{
			Enabled:           true,
			PriorityClassName: corednsCfg.Options[cluster.CoreDNSAutoscalerPriorityClassNameKey],
		},
		RBAC: RBACConfig{
			Create: rbac,
		},
	}
	if corednsCfg.LinearAutoscalerParams != nil {
		dnsValues.AutoScalerConfig.CoresPerReplica = corednsCfg.LinearAutoscalerParams.CoresPerReplica
		dnsValues.AutoScalerConfig.NodesPerReplica = corednsCfg.LinearAutoscalerParams.NodesPerReplica
		dnsValues.AutoScalerConfig.Min = corednsCfg.LinearAutoscalerParams.Min
		dnsValues.AutoScalerConfig.Max = corednsCfg.LinearAutoscalerParams.Max
		dnsValues.AutoScalerConfig.PreventSinglePointFailure = corednsCfg.LinearAutoscalerParams.PreventSinglePointFailure
	} else {
		// add the default values in rke1 if params are not set
		dnsValues.AutoScalerConfig.Min = 1
		dnsValues.AutoScalerConfig.CoresPerReplica = 128
		dnsValues.AutoScalerConfig.NodesPerReplica = 4
		dnsValues.AutoScalerConfig.PreventSinglePointFailure = true

	}
	helmChartConfig, err := toHelmChartConfig("rke2-"+coredns, dnsValues)
	if err != nil {
		return err
	}

	manifestsDir := manifestsDir(dataDir)
	manifestFile := filepath.Join(manifestsDir, "rke2-"+coredns+"-config.yaml")
	err = os.MkdirAll(manifestsDir, 0700)
	if err != nil {
		return err
	}

	// deploy manifest file
	return ioutil.WriteFile(manifestFile, helmChartConfig, 0600)
}

func doMigrateMetricsServer(ctx context.Context, metricsCfg *types.MonitoringConfig, dataDir string, rbac bool) error {
	if metricsCfg.Provider != "metrics-server" {
		return nil
	}
	metricsValues := MetricsServerConfig{
		PriorityClassName: metricsCfg.MetricsServerPriorityClassName,
		NodeSelector:      metricsCfg.NodeSelector,
		Tolerations:       metricsCfg.Tolerations,
		Replicas:          int(*metricsCfg.Replicas),
		Args:              mapToSlice(metricsCfg.Options),
		RBAC: RBACConfig{
			Create: rbac,
		},
	}
	helmChartConfig, err := toHelmChartConfig("rke2-"+metricsServer, metricsValues)
	if err != nil {
		return err
	}

	manifestsDir := manifestsDir(dataDir)
	manifestFile := filepath.Join(manifestsDir, "rke2-"+metricsServer+"-config.yaml")
	err = os.MkdirAll(manifestsDir, 0700)
	if err != nil {
		return err
	}

	// deploy manifest file
	return ioutil.WriteFile(manifestFile, helmChartConfig, 0600)
}

func mapToSlice(args map[string]string) []string {
	argsSlice := []string{}
	for k, v := range args {
		argsSlice = append(argsSlice, k+"="+v)
	}
	return argsSlice
}
