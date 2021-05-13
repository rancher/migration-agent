package etcd

import (
	"context"
	"net/http"

	"github.com/rancher/k3s/pkg/daemons/config"
	"github.com/rancher/k3s/pkg/etcd"
	"github.com/rancher/rke/pki"
	"github.com/sirupsen/logrus"
)

func Restore(ctx context.Context, config *config.Control, apiCert pki.CertificatePKI) error {
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
	return nil
}
