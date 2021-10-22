package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rancher/k3s/pkg/daemons/config"
	"github.com/rancher/k3s/pkg/etcd"
	"github.com/rancher/migration-agent/pkg/certs"
	migrationconfig "github.com/rancher/migration-agent/pkg/config"
	etcdmigrate "github.com/rancher/migration-agent/pkg/etcd"
	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/pki"
	"github.com/rancher/rke/types"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
)

type Agent struct {
	isETCD             bool
	isControlPlane     bool
	isWorker           bool
	nodeName           string
	fullState          *cluster.FullState
	snapshotPath       string
	dataDir            string
	controlConfig      *config.Control
	sc                 *Context
	disableETCDRestore bool
	registries         []string
}

func (a *Agent) Do(ctx context.Context) error {
	if a.isControlPlane || a.isETCD {
		// certificate restoration
		if err := certs.RecoverCertsFromState(ctx, a.controlConfig, a.fullState); err != nil {
			return err
		}
		if err := migrationconfig.ExportClusterConfiguration(ctx, a.fullState, a.nodeName, true, a.registries); err != nil {
			return err
		}
		if err := migrationconfig.RemoveOldAddons(ctx, a.dataDir); err != nil {
			return err
		}
		if err := migrationconfig.MigrateCNIConfig(ctx, a.fullState, a.dataDir); err != nil {
			return err
		}
		if err := migrationconfig.MigrateAddonsConfig(ctx, a.fullState, a.dataDir); err != nil {
			return err
		}
		if err := migrationconfig.MigrateUserAddonsConfig(ctx, a.fullState, a.dataDir, a.sc.Core.Core().V1().ConfigMap()); err != nil {
			return err
		}
	}

	if a.isETCD && !a.disableETCDRestore {
		// Do snapshot restore on the node
		if err := etcdmigrate.Restore(ctx, a.controlConfig, a.fullState.CurrentState.CertificatesBundle[pki.KubeAPICertName]); err != nil {
			return err
		}
	}

	if a.isWorker && !(a.isControlPlane || a.isETCD) {
		// configure kubeproxy pod to work without rke2 installed
		// by dropping a kubeconfig for kubeproxy
		if err := migrationconfig.ExportKubeProxyConfig(a.fullState, a.dataDir); err != nil {
			return err
		}
	}
	return nil
}

func New(ctx context.Context, sc *Context, config *MigrationConfig, k8sConn bool) (*Agent, error) {
	k3sConfig := get(config)
	snapshotPath := config.Snapshot

	// download s3 config if set
	if config.EtcdS3BucketName != "" {
		logrus.Infof("Downloading s3 snapshot")
		s3, err := etcd.NewS3(ctx, k3sConfig)
		if err != nil {
			return nil, err
		}
		if err := s3.Download(ctx); err != nil {
			return nil, err
		}
		snapshotPath = filepath.Join(config.DataDir, "server", "db", "snapshots", config.Snapshot)
	}
	if _, err := os.Stat(snapshotPath); err != nil {
		return nil, err
	}

	// unzip and extract snapshot
	snapshotDir := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d", decompressedPathPrefix, time.Now().Unix()))
	logrus.Infof("Extracting snapshot to %s", snapshotDir)
	snapshot, fullState, err := extractSnapshot(ctx, k3sConfig.ClusterResetRestorePath, snapshotDir)
	if err != nil {
		return nil, err
	}
	// find the node roles
	node, err := findNode(ctx, fullState, k3sConfig, sc, config.NodeName, k8sConn)
	if err != nil {
		return nil, err
	}
	logrus.Infof("Node found in RKE state file")
	k3sConfig.ClusterResetRestorePath = snapshot

	var worker, etcd, controlplane bool
	for _, role := range node.Role {
		switch role {
		case controlPlaneRole:
			controlplane = true
		case workerRole:
			worker = true
		case etcdRole:
			etcd = true
		}
	}

	return &Agent{
		fullState:          fullState,
		snapshotPath:       snapshot,
		dataDir:            config.DataDir,
		controlConfig:      k3sConfig,
		sc:                 sc,
		isETCD:             etcd,
		isWorker:           worker,
		isControlPlane:     controlplane,
		nodeName:           node.HostnameOverride,
		disableETCDRestore: config.DisableETCDRestore,
		registries:         config.RegistriesTLS,
	}, nil
}

func get(mConfig *MigrationConfig) *config.Control {
	return &config.Control{
		EtcdS3Endpoint:          mConfig.EtcdS3Endpoint,
		EtcdS3EndpointCA:        mConfig.EtcdS3EndpointCA,
		EtcdS3AccessKey:         mConfig.EtcdS3AccessKey,
		EtcdS3SecretKey:         mConfig.EtcdS3SecretKey,
		EtcdS3SkipSSLVerify:     mConfig.EtcdS3SkipSSLVerify,
		EtcdS3BucketName:        mConfig.EtcdS3BucketName,
		EtcdS3Folder:            mConfig.EtcdS3Folder,
		ClusterResetRestorePath: mConfig.Snapshot,
		DataDir:                 filepath.Join(mConfig.DataDir, "server"),
	}
}

func extractSnapshot(ctx context.Context, snapshotPath, snapshotDir string) (string, *cluster.FullState, error) {
	if err := unzip(snapshotPath, snapshotDir); err != nil {
		return "", nil, err
	}
	snapshot, err := findSnapshotFile(snapshotDir)
	if err != nil {
		return "", nil, err
	}
	stateFile, err := findStateFile(snapshotDir)
	if err != nil {
		return "", nil, err
	}
	fullState, err := cluster.ReadStateFile(ctx, stateFile)
	if err != nil {
		return "", nil, err
	}
	return snapshot, fullState, nil
}

func findNode(ctx context.Context, fullState *cluster.FullState, config *config.Control, sc *Context, overrideNodeName string, k8sConn bool) (*types.RKEConfigNode, error) {
	rkeNodes := fullState.CurrentState.RancherKubernetesEngineConfig.Nodes
	// find name by IP, then hostname
	nodeName, nodeIP, err := getHostnameAndIP()
	if err != nil {
		return nil, err
	}
	if overrideNodeName != "" {
		logrus.Infof("Searching in RKE state file for node name %s passed to migration agent", overrideNodeName)
	} else {
		logrus.Infof("Searching for node %s with IP address %s in RKE state file", nodeName, nodeIP)
	}
	for _, node := range rkeNodes {
		if overrideNodeName == node.Address || overrideNodeName == node.InternalAddress || overrideNodeName == node.HostnameOverride {
			return &node, nil
		}
		if node.Address == nodeIP || node.InternalAddress == nodeIP || node.HostnameOverride == nodeName {
			config.PrivateIP = nodeIP
			return &node, nil
		}
	}
	// in case we cant find it by using the address on host we fallback
	// to checking annotations for the private IPs since sometimes
	// public IP is not bound to an interface (eg. ec2 instances)
	logrus.Infof("Searching for node %s in kubernetes node object list", nodeName)
	IPAnnotations := []string{
		flannelPublicIPAnnotation,
		calicoIPAnnotation,
	}
	if k8sConn && sc != nil {
		nodes, err := sc.Core.Core().V1().Node().List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, node := range nodes.Items {
			for _, annotation := range IPAnnotations {
				if v, ok := node.Annotations[annotation]; ok {
					if strings.Contains(v, nodeIP) {
						return k8sNodeToRKENode(&node, rkeNodes)
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("Cant find node in current state")
}

func getHostnameAndIP() (string, string, error) {
	hostIP, err := net.ChooseHostInterface()
	if err != nil {
		return "", "", err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", "", err
	}
	return strings.ToLower(hostname), hostIP.String(), nil
}

func k8sNodeToRKENode(node *v1.Node, rkeNodes []types.RKEConfigNode) (*types.RKEConfigNode, error) {
	for _, rkeNode := range rkeNodes {
		if node.Name == rkeNode.Address || node.Name == rkeNode.HostnameOverride || node.Name == rkeNode.InternalAddress {
			return &rkeNode, nil
		}
	}
	return nil, fmt.Errorf("Failed to find node")
}
