package dogu

import (
	"context"
	"github.com/cloudogu/ces-commons-lib/dogu"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/cloudogu/cesapp-lib/core"
)

type configMapClient interface {
	corev1client.ConfigMapInterface
}

type DoguVersion struct {
	Name    dogu.SimpleName
	Version core.Version
}

type DoguVersionRegistry interface {
	GetCurrent(context.Context, dogu.SimpleName) (DoguVersion, error)
	GetCurrentOfAll(context.Context) ([]DoguVersion, error)
	IsEnabled(context.Context, DoguVersion) (bool, error)
	Enable(context.Context, DoguVersion) error
	WatchAllCurrent(context.Context) (<-chan CurrentVersionsWatchResult, error)
}

type CurrentVersionsWatchResult struct {
	Versions     map[dogu.SimpleName]core.Version
	PrevVersions map[dogu.SimpleName]core.Version
	Diff         []DoguVersion
	Err          error
}

// LocalDoguDescriptorRepository is an append-only Repository, no updates will happen
type LocalDoguDescriptorRepository interface {
	Get(context.Context, DoguVersion) (*core.Dogu, error)
	GetAll(context.Context, []DoguVersion) (map[DoguVersion]*core.Dogu, error)
	Add(context.Context, dogu.SimpleName, *core.Dogu) error
	DeleteAll(context.Context, dogu.SimpleName) error
}
