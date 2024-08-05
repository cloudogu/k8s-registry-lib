package dogu

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	errors2 "github.com/cloudogu/k8s-registry-lib/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"maps"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type versionRegistry struct {
	configMapClient           configMapClient
	currentPersistenceContext map[SimpleDoguName]core.Dogu
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
	return errors2.NewNotFoundError(fmt.Errorf("failed to get value for key %q for dogu registry %q", key, name))
}

func getOrCreateSpecConfigMapForDogu(ctx context.Context, configMapClient configMapClient, simpleDoguName SimpleDoguName) (*corev1.ConfigMap, error) {
	specConfigMap, err := getSpecConfigMapForDogu(ctx, configMapClient, simpleDoguName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			var createErr error
			specConfigMap, createErr = createSpecConfigMapForDogu(ctx, configMapClient, simpleDoguName)
			if createErr != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("failed to get local registry for dogu %q: %w", simpleDoguName, err)
		}
	}

	return specConfigMap, nil
}

func getSpecConfigMapForDogu(ctx context.Context, configMapClient configMapClient, simpleDoguName SimpleDoguName) (*corev1.ConfigMap, error) {
	specConfigMapName := getSpecConfigMapName(simpleDoguName)
	get, err := configMapClient.Get(ctx, specConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get dogu spec config map for dogu %s", simpleDoguName)
	}
	return get, nil
}

func createSpecConfigMapForDogu(ctx context.Context, configMapClient configMapClient, simpleDoguName SimpleDoguName) (*corev1.ConfigMap, error) {
	specConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: getSpecConfigMapName(simpleDoguName),
			Labels: map[string]string{
				appLabelKey:      appLabelValueCes,
				doguNameLabelKey: string(simpleDoguName),
				typeLabelKey:     typeLabelValueLocalDoguRegistry,
			},
		},
	}

	_, createErr := configMapClient.Create(ctx, specConfigMap, metav1.CreateOptions{})
	if createErr != nil {
		return nil, fmt.Errorf("failed to create local registry for dogu %q: %w", simpleDoguName, createErr)
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
		return nil, fmt.Errorf("failed to get all cluster native local dogu registries: %w", err)
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

func (vr *versionRegistry) Enable(ctx context.Context, doguVersion DoguVersion) error {
	specConfigMap, err := getOrCreateSpecConfigMapForDogu(ctx, vr.configMapClient, doguVersion.Name)
	if err != nil {
		return err
	}

	if !isDoguVersionInstalled(*specConfigMap, doguVersion.Version) {
		return fmt.Errorf("failed to enable dogu. dogu spec is not available")
	}

	specConfigMap.Data[currentVersionKey] = doguVersion.Version.Raw

	// TODO implement retry
	_, err = vr.configMapClient.Update(ctx, specConfigMap, metav1.UpdateOptions{})
	return err
}

func isDoguVersionInstalled(specConfigMap corev1.ConfigMap, version core.Version) bool {
	for key := range specConfigMap.Data {
		if key == version.Raw {
			return true
		}
	}

	return false
}

func (vr *versionRegistry) WatchAllCurrent(ctx context.Context) (CurrentVersionsWatch, error) {
	selector := getAllLocalDoguRegistriesSelector()
	watchInterface, watchErr := vr.configMapClient.Watch(ctx, metav1.ListOptions{LabelSelector: selector})
	if watchErr != nil {
		return CurrentVersionsWatch{}, fmt.Errorf("failed to create watches for selector %q: %w", selector, watchErr)
	}

	currentVersionsWatchResult := make(chan CurrentVersionsWatchResult)
	currentVersionsWatch := CurrentVersionsWatch{
		ResultChan: currentVersionsWatchResult,
		cancelFunc: watchInterface.Stop,
	}

	go func() {
		// Fetch all specConfigMaps
		list, err := getAllSpecConfigMaps(ctx, vr.configMapClient)
		if err != nil {
			// TODO Log error
			return
		}
		persistenceContext, err := createCurrentPersistenceContext(list.Items)
		if err != nil {
			// TODO Log error
			return
		}

		waitForWatchEvents(ctx, watchInterface, persistenceContext, currentVersionsWatchResult)
	}()

	return currentVersionsWatch, nil
}

func waitForWatchEvents(ctx context.Context, watchInterface watch.Interface, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) {
	logger := log.FromContext(ctx).WithName("VersionRegistry.waitForWatchEvents")
	for {
		select {
		case <-ctx.Done():
			watchInterface.Stop()
			logger.Info("context canceled. Stop watch channel.")
			return
		case event, open := <-watchInterface.ResultChan():
			if !open {
				logger.Info("watch channel canceled. Stop watch.")
				return
			}
			switch event.Type {
			case watch.Added:
				err := handleAddWatchEvent(event, persistenceContext, currentVersionsWatchResult)
				if err != nil {
					logger.Error(err, "failed to handle add watch event")
				}
				break
			case watch.Modified:
				err := handleModifiedWatchEvent(ctx, event, persistenceContext, currentVersionsWatchResult)
				if err != nil {
					logger.Error(err, "failed to handle modified watch event")
				}
				break
			case watch.Deleted:
				err := handleDeleteWatchEvent(event, persistenceContext, currentVersionsWatchResult)
				if err != nil {
					logger.Error(err, "failed to handle delete watch event")
				}
				break
			case watch.Error:
				// TODO Map errors
				status, ok := event.Object.(*metav1.Status)
				if !ok {
					logger.Error(fmt.Errorf("failed to cast event object to %T", metav1.Status{}), "failed to handle error watch event")
					break
				}

				logger.Error(fmt.Errorf("failed to handle error watch event"), status.String())

				return
			}
		}
	}
}

func handleDeleteWatchEvent(event watch.Event, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) error {
	eventDoguVersion, err := getCurrentDoguVersionFromEvent(event)
	if err != nil {
		return err
	}

	oldPersistenceContext := copyPersistenceContext(persistenceContext)
	delete(persistenceContext, eventDoguVersion.Name)

	fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{eventDoguVersion})
	return nil
}

func handleModifiedWatchEvent(ctx context.Context, event watch.Event, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) error {
	logger := log.FromContext(ctx).WithName("VersionRegistry.handleModifiedWatchEvent")
	eventDoguVersion, err := getCurrentDoguVersionFromEvent(event)
	if err != nil {
		return err
	}

	// Detect change
	version, ok := persistenceContext[eventDoguVersion.Name]
	if ok && version.IsEqualTo(eventDoguVersion.Version) {
		logger.Info("current versions %s for dogu %s from persistent context and modified event are equal", eventDoguVersion.Version.Raw, eventDoguVersion.Name)
		return nil
	}

	oldPersistenceContext := copyPersistenceContext(persistenceContext)
	persistenceContext[eventDoguVersion.Name] = eventDoguVersion.Version

	fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{eventDoguVersion})
	return nil
}

func handleAddWatchEvent(event watch.Event, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) error {
	eventDoguVersion, err := getCurrentDoguVersionFromEvent(event)
	if err != nil {
		return err
	}

	oldPersistenceContext := copyPersistenceContext(persistenceContext)
	persistenceContext[eventDoguVersion.Name] = eventDoguVersion.Version
	fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{eventDoguVersion})
	return nil
}

func copyPersistenceContext(persistenceContext map[SimpleDoguName]core.Version) map[SimpleDoguName]core.Version {
	oldPersistenceContext := make(map[SimpleDoguName]core.Version, len(persistenceContext))
	maps.Copy(oldPersistenceContext, persistenceContext)

	return oldPersistenceContext
}

func fireWatchResult(channel chan CurrentVersionsWatchResult, prevPersistenceContext map[SimpleDoguName]core.Version, newPersistenceContext map[SimpleDoguName]core.Version, diffs []DoguVersion) {
	fireWatchResultWithError(channel, prevPersistenceContext, newPersistenceContext, diffs, nil)
}

func fireWatchResultWithError(channel chan CurrentVersionsWatchResult, prevPersistenceContext map[SimpleDoguName]core.Version, newPersistenceContext map[SimpleDoguName]core.Version, diffs []DoguVersion, err error) {
	result := CurrentVersionsWatchResult{
		PrevVersions: prevPersistenceContext,
		Versions:     newPersistenceContext,
		Diff:         diffs,
		Err:          err,
	}

	channel <- result
}

func getCurrentDoguVersionFromEvent(event watch.Event) (DoguVersion, error) {
	configMap, ok := event.Object.(*corev1.ConfigMap)
	if !ok {
		return DoguVersion{}, fmt.Errorf("failed to cast event object to %T. wrong type %T", corev1.ConfigMap{}, event.Object)
	}

	eventDoguVersion, err := getCurrentDoguVersionFromDoguSpecConfigMap(*configMap)
	if err != nil {
		return DoguVersion{}, err
	}

	return eventDoguVersion, nil
}

func getCurrentDoguVersionFromDoguSpecConfigMap(cm corev1.ConfigMap) (DoguVersion, error) {
	doguName, ok := cm.Labels[doguNameLabelKey]
	if !ok {
		return DoguVersion{}, fmt.Errorf("dogu spec configmap does not contain label %q", doguNameLabelKey)
	}

	currentVersion, ok := cm.Data[currentVersionKey]
	if !ok {
		return DoguVersion{}, fmt.Errorf("dogu spec configmap does not contain key %q", currentVersionKey)
	}

	version, err := core.ParseVersion(currentVersion)
	if err != nil {
		return DoguVersion{}, fmt.Errorf("error parsing version %q for dogu version registry %q", currentVersion, cm.Name)
	}

	return DoguVersion{Name: SimpleDoguName(doguName), Version: version}, nil
}

func createCurrentPersistenceContext(specConfigMaps []corev1.ConfigMap) (map[SimpleDoguName]core.Version, error) {
	currentPersistenceContext := make(map[SimpleDoguName]core.Version)

	var multiErr []error
	for _, cm := range specConfigMaps {
		doguStr, ok := cm.Data[currentVersionKey]
		if !ok {
			continue
		}
		doguName := SimpleDoguName(cm.Labels[doguNameLabelKey])
		dogu, err := unmarshalDoguJsonStr(doguStr, doguName, currentVersionKey)
		if err != nil {
			multiErr = append(multiErr, err)
			continue
		}

		version, err := dogu.GetVersion()
		if err != nil {
			multiErr = append(multiErr, err)
			continue
		}

		currentPersistenceContext[doguName] = version
	}

	err := errors.Join(multiErr...)

	return currentPersistenceContext, err
}
