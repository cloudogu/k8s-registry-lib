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

func NewSensitiveDoguConfigRegistryProvider[T any](k8sClient SecretClient) (ConfigRegistryProvider[T], error) {
	sanityReg := &SensitiveDoguRegistry{}

	_, ok := any(sanityReg).(T)
	if !ok {
		return nil, fmt.Errorf("the used interface %s is not compatible with registry %T", reflect.TypeOf(new(T)).Elem().Name(), sanityReg)
	}

	return func(ctx context.Context, doguName string) (T, error) {
		reg, err := NewSensitiveDoguRegistry(ctx, doguName, k8sClient)
		if err != nil {
			return *new(T), fmt.Errorf("could not create new sensitive dogu config registry: %w", err)
		}

		// As we already did a sanity check, we can force the cast here and let it panic in case of error.
		return any(reg).(T), nil

	}, nil
}

type DefaultSensitiveDoguConfigRegistryProvider DefaultDoguConfigRegistryProvider

func (dcp DefaultSensitiveDoguConfigRegistryProvider) GetConfig(ctx context.Context, doguName string) (ConfigurationRegistry, error) {
	return dcp(ctx, doguName)
}

func NewDefaultSensitiveDoguConfigRegistryProvider(k8sClient SecretClient) (DefaultSensitiveDoguConfigRegistryProvider, error) {
	return func(ctx context.Context, doguName string) (ConfigurationRegistry, error) {
		provider, err := NewSensitiveDoguRegistry(ctx, doguName, k8sClient)
		if err != nil {
			return nil, fmt.Errorf("unable to create underlying config registry: %w", err)
		}

		return provider, nil
	}, nil
}

type DoguConfigWatcherProvider[T any] func(ctx context.Context, doguName string) (T, error)

func (dcwp DoguConfigWatcherProvider[T]) GetDoguConfigWatcher(ctx context.Context, doguName string) (T, error) {
	return dcwp(ctx, doguName)
}

func NewDoguConfigWatcherProvider[T any](k8sClient ConfigMapClient) (DoguConfigWatcherProvider[T], error) {
	sanityWatcher := &DoguWatcher{}

	_, ok := any(sanityWatcher).(T)
	if !ok {
		return nil, fmt.Errorf("the used interface %s is not compatible with watcher %T", reflect.TypeOf(new(T)).Elem().Name(), sanityWatcher)
	}

	return func(ctx context.Context, doguName string) (T, error) {
		reg, err := NewDoguConfigWatcher(ctx, doguName, k8sClient)
		if err != nil {
			return *new(T), fmt.Errorf("could not create new dogu config watcher: %w", err)
		}

		// As we already did a sanity check, we can force the cast here and let it panic in case of error.
		return any(reg).(T), nil

	}, nil
}

type DefaultDoguConfigWatcherProvider func(ctx context.Context, doguName string) (ConfigurationWatcher, error)

func (dcwp DefaultDoguConfigWatcherProvider) GetDoguConfigWatcher(ctx context.Context, doguName string) (ConfigurationWatcher, error) {
	return dcwp(ctx, doguName)
}

func NewDefaultDoguConfigWatcherProvider(k8sClient ConfigMapClient) (DefaultDoguConfigWatcherProvider, error) {
	return func(ctx context.Context, doguName string) (ConfigurationWatcher, error) {
		provider, err := NewDoguConfigWatcher(ctx, doguName, k8sClient)
		if err != nil {
			return nil, fmt.Errorf("unable to create underlying watch provider: %w", err)
		}

		return provider, nil
	}, nil
}

type SensitiveDoguWatcherProvider[T any] func(ctx context.Context, doguName string) (T, error)

func (sdcwp SensitiveDoguWatcherProvider[T]) GetSensitiveDoguConfigWatcher(ctx context.Context, doguName string) (T, error) {
	return sdcwp(ctx, doguName)
}

func NewSensitiveDoguWatcherProvider[T any](sc SecretClient) (SensitiveDoguWatcherProvider[T], error) {
	sanitySensitiveWatcher := &SensitiveDoguWatcher{}

	_, ok := any(sanitySensitiveWatcher).(T)
	if !ok {
		return nil, fmt.Errorf("the used interface %s is not compatible with watcher %T", reflect.TypeOf(new(T)).Elem().Name(), sanitySensitiveWatcher)
	}

	return func(ctx context.Context, doguName string) (T, error) {
		reg, err := NewSensitiveDoguWatcher(ctx, doguName, sc)
		if err != nil {
			return *new(T), fmt.Errorf("could not create new sensitive dogu config watcher: %w", err)
		}

		// As we already did a sanity check, we can force the cast here and let it panic in case of error.
		return any(reg).(T), nil

	}, nil
}

type DefaultSensitiveDoguWatcherProvider DefaultDoguConfigWatcherProvider

func (sdcwp DefaultSensitiveDoguWatcherProvider) GetSensitiveDoguConfigWatcher(ctx context.Context, doguName string) (ConfigurationWatcher, error) {
	return sdcwp(ctx, doguName)
}

func NewDefaultSensitiveDoguWatcherProvider(sc SecretClient) (DefaultSensitiveDoguWatcherProvider, error) {
	return func(ctx context.Context, doguName string) (ConfigurationWatcher, error) {
		provider, err := NewSensitiveDoguWatcher(ctx, doguName, sc)
		if err != nil {
			return nil, fmt.Errorf("unable to create watcher: %w", err)
		}

		return provider, nil
	}, nil
}
