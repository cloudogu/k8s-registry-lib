package dogu

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type versionRegistry struct {
	configMapClient configMapClient
}

func NewDoguVersionRegistry(configMapClient configMapClient) *versionRegistry {
	return &versionRegistry{
		configMapClient: configMapClient,
	}
}

func (vr *versionRegistry) GetCurrent(ctx context.Context, name SimpleDoguName) (DoguVersion, error) {
	specConfigMap, err := getSpecConfigMapForDogu(ctx, vr.configMapClient, name)
	if err != nil {
		return DoguVersion{}, err
	}

	currentVersion, ok := specConfigMap.Data[currentVersionKey]
	if !ok {
		return DoguVersion{}, getDoguRegistryKeyNotFoundError(currentVersion, name)
	}

	version, err := parseDoguVersion(currentVersion, name)
	if err != nil {
		return DoguVersion{}, err
	}

	return DoguVersion{Name: name, Version: version}, nil
}

func parseDoguVersion(version string, name SimpleDoguName) (core.Version, error) {
	parsedVersion, err := core.ParseVersion(version)
	if err != nil {
		return core.Version{}, getDoguVersionParseError(version, name, err)
	}
	return parsedVersion, nil
}

func getDoguVersionParseError(currentVersion string, name SimpleDoguName, err error) error {
	return fmt.Errorf("failed to parse version %q for dogu %q: %w", currentVersion, name, err)
}

func getDoguRegistryKeyNotFoundError(key string, name SimpleDoguName) error {
	return fmt.Errorf("failed to get value for key %q for dogu registry %q", key, name)
}

func getSpecConfigMapForDogu(ctx context.Context, configMapClient configMapClient, simpleDoguName SimpleDoguName) (*corev1.ConfigMap, error) {
	specConfigMapName := getSpecConfigMapName(simpleDoguName)
	specConfigMap, err := configMapClient.Get(ctx, specConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get local registry for dogu %q: %w", simpleDoguName, err)
	}

	return specConfigMap, nil
}

func (vr *versionRegistry) GetCurrentOfAll(ctx context.Context) ([]DoguVersion, error) {
	registryList, err := getAllSpecConfigMaps(ctx, vr.configMapClient)
	if err != nil {
		return []DoguVersion{}, err
	}

	var errs []error
	doguVersions := make([]DoguVersion, 0, len(registryList.Items))
	for _, localRegistry := range registryList.Items {
		doguVersion, getErr := vr.GetCurrent(ctx, SimpleDoguName(localRegistry.Labels[doguNameLabelKey]))
		errs = append(errs, getErr)
		doguVersions = append(doguVersions, doguVersion)
	}

	err = errors.Join(errs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get some dogu versions: %w", err)
	}

	return doguVersions, nil
}

func getAllSpecConfigMaps(ctx context.Context, configMapClient configMapClient) (*corev1.ConfigMapList, error) {
	allLocalDoguRegistriesSelector := getAllLocalDoguRegistriesSelector()
	registryList, err := configMapClient.List(ctx, metav1.ListOptions{LabelSelector: allLocalDoguRegistriesSelector})
	if err != nil {
		return nil, fmt.Errorf("failed to list cluster native local dogu registries: %w", err)
	}
	return registryList, err
}

func getAllLocalDoguRegistriesSelector() string {
	return fmt.Sprintf("%s=%s,%s,%s=%s", appLabelKey, appLabelValueCes, doguNameLabelKey, typeLabelKey, typeLabelValueLocalDoguRegistry)
}

func (vr *versionRegistry) IsEnabled(ctx context.Context, name SimpleDoguName) (bool, error) {
	specConfigMap, err := getSpecConfigMapForDogu(ctx, vr.configMapClient, name)
	if err != nil {
		return false, err
	}

	_, enabled := specConfigMap.Data[currentVersionKey]
	return enabled, nil
}

func (vr *versionRegistry) WatchAllCurrent(ctx context.Context) (CurrentVersionsWatch, error) {
	selector := getAllLocalDoguRegistriesSelector()
	watchInterface, err := vr.configMapClient.Watch(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return CurrentVersionsWatch{}, fmt.Errorf("failed to create watches for selector %q: %w", selector, err)
	}

	currentVersionsWatchResult := make(chan CurrentVersionsWatchResult)
	currentVersionsWatch := CurrentVersionsWatch{
		ResultChan: currentVersionsWatchResult,
		cancelFunc: watchInterface.Stop,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				// TODO Logging
				watchInterface.Stop()
				return
			case event, open := <-watchInterface.ResultChan():
				if !open {
					// TODO Logging
					return
				}

				switch event.Type {
				case watch.Added:
					// TODO
					// Check if current exists in added cm
					// If yes, fetch currents from all dogus and return them
					// If no, no channel event needed

					break
				case watch.Modified:
					// TODO
					// cm, ok := event.Object.(*corev1.ConfigMap)

					// Check if current was modified???? This is not possible

					break
				case watch.Deleted:
					// TODO

					break
				case watch.Error:
					// TODO
					return
				}
			}
		}
	}()

	return currentVersionsWatch, nil
}
