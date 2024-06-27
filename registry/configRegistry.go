package registry

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

const globalConfigMapName = "global"

func createConfigName(cName string) string {
	return fmt.Sprintf("%s-config", cName)
}

type configRepository interface {
	get(ctx context.Context) (config.Config, error)
	delete(ctx context.Context) error
	write(ctx context.Context, cfg config.Config) error
	watch(ctx context.Context) (*configWatch, error)
}

type GlobalRegistry struct {
	configRegistry
}

type DoguRegistry struct {
	configRegistry
}

type SensitiveDoguRegistry struct {
	configRegistry
}

type configRegistry struct {
	configReader
	configWriter
	configWatcher
}

func NewGlobalConfigRegistry(ctx context.Context, k8sClient ConfigMapClient) (*GlobalRegistry, error) {
	repo, _ := newConfigRepo(createConfigName(globalConfigMapName), createConfigMapClient(k8sClient, globalConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Data))); lErr != nil {
		return nil, fmt.Errorf("could not create initial global config: %w", lErr)
	}

	return &GlobalRegistry{configRegistry{
		configReader{repo: repo},
		configWriter{repo: repo},
		configWatcher{repo: repo},
	}}, nil
}

func NewDoguConfigRegistry(ctx context.Context, doguName string, k8sClient ConfigMapClient) (*DoguRegistry, error) {
	repo, _ := newConfigRepo(createConfigName(doguName), createConfigMapClient(k8sClient, doguConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Data))); lErr != nil {
		return nil, fmt.Errorf("could not create initial dogu config %s: %w", doguName, lErr)
	}

	return &DoguRegistry{configRegistry{
		configReader{repo: repo},
		configWriter{repo: repo},
		configWatcher{repo: repo},
	}}, nil
}

func NewSensitiveDoguRegistry(ctx context.Context, doguName string, sc SecretClient) (*SensitiveDoguRegistry, error) {
	repo, _ := newConfigRepo(createConfigName(doguName), createSecretClient(sc, sensitiveConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Data))); lErr != nil {
		return nil, fmt.Errorf("could not create initial sensitive dogu config %s: %w", doguName, lErr)
	}

	return &SensitiveDoguRegistry{configRegistry{
		configReader{repo: repo},
		configWriter{repo: repo},
		configWatcher{repo: repo},
	}}, nil
}

type GlobalReader struct {
	configReader
}

type DoguReader struct {
	configReader
}

type SensitiveDoguReader struct {
	configReader
}

func NewGlobalConfigReader(ctx context.Context, k8sClient ConfigMapClient) (*GlobalReader, error) {
	repo, _ := newConfigRepo(createConfigName(globalConfigMapName), createConfigMapClient(k8sClient, globalConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Data))); lErr != nil {
		return nil, fmt.Errorf("could not create initial global config reader: %w", lErr)
	}

	return &GlobalReader{
		configReader{repo: repo},
	}, nil
}

func NewDoguConfigReader(ctx context.Context, doguName string, k8sClient ConfigMapClient) (*DoguReader, error) {
	repo, _ := newConfigRepo(createConfigName(doguName), createConfigMapClient(k8sClient, doguConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Data))); lErr != nil {
		return nil, fmt.Errorf("could not create initial dogu config for reader %s: %w", doguName, lErr)
	}

	return &DoguReader{
		configReader{repo: repo},
	}, nil
}

func NewSensitiveDoguReader(ctx context.Context, doguName string, sc SecretClient) (*SensitiveDoguReader, error) {
	repo, _ := newConfigRepo(createConfigName(doguName), createSecretClient(sc, sensitiveConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Data))); lErr != nil {
		return nil, fmt.Errorf("could not create initial sensitive dogu config %s: %w", doguName, lErr)
	}

	return &SensitiveDoguReader{
		configReader{repo: repo},
	}, nil
}

type GlobalWatcher struct {
	configWatcher
}

type DoguWatcher struct {
	configWatcher
}

type SensitiveDoguWatcher struct {
	configWatcher
}

func NewGlobalConfigWatcher(ctx context.Context, k8sClient ConfigMapClient) (*GlobalWatcher, error) {
	repo, _ := newConfigRepo(createConfigName(globalConfigMapName), createConfigMapClient(k8sClient, globalConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Data))); lErr != nil {
		return nil, fmt.Errorf("could not create initial global config for global watcher: %w", lErr)
	}

	return &GlobalWatcher{
		configWatcher{repo: repo},
	}, nil
}

type DoguConfigWatcherProvider func(ctx context.Context, doguName string) (*DoguWatcher, error)

func (dcwp DoguConfigWatcherProvider) GetDoguConfigWatcher(ctx context.Context, doguName string) (*DoguWatcher, error) {
	return dcwp(ctx, doguName)
}

func NewDoguConfigWatcherProvider(k8sClient ConfigMapClient) DoguConfigWatcherProvider {
	return func(ctx context.Context, doguName string) (*DoguWatcher, error) {
		return NewDoguConfigWatcher(ctx, doguName, k8sClient)
	}
}

func NewDoguConfigWatcher(ctx context.Context, doguName string, k8sClient ConfigMapClient) (*DoguWatcher, error) {
	repo, _ := newConfigRepo(createConfigName(doguName), createConfigMapClient(k8sClient, doguConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Data))); lErr != nil {
		return nil, fmt.Errorf("could not create initial dogu config for watcher %s: %w", doguName, lErr)
	}

	return &DoguWatcher{
		configWatcher{repo: repo},
	}, nil
}

type SensitiveDoguWatcherProvider func(ctx context.Context, doguName string) (*SensitiveDoguWatcher, error)

func (sdcwp SensitiveDoguWatcherProvider) GetSensitiveDoguConfigWatcher(ctx context.Context, doguName string) (*SensitiveDoguWatcher, error) {
	return sdcwp(ctx, doguName)
}

func NewSensitiveDoguWatcherProvider(sc SecretClient) SensitiveDoguWatcherProvider {
	return func(ctx context.Context, doguName string) (*SensitiveDoguWatcher, error) {
		return NewSensitiveDoguWatcher(ctx, doguName, sc)
	}
}

func NewSensitiveDoguWatcher(ctx context.Context, doguName string, sc SecretClient) (*SensitiveDoguWatcher, error) {
	repo, _ := newConfigRepo(createConfigName(doguName), createSecretClient(sc, sensitiveConfigType))

	if lErr := repo.write(ctx, config.CreateConfig(make(config.Data))); lErr != nil {
		return nil, fmt.Errorf("could not create initial sensitive dogu config for watcher %s: %w", doguName, lErr)
	}

	return &SensitiveDoguWatcher{
		configWatcher{repo: repo},
	}, nil
}
