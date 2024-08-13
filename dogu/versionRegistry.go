package dogu

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	cloudoguerrors "github.com/cloudogu/k8s-registry-lib/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/util/retry"
	"maps"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	errMsgWatch = "failed to watch dogu registry"

	appLabelKey                     = "app"
	appLabelValueCes                = "ces"
	doguNameLabelKey                = "dogu.name"
	typeLabelKey                    = "k8s.cloudogu.com/type"
	typeLabelValueLocalDoguRegistry = "local-dogu-registry"
	currentVersionKey               = "current"
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
		return DoguVersion{}, cloudoguerrors.NewGenericError(err)
	}

	currentVersion, ok := specConfigMap.Data[currentVersionKey]
	if !ok {
		return DoguVersion{}, getDoguRegistryKeyNotFoundError(currentVersionKey, name)
	}

	version, err := parseDoguVersion(currentVersion, name)
	if err != nil {
		return DoguVersion{}, cloudoguerrors.NewGenericError(err)
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
	return cloudoguerrors.NewNotFoundError(fmt.Errorf("failed to get value for key %q for dogu registry %q", key, name))
}

func getSpecConfigMapForDogu(ctx context.Context, configMapClient configMapClient, simpleDoguName SimpleDoguName) (*corev1.ConfigMap, error) {
	specConfigMapName := getSpecConfigMapName(simpleDoguName)
	get, err := configMapClient.Get(ctx, specConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get dogu spec config map for dogu %q: %w", simpleDoguName, err)
	}
	return get, nil
}

func (vr *versionRegistry) GetCurrentOfAll(ctx context.Context) ([]DoguVersion, error) {
	registryList, err := getAllSpecConfigMaps(ctx, vr.configMapClient)
	if err != nil {
		return nil, cloudoguerrors.NewGenericError(err)
	}

	var errs []error
	doguVersions := make([]DoguVersion, 0, len(registryList.Items))
	for _, localRegistry := range registryList.Items {
		currentVersion, ok := localRegistry.Data[currentVersionKey]
		if !ok {
			continue
		}

		doguName := SimpleDoguName(localRegistry.Labels[doguNameLabelKey])
		version, parseErr := parseDoguVersion(currentVersion, doguName)
		if parseErr != nil {
			errs = append(errs, parseErr)
			continue
		}

		doguVersions = append(doguVersions, DoguVersion{Name: doguName, Version: version})
	}

	err = errors.Join(errs...)
	if err != nil {
		return nil, cloudoguerrors.NewGenericError(fmt.Errorf("failed to get some dogu versions: %w", err))
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
		return false, cloudoguerrors.NewGenericError(err)
	}

	_, enabled := specConfigMap.Data[currentVersionKey]
	return enabled, nil
}

func (vr *versionRegistry) Enable(ctx context.Context, doguVersion DoguVersion) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// do not create the registry here if not existent because it would be an invalid state without the dogu spec.
		specConfigMap, err := getSpecConfigMapForDogu(ctx, vr.configMapClient, doguVersion.Name)
		if err != nil {
			return err
		}
		if !isDoguVersionInstalled(*specConfigMap, doguVersion.Version) {
			return fmt.Errorf("dogu spec is not available")
		}
		specConfigMap.Data[currentVersionKey] = doguVersion.Version.Raw
		_, err = vr.configMapClient.Update(ctx, specConfigMap, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return cloudoguerrors.NewGenericError(fmt.Errorf("failed to enable dogu %q with version %q: %w", doguVersion.Name, doguVersion.Version.Raw, err))
	}

	return nil
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
		return CurrentVersionsWatch{}, cloudoguerrors.NewGenericError(fmt.Errorf("failed to create watches for selector %q: %w", selector, watchErr))
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
			throwAndLogWatchError(ctx, err, currentVersionsWatchResult)
			watchInterface.Stop()
			return
		}
		persistenceContext, err := createCurrentPersistenceContext(list.Items)
		if err != nil {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to create persistent context: %w", err), currentVersionsWatchResult)
			watchInterface.Stop()
			return
		}

		waitForWatchEvents(ctx, watchInterface, persistenceContext, currentVersionsWatchResult)
	}()

	return currentVersionsWatch, nil
}

func throwAndLogWatchError(ctx context.Context, err error, resultChannel chan CurrentVersionsWatchResult) {
	logger := log.FromContext(ctx).WithName("VersionRegistry.WatchAllCurrent")
	logger.Error(err, errMsgWatch)
	resultChannel <- CurrentVersionsWatchResult{
		Err: cloudoguerrors.NewGenericError(err),
	}
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
					throwAndLogWatchError(ctx, fmt.Errorf("failed to handle add watch event: %w", err), currentVersionsWatchResult)
				}
			case watch.Modified:
				err := handleModifiedWatchEvent(ctx, event, persistenceContext, currentVersionsWatchResult)
				if err != nil {
					throwAndLogWatchError(ctx, fmt.Errorf("failed to handle modified watch event: %w", err), currentVersionsWatchResult)
				}
			case watch.Deleted:
				err := handleDeleteWatchEvent(event, persistenceContext, currentVersionsWatchResult)
				if err != nil {
					throwAndLogWatchError(ctx, fmt.Errorf("failed to handle delete watch event: %w", err), currentVersionsWatchResult)
				}
			case watch.Error:
				status, ok := event.Object.(*metav1.Status)
				if !ok {
					throwAndLogWatchError(ctx, fmt.Errorf("failed to cast event object to %T", metav1.Status{}), currentVersionsWatchResult)
					break
				}
				throwAndLogWatchError(ctx, fmt.Errorf("watch event type is error: %q", status.String()), currentVersionsWatchResult)
				return
			}
		}
	}
}

func handleDeleteWatchEvent(event watch.Event, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) error {
	specConfigMap, err := getSpecConfigMapFromEvent(event)
	if err != nil {
		return err
	}

	if !hasDoguSpecConfigMapCurrentKey(specConfigMap) {
		// disabled dogus deleted. Do nothing
		return nil
	}

	eventDoguVersion, err := getCurrentDoguVersionFromDoguSpecConfigMap(*specConfigMap)
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
	specConfigMap, err := getSpecConfigMapFromEvent(event)
	if err != nil {
		return err
	}

	oldPersistenceContext := copyPersistenceContext(persistenceContext)

	// Skip process. Configmap was possible created empty and will get modified event on Enable.
	if !hasDoguSpecConfigMapCurrentKey(specConfigMap) {
		// Dogu was disabled
		doguName := SimpleDoguName(specConfigMap.Labels[doguNameLabelKey])
		version, ok := oldPersistenceContext[doguName]
		if !ok {
			// Dogu ist still disabled and cm got other updates than current deletion
			return nil
		}
		fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{{Name: doguName, Version: version}})
		delete(persistenceContext, doguName)
	} else {
		// Detect change
		eventDoguVersion, getErr := getCurrentDoguVersionFromDoguSpecConfigMap(*specConfigMap)
		if getErr != nil {
			return getErr
		}

		version, ok := persistenceContext[eventDoguVersion.Name]
		if ok && version.IsEqualTo(eventDoguVersion.Version) {
			logger.Info("current versions %s for dogu %s from persistent context and modified event are equal", eventDoguVersion.Version.Raw, eventDoguVersion.Name)
			return nil
		}

		persistenceContext[eventDoguVersion.Name] = eventDoguVersion.Version
		fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{eventDoguVersion})
	}

	return nil
}

func handleAddWatchEvent(event watch.Event, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) error {
	specConfigMap, err := getSpecConfigMapFromEvent(event)
	if err != nil {
		return err
	}

	// Skip process. Configmap was possible created empty and will get modified event on Enable.
	if !hasDoguSpecConfigMapCurrentKey(specConfigMap) {
		return nil
	}

	eventDoguVersion, err := getCurrentDoguVersionFromDoguSpecConfigMap(*specConfigMap)
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

func getSpecConfigMapFromEvent(event watch.Event) (*corev1.ConfigMap, error) {
	configMap, ok := event.Object.(*corev1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("failed to cast event object to %T. wrong type %T", corev1.ConfigMap{}, event.Object)
	}

	return configMap, nil
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

func hasDoguSpecConfigMapCurrentKey(cm *corev1.ConfigMap) bool {
	return hasDoguSpecConfigMapKey(cm, currentVersionKey)
}

func hasDoguSpecConfigMapKey(cm *corev1.ConfigMap, key string) bool {
	if cm != nil && cm.Data != nil {
		_, ok := cm.Data[key]
		return ok
	}

	return false
}

func createCurrentPersistenceContext(specConfigMaps []corev1.ConfigMap) (map[SimpleDoguName]core.Version, error) {
	currentPersistenceContext := make(map[SimpleDoguName]core.Version)

	var multiErr []error
	for _, cm := range specConfigMaps {
		versionStr, ok := cm.Data[currentVersionKey]
		if !ok {
			// TODO Logging
			continue
		}
		doguName := SimpleDoguName(cm.Labels[doguNameLabelKey])
		parseVersion, err := parseDoguVersion(versionStr, doguName)
		if err != nil {
			multiErr = append(multiErr, err)
			continue
		}

		currentPersistenceContext[doguName] = parseVersion
	}

	err := errors.Join(multiErr...)

	return currentPersistenceContext, err
}
