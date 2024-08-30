package dogu

import (
	"context"
	"k8s.io/client-go/tools/cache"

	informerCore "k8s.io/client-go/informers/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/cloudogu/cesapp-lib/core"
)

type configMapClient interface {
	corev1client.ConfigMapInterface
}

type configMapInformer interface {
	informerCore.ConfigMapInformer
}

type sharedInformer interface {
	cache.SharedIndexInformer
}

type SimpleDoguName string

// in common lib
type DoguVersion struct {
	Name    SimpleDoguName
	Version core.Version
}

type DoguVersionRegistry interface {
	GetCurrent(context.Context, SimpleDoguName) (DoguVersion, error)
	GetCurrentOfAll(context.Context) ([]DoguVersion, error)
	IsEnabled(context.Context, DoguVersion) (bool, error)
	Enable(context.Context, DoguVersion) error
	WatchAllCurrent(context.Context) (<-chan CurrentVersionsWatchResult, error)
}

type CurrentVersionsWatchResult struct {
	Versions     map[SimpleDoguName]core.Version
	PrevVersions map[SimpleDoguName]core.Version
	Diff         []DoguVersion
	Err          error
}

// LocalDoguDescriptorRepository is an append-only Repository, no updates will happen
type LocalDoguDescriptorRepository interface {
	Get(context.Context, DoguVersion) (*core.Dogu, error)
	GetAll(context.Context, []DoguVersion) (map[DoguVersion]*core.Dogu, error)
	Add(context.Context, SimpleDoguName, *core.Dogu) error
	DeleteAll(context.Context, SimpleDoguName) error
}
