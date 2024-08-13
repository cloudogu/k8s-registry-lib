package dogu

import (
	"context"
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
	UnregisterAllVersions(ctx context.Context, simpleDoguName string) error
	// GetCurrent retrieves the spec of the referenced dogu's currently installed version.
	GetCurrent(ctx context.Context, simpleDoguName string) (*core.Dogu, error)
	// GetCurrentOfAll retrieves the specs of all dogus' currently installed versions.
	GetCurrentOfAll(ctx context.Context) ([]*core.Dogu, error)
	// IsEnabled checks if the current spec of the referenced dogu is reachable.
	IsEnabled(ctx context.Context, simpleDoguName string) (bool, error)
}

type configMapClient interface {
	corev1client.ConfigMapInterface
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
	WatchAllCurrent(context.Context) (CurrentVersionsWatch, error)
}

type CurrentVersionsWatch struct {
	ResultChan <-chan CurrentVersionsWatchResult
	cancelFunc context.CancelFunc
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
	GetAll(context.Context, []DoguVersion) map[DoguVersion]*core.Dogu
	Add(context.Context, SimpleDoguName, *core.Dogu) error
	DeleteAll(context.Context, SimpleDoguName) error
}
