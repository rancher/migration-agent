package etcd

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rancher/k3s/pkg/daemons/config"
	"github.com/rancher/k3s/pkg/etcd"
	"github.com/rancher/rke/pki"
	"github.com/sirupsen/logrus"
)

func Restore(ctx context.Context, config *config.Control, apiCert pki.CertificatePKI) error {
	if _, err := os.Stat(apiCert.Path); err != nil {
		if err := ioutil.WriteFile(apiCert.Path, []byte(apiCert.CertificatePEM), 0600); err != nil {
			return err
		}
	}
	if _, err := os.Stat(apiCert.KeyPath); err != nil {
		if err := ioutil.WriteFile(apiCert.KeyPath, []byte(apiCert.KeyPEM), 0600); err != nil {
			return err
		}
	}
	etcdNew := etcd.NewETCD()

	// setting up certs for the etcd client so that register passes
	config.Runtime.ClientETCDCert = apiCert.Path
	config.Runtime.ClientETCDKey = apiCert.KeyPath

	_, err := etcdNew.Register(ctx, config, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
	if err != nil {
		logrus.Info(err)
		return err
	}

	if err := etcdNew.Restore(ctx); err != nil {
		logrus.Info(err)
		return err
	}

	// wite the tombstone file to db dir
	tombstoneFile := filepath.Join(config.DataDir, "db", "tombstone")
	if err := ioutil.WriteFile(tombstoneFile, []byte{}, 0600); err != nil {
		logrus.Fatalf("failed to write tombstone file to %s", tombstoneFile)
	}

	return nil
}
