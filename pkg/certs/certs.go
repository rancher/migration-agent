package certs

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rancher/k3s/pkg/daemons/config"
	"github.com/rancher/rke/cluster"
	"github.com/rancher/rke/pki"
	"github.com/sirupsen/logrus"
)

const (
	certType = iota
	keyType
)

func RecoverCertsFromState(ctx context.Context, config *config.Control, state *cluster.FullState) error {
	logrus.Infof("Migrating CA certificates from RKE state file")
	if err := setCertsAndDirs(config); err != nil {
		return err
	}
	if err := writeCertBundle(config.Runtime, state.CurrentState.CertificatesBundle); err != nil {
		return err
	}
	return nil
}

func setCertsAndDirs(cfg *config.Control) error {
	var err error
	cfg.DataDir, err = filepath.Abs(cfg.DataDir)
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Join(cfg.DataDir, "tls"), 0700)
	os.MkdirAll(filepath.Join(cfg.DataDir, "cred"), 0700)
	os.MkdirAll(filepath.Join(cfg.DataDir, "tls", "etcd"), 0700)
	os.MkdirAll(filepath.Join(cfg.DataDir, "db"), 0700)

	cfg.Runtime = &config.ControlRuntime{}

	cfg.Runtime.ClientCA = filepath.Join(cfg.DataDir, "tls", "client-ca.crt")
	cfg.Runtime.ClientCAKey = filepath.Join(cfg.DataDir, "tls", "client-ca.key")
	cfg.Runtime.ServerCA = filepath.Join(cfg.DataDir, "tls", "server-ca.crt")
	cfg.Runtime.ServerCAKey = filepath.Join(cfg.DataDir, "tls", "server-ca.key")
	cfg.Runtime.RequestHeaderCA = filepath.Join(cfg.DataDir, "tls", "request-header-ca.crt")
	cfg.Runtime.RequestHeaderCAKey = filepath.Join(cfg.DataDir, "tls", "request-header-ca.key")
	cfg.Runtime.ServiceKey = filepath.Join(cfg.DataDir, "tls", "service.key")

	cfg.Runtime.ETCDServerCA = filepath.Join(cfg.DataDir, "tls", "etcd", "server-ca.crt")
	cfg.Runtime.ETCDServerCAKey = filepath.Join(cfg.DataDir, "tls", "etcd", "server-ca.key")
	cfg.Runtime.ETCDPeerCA = filepath.Join(cfg.DataDir, "tls", "etcd", "peer-ca.crt")
	cfg.Runtime.ETCDPeerCAKey = filepath.Join(cfg.DataDir, "tls", "etcd", "peer-ca.key")

	return nil
}

func writeCertBundle(runtime *config.ControlRuntime, certBundle map[string]pki.CertificatePKI) error {
	for certName, currentCert := range certBundle {
		switch certName {
		case pki.CACertName:
			if err := writeFile(
				currentCert, certType, runtime.ControlRuntimeBootstrap.ClientCA,
				runtime.ControlRuntimeBootstrap.ETCDPeerCA,
				runtime.ControlRuntimeBootstrap.ETCDServerCA,
				runtime.ControlRuntimeBootstrap.ServerCA); err != nil {
				return err
			}
			if err := writeFile(
				currentCert, keyType, runtime.ControlRuntimeBootstrap.ClientCAKey,
				runtime.ControlRuntimeBootstrap.ETCDPeerCAKey,
				runtime.ControlRuntimeBootstrap.ETCDServerCAKey,
				runtime.ControlRuntimeBootstrap.ServerCAKey); err != nil {
				return err
			}
		case pki.RequestHeaderCACertName:
			if err := writeFile(
				currentCert, certType, runtime.ControlRuntimeBootstrap.RequestHeaderCA); err != nil {
				return err
			}
			if err := writeFile(
				currentCert, keyType, runtime.ControlRuntimeBootstrap.RequestHeaderCAKey); err != nil {
				return err
			}
		case pki.EtcdClientCACertName:
			if err := writeFile(
				currentCert, certType, runtime.ControlRuntimeBootstrap.ETCDServerCA); err != nil {
				return err
			}
			if err := writeFile(
				currentCert, keyType, runtime.ControlRuntimeBootstrap.RequestHeaderCAKey); err != nil {
				return err
			}
		case pki.ServiceAccountTokenKeyName:
			if err := writeFile(
				currentCert, keyType, runtime.ControlRuntimeBootstrap.ServiceKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFile(cert pki.CertificatePKI, fileType int, certPaths ...string) error {
	for _, path := range certPaths {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return errors.Wrapf(err, "failed to mkdir %s", filepath.Dir(path))
		}
		if fileType == certType {
			if err := ioutil.WriteFile(path, []byte(cert.CertificatePEM), 0600); err != nil {
				return errors.Wrapf(err, "failed to write to %s", path)
			}
		} else if fileType == keyType {
			if err := ioutil.WriteFile(path, []byte(cert.KeyPEM), 0600); err != nil {
				return errors.Wrapf(err, "failed to write to %s", path)
			}
		}
	}
	return nil
}
