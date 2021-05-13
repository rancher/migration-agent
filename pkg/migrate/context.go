package migrate

import (
	"context"

	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/batch"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/rbac"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/start"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Context struct {
	Batch *batch.Factory
	Apps  *apps.Factory
	Auth  *rbac.Factory
	Core  *core.Factory
	K8s   kubernetes.Interface
	Apply apply.Apply
}

func (c *Context) Start(ctx context.Context) error {
	return start.All(ctx, 5, c.Apps, c.Auth, c.Batch, c.Core)
}

func NewContext(ctx context.Context, restConfig *rest.Config) (*Context, error) {

	k8s := kubernetes.NewForConfigOrDie(restConfig)
	return &Context{
		K8s:   k8s,
		Auth:  rbac.NewFactoryFromConfigOrDie(restConfig),
		Apps:  apps.NewFactoryFromConfigOrDie(restConfig),
		Batch: batch.NewFactoryFromConfigOrDie(restConfig),
		Core:  core.NewFactoryFromConfigOrDie(restConfig),
		Apply: apply.New(k8s, apply.NewClientFactory(restConfig)).WithDynamicLookup(),
	}, nil
}
