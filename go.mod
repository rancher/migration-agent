module github.com/rancher/migration-agent

go 1.13

replace (
	github.com/containerd/containerd => github.com/k3s-io/containerd v1.4.4-k3s2 // k3s-release/1.4
	github.com/docker/distribution => github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker => github.com/docker/docker v20.10.2+incompatible
	github.com/google/cadvisor => github.com/google/cadvisor v0.39.0
	github.com/kubernetes-sigs/cri-tools => github.com/rancher/cri-tools v1.19.0-k3s1
	github.com/moby/sys/mountinfo => github.com/moby/sys/mountinfo v0.4.0
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.0-rc95
	github.com/opencontainers/runtime-spec => github.com/opencontainers/runtime-spec v1.0.3-0.20200728170252-4d89ac9fbff6
	github.com/rancher/wrangler => github.com/rancher/wrangler v0.6.1
	go.etcd.io/etcd => github.com/k3s-io/etcd v0.5.0-alpha.5.0.20201208200253-50621aee4aea
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
	gopkg.in/square/go-jose.v2 => gopkg.in/square/go-jose.v2 v2.2.2
	k8s.io/api => github.com/k3s-io/kubernetes/staging/src/k8s.io/api v1.21.2-k3s1
	k8s.io/apiextensions-apiserver => github.com/k3s-io/kubernetes/staging/src/k8s.io/apiextensions-apiserver v1.21.2-k3s1
	k8s.io/apimachinery => github.com/k3s-io/kubernetes/staging/src/k8s.io/apimachinery v1.21.2-k3s1
	k8s.io/apiserver => github.com/k3s-io/kubernetes/staging/src/k8s.io/apiserver v1.21.2-k3s1
	k8s.io/cli-runtime => github.com/k3s-io/kubernetes/staging/src/k8s.io/cli-runtime v1.21.2-k3s1
	k8s.io/client-go => github.com/k3s-io/kubernetes/staging/src/k8s.io/client-go v1.21.2-k3s1
	k8s.io/cloud-provider => github.com/k3s-io/kubernetes/staging/src/k8s.io/cloud-provider v1.21.2-k3s1
	k8s.io/cluster-bootstrap => github.com/k3s-io/kubernetes/staging/src/k8s.io/cluster-bootstrap v1.21.2-k3s1
	k8s.io/code-generator => github.com/k3s-io/kubernetes/staging/src/k8s.io/code-generator v1.21.2-k3s1
	k8s.io/component-base => github.com/k3s-io/kubernetes/staging/src/k8s.io/component-base v1.21.2-k3s1
	k8s.io/component-helpers => github.com/k3s-io/kubernetes/staging/src/k8s.io/component-helpers v1.21.2-k3s1
	k8s.io/controller-manager => github.com/k3s-io/kubernetes/staging/src/k8s.io/controller-manager v1.21.2-k3s1
	k8s.io/cri-api => github.com/k3s-io/kubernetes/staging/src/k8s.io/cri-api v1.21.2-k3s1
	k8s.io/csi-translation-lib => github.com/k3s-io/kubernetes/staging/src/k8s.io/csi-translation-lib v1.21.2-k3s1
	k8s.io/kube-aggregator => github.com/k3s-io/kubernetes/staging/src/k8s.io/kube-aggregator v1.21.2-k3s1
	k8s.io/kube-controller-manager => github.com/k3s-io/kubernetes/staging/src/k8s.io/kube-controller-manager v1.21.2-k3s1
	k8s.io/kube-proxy => github.com/k3s-io/kubernetes/staging/src/k8s.io/kube-proxy v1.21.2-k3s1
	k8s.io/kube-scheduler => github.com/k3s-io/kubernetes/staging/src/k8s.io/kube-scheduler v1.21.2-k3s1
	k8s.io/kubectl => github.com/k3s-io/kubernetes/staging/src/k8s.io/kubectl v1.21.2-k3s1
	k8s.io/kubelet => github.com/k3s-io/kubernetes/staging/src/k8s.io/kubelet v1.21.2-k3s1
	k8s.io/kubernetes => github.com/k3s-io/kubernetes v1.21.2-k3s1
	k8s.io/legacy-cloud-providers => github.com/k3s-io/kubernetes/staging/src/k8s.io/legacy-cloud-providers v1.21.2-k3s1
	k8s.io/metrics => github.com/k3s-io/kubernetes/staging/src/k8s.io/metrics v1.21.2-k3s1
	k8s.io/mount-utils => github.com/k3s-io/kubernetes/staging/src/k8s.io/mount-utils v1.21.2-k3s1
	k8s.io/node-api => github.com/k3s-io/kubernetes/staging/src/k8s.io/node-api v1.21.2-k3s1
	k8s.io/sample-apiserver => github.com/k3s-io/kubernetes/staging/src/k8s.io/sample-apiserver v1.21.2-k3s1
	k8s.io/sample-cli-plugin => github.com/k3s-io/kubernetes/staging/src/k8s.io/sample-cli-plugin v1.21.2-k3s1
	k8s.io/sample-controller => github.com/k3s-io/kubernetes/staging/src/k8s.io/sample-controller v1.21.2-k3s1
	mvdan.cc/unparam => mvdan.cc/unparam v0.0.0-20190209190245-fbb59629db34
	sigs.k8s.io/yaml => github.com/kubernetes-sigs/yaml v1.2.0
)

require (
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/bronze1man/goStrongswanVici v0.0.0-20190828090544-27d02f80ba40 // indirect
	github.com/containerd/aufs v0.0.0-20210316121734-20793ff83c97 // indirect
	github.com/containerd/btrfs v0.0.0-20210316141732-918d888fb676 // indirect
	github.com/containerd/cri v1.11.1-0.20200820101445-b0cc07999aa5 // indirect
	github.com/containerd/fifo v0.0.0-20210316144830-115abcc95a1d // indirect
	github.com/containerd/go-runc v0.0.0-20201020171139-16b287bc67d0 // indirect
	github.com/containerd/imgcrypt v1.1.1-0.20210312161619-7ed62a527887 // indirect
	github.com/containerd/nri v0.0.0-20210316161719-dbaa18c31c14 // indirect
	github.com/containerd/zfs v0.0.0-20210315114300-dde8f0fda960 // indirect
	github.com/coreos/flannel v0.12.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/erikdubbelboer/gspt v0.0.0-20190125194910-e68493906b83 // indirect
	github.com/google/cadvisor v0.39.0 // indirect
	github.com/google/tcpproxy v0.0.0-20180808230851-dfa16c61dad2 // indirect
	github.com/kubernetes-sigs/cri-tools v0.0.0-00010101000000-000000000000 // indirect
	github.com/magefile/mage v1.10.0 // indirect
	github.com/minio/minio-go/v7 v7.0.7 // indirect
	github.com/moby/sys v0.0.0-20210311035424-40883be4345c // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/rancher/dynamiclistener v0.2.3 // indirect
	github.com/rancher/k3s v1.21.1-rc1.0.20210701083441-4a6e87e5a207
	github.com/rancher/lasso v0.0.0-20200905045615-7fcb07d6a20b // indirect
	github.com/rancher/remotedialer v0.2.0 // indirect
	github.com/rancher/rke v1.2.7
	github.com/rancher/wrangler v0.6.2
	github.com/rancher/wrangler-api v0.6.0
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/tchap/go-patricia v2.3.0+incompatible // indirect
	github.com/urfave/cli v1.22.2
	go.etcd.io/etcd v0.5.0-alpha.5.0.20201208200253-50621aee4aea // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v1.20.5
	k8s.io/apimachinery v1.20.5
	k8s.io/apiserver v1.20.0 // indirect
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cloud-provider v1.20.0 // indirect
	k8s.io/component-base v1.20.0 // indirect
	k8s.io/controller-manager v1.20.0 // indirect
	k8s.io/cri-api v1.20.0 // indirect
	k8s.io/kubectl v1.20.5 // indirect
	sigs.k8s.io/yaml v1.2.0
)
