package dogu

import (
	"context"
	"github.com/cloudogu/ces-commons-lib/dogu"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/cloudogu/cesapp-lib/core"
)

// LocalRegistry abstracts accessing various backends for reading and writing dogu specs (dogu.json).
type LocalRegistry interface {
	// Enable makes the dogu spec reachable.
	Enable(ctx context.Context, dogu *core.Dogu) error
	// Register adds the given dogu spec to the local registry.
	Register(ctx context.Context, dogu *core.Dogu) error
	// UnregisterAllVersions deletes all versions of the dogu spec from the local registry and makes the spec unreachable.
	UnregisterAllVersions(ctx context.Context, SimpleName string) error
	// GetCurrent retrieves the spec of the referenced dogu's currently installed version.
	GetCurrent(ctx context.Context, SimpleName string) (*core.Dogu, error)
	// GetCurrentOfAll retrieves the specs of all dogus' currently installed versions.
	GetCurrentOfAll(ctx context.Context) ([]*core.Dogu, error)
	// IsEnabled checks if the current spec of the referenced dogu is reachable.
	IsEnabled(ctx context.Context, SimpleName string) (bool, error)
}

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
