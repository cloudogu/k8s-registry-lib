package repository

import (
	"context"

	"github.com/cloudogu/ces-commons-lib/dogu"
	"github.com/cloudogu/k8s-registry-lib/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type generalConfigRepository interface {
	get(context.Context, configName) (config.Config, error)
	delete(context.Context, configName) error
	create(context.Context, configName, dogu.SimpleName, config.Config) (config.Config, error)
	update(context.Context, configName, dogu.SimpleName, config.Config) (config.Config, error)
	setOwnerReference(ctx context.Context, name configName, owners []metav1.OwnerReference) error
	saveOrMerge(context.Context, configName, config.Config) (config.Config, error)
	watch(ctx context.Context, name configName, filters ...config.WatchFilter) (<-chan configWatchResult, error)
}

type resourceMetaAccessor interface {
	GetResourceVersion() string
	GetCreationTimestamp() metav1.Time
	GetManagedFields() []metav1.ManagedFieldsEntry
}

type configClient interface {
	Get(ctx context.Context, name string) (clientData, error)
	GetWithListResourceVersion(ctx context.Context, name string) (clientData, string, error)
	Delete(ctx context.Context, name string) error
	Create(ctx context.Context, name string, doguName string, dataStr string) (resourceMetaAccessor, error)
	Update(ctx context.Context, pCtx string, name string, doguName string, dataStr string) (resourceMetaAccessor, error)
	UpdateClientData(ctx context.Context, update clientData) (resourceMetaAccessor, error)
	Watch(ctx context.Context, name string, resourceVersion string) (<-chan clientWatchResult, error)
	SetOwnerReference(ctx context.Context, cmName string, owner []metav1.OwnerReference) (resourceMetaAccessor, error)
}

type k8sClient interface {
	client.Client
}
