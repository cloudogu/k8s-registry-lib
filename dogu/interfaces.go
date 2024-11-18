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

type DoguVersionRegistry interface {
	GetCurrent(context.Context, dogu.SimpleName) (dogu.QualifiedVersion, error)
	GetCurrentOfAll(context.Context) ([]dogu.QualifiedVersion, error)
	IsEnabled(context.Context, dogu.QualifiedVersion) (bool, error)
	Enable(context.Context, dogu.QualifiedVersion) error
	WatchAllCurrent(context.Context) (<-chan CurrentVersionsWatchResult, error)
}

type CurrentVersionsWatchResult struct {
	Versions     map[dogu.SimpleName]core.Version
	PrevVersions map[dogu.SimpleName]core.Version
	Diff         []dogu.QualifiedVersion
	Err          error
}

// LocalDoguDescriptorRepository is an append-only Repository, no updates will happen
type LocalDoguDescriptorRepository interface {
	Get(context.Context, dogu.QualifiedVersion) (*core.Dogu, error)
	GetAll(context.Context, []dogu.QualifiedVersion) (map[dogu.QualifiedVersion]*core.Dogu, error)
	Add(context.Context, dogu.SimpleName, *core.Dogu) error
	DeleteAll(context.Context, dogu.SimpleName) error
}
