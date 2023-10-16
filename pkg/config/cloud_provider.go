package config

import (
	"os"
	"path/filepath"

	"github.com/rancher/rke/cluster"
	"github.com/sirupsen/logrus"
)

const (
	cloudConfigFileRKE1     = "/etc/kubernetes/cloud-config"
	cloudConfigFileRKE2     = "/etc/rancher/rke2/cloud.conf"
	cloudProviderNameFlag   = "cloud-provider-name"
	cloudProviderConfigFlag = "cloud-provider-config"
)

func migrateCloudProviders(fullState *cluster.FullState, args map[string]string) error {
	cloudProviderName := fullState.CurrentState.RancherKubernetesEngineConfig.CloudProvider.Name
	if cloudProviderName == "" {
		return nil
	}
	logrus.Infof("Migrating RKE cloud provider config")
	// add cloud config name to the args
	args[cloudProviderNameFlag] = cloudProviderName
	if _, err := os.Stat(cloudConfigFileRKE1); err == nil {
		// copy cloud config file to the rke2 location
		if err := copy(cloudConfigFileRKE1, cloudConfigFileRKE2); err != nil {
			return err
		}
		// add cloud config file to the args
		args[cloudProviderConfigFlag] = cloudConfigFileRKE2
	}
	return nil
}

// copy will copy the src file to destination and will create the base directory
// of the destination first.
func copy(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return err
	}
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dest, input, 0600)
}
