package registry

import (
	"context"
	"fmt"
	"reflect"
)

type ConfigRegistryProvider[T any] func(ctx context.Context, name string) (T, error)

func (crp ConfigRegistryProvider[T]) GetConfig(ctx context.Context, name string) (T, error) {
	return crp(ctx, name)
}

func NewDoguConfigRegistryProvider[T any](k8sClient ConfigMapClient) (ConfigRegistryProvider[T], error) {
	sanityReg := &DoguRegistry{}

	_, ok := any(sanityReg).(T)
	if !ok {
		return nil, fmt.Errorf("the used interface %s is not compatible with registry %T", reflect.TypeOf(new(T)).Elem().Name(), sanityReg)
	}

	return func(ctx context.Context, doguName string) (T, error) {
		reg, err := NewDoguConfigRegistry(ctx, doguName, k8sClient)
		if err != nil {
			return *new(T), fmt.Errorf("could not create new dogu config registry: %w", err)
		}

		// As we already did a sanity check, we can force the cast here and let it panic in case of error.
		return any(reg).(T), nil

	}, nil
}

type DefaultDoguConfigRegistryProvider func(ctx context.Context, doguName string) (ConfigurationRegistry, error)

func (dcp DefaultDoguConfigRegistryProvider) GetConfig(ctx context.Context, doguName string) (ConfigurationRegistry, error) {
	return dcp(ctx, doguName)
}

func NewDefaultDoguConfigRegistryProvider(k8sClient ConfigMapClient) (DefaultDoguConfigRegistryProvider, error) {
	return func(ctx context.Context, doguName string) (ConfigurationRegistry, error) {
		provider, err := NewDoguConfigRegistry(ctx, doguName, k8sClient)
		if err != nil {
			return nil, fmt.Errorf("unable to create underlying config provider: %w", err)
		}

		return provider, nil
	}, nil
}
