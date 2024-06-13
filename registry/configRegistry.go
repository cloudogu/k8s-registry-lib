package registry

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
)

const globalConfigMapName = "global"

func createConfigName(doguName string) string {
	return fmt.Sprintf("%s-config", doguName)
}

type ConfigRepository interface {
	get(ctx context.Context) (config.Config, error)
	delete(ctx context.Context) error
	write(ctx context.Context, cfg config.Config) error
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
}

func NewGlobalConfigRegistry(k8sClient ConfigMapClient) *GlobalRegistry {
	repo := newConfigRepo(globalConfigMapName, createConfigMapClient(k8sClient, globalConfigType))
	return &GlobalRegistry{configRegistry{
		configReader{repo: repo},
		configWriter{repo: repo},
	}}
}

func NewDoguConfigRegistry(doguName string, k8sClient ConfigMapClient) *DoguRegistry {
	repo := newConfigRepo(createConfigName(doguName), createConfigMapClient(k8sClient, doguConfigType))
	return &DoguRegistry{configRegistry{
		configReader{repo: repo},
		configWriter{repo: repo},
	}}
}

func NewSensitiveDoguRegistry(sc SecretClient, doguName string) *SensitiveDoguRegistry {
	repo := newConfigRepo(createConfigName(doguName), createSecretClient(sc, sensitiveConfigType))
	return &SensitiveDoguRegistry{configRegistry{
		configReader{repo: repo},
		configWriter{repo: repo},
	}}
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

func NewGlobalConfigReader(k8sClient ConfigMapClient) *GlobalReader {
	repo := newConfigRepo(globalConfigMapName, createConfigMapClient(k8sClient, globalConfigType))
	return &GlobalReader{
		configReader{repo: repo},
	}
}

func NewDoguConfigReader(doguName string, k8sClient ConfigMapClient) *DoguReader {
	repo := newConfigRepo(createConfigName(doguName), createConfigMapClient(k8sClient, doguConfigType))
	return &DoguReader{
		configReader{repo: repo},
	}
}

func NewSensitiveDoguReader(sc SecretClient, doguName string) *SensitiveDoguReader {
	repo := newConfigRepo(createConfigName(doguName), createSecretClient(sc, sensitiveConfigType))
	return &SensitiveDoguReader{
		configReader{repo: repo},
	}
}
