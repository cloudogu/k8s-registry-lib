package dogu

import (
	"context"
	"errors"
	"fmt"
	"maps"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/cloudogu/cesapp-lib/core"
	cloudoguerrors "github.com/cloudogu/k8s-registry-lib/errors"
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

type doguVersionRegistry struct {
	configMapClient configMapClient
}

func NewDoguVersionRegistry(configMapClient configMapClient) *doguVersionRegistry {
	return &doguVersionRegistry{
		configMapClient: configMapClient,
	}
}

func (vr *doguVersionRegistry) GetCurrent(ctx context.Context, name SimpleDoguName) (DoguVersion, error) {
	descriptor, err := getDescriptorConfigMapForDogu(ctx, vr.configMapClient, name)
	if err != nil {
		return DoguVersion{}, cloudoguerrors.NewGenericError(err)
	}

	currentDoguVersion, ok := descriptor.Data[currentVersionKey]
	if !ok {
		return DoguVersion{}, getDoguRegistryKeyNotFoundError(currentVersionKey, name)
	}

	version, err := parseDoguVersion(currentDoguVersion, name)
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

func getDescriptorConfigMapForDogu(ctx context.Context, configMapClient configMapClient, simpleDoguName SimpleDoguName) (*corev1.ConfigMap, error) {
	descriptorConfigMapName := getDescriptorConfigMapName(simpleDoguName)
	get, err := configMapClient.Get(ctx, descriptorConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get dogu descriptor config map for dogu %q: %w", simpleDoguName, err)
	}
	return get, nil
}

func (vr *doguVersionRegistry) GetCurrentOfAll(ctx context.Context) ([]DoguVersion, error) {
	registryList, err := getAllDescriptorConfigMaps(ctx, vr.configMapClient)
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
		return doguVersions, cloudoguerrors.NewGenericError(fmt.Errorf("failed to get some dogu versions: %w", err))
	}

	return doguVersions, nil
}

func getAllDescriptorConfigMaps(ctx context.Context, configMapClient configMapClient) (*corev1.ConfigMapList, error) {
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

func (vr *doguVersionRegistry) IsEnabled(ctx context.Context, doguVersion DoguVersion) (bool, error) {
	descriptorConfigMap, err := getDescriptorConfigMapForDogu(ctx, vr.configMapClient, doguVersion.Name)
	if err != nil {
		return false, cloudoguerrors.NewGenericError(err)
	}

	enabledVersion, found := descriptorConfigMap.Data[currentVersionKey]
	if !found || doguVersion.Version.Raw != enabledVersion {
		return false, nil
	}

	return true, nil
}

func (vr *doguVersionRegistry) Enable(ctx context.Context, doguVersion DoguVersion) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// do not create the registry here if not existent because it would be an invalid state without the dogu descriptor.
		descriptorConfigMap, err := getDescriptorConfigMapForDogu(ctx, vr.configMapClient, doguVersion.Name)
		if err != nil {
			return err
		}
		if !isDoguVersionInstalled(*descriptorConfigMap, doguVersion.Version) {
			return fmt.Errorf("dogu descriptor is not available")
		}
		descriptorConfigMap.Data[currentVersionKey] = doguVersion.Version.Raw
		_, err = vr.configMapClient.Update(ctx, descriptorConfigMap, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return cloudoguerrors.NewGenericError(fmt.Errorf("failed to enable dogu %q with version %q: %w", doguVersion.Name, doguVersion.Version.Raw, err))
	}

	return nil
}

func isDoguVersionInstalled(descriptorConfigMap corev1.ConfigMap, version core.Version) bool {
	for key := range descriptorConfigMap.Data {
		if key == version.Raw {
			return true
		}
	}

	return false
}

func (vr *doguVersionRegistry) WatchAllCurrent(ctx context.Context) (<-chan CurrentVersionsWatchResult, error) {
	// Fetch all descriptor ConfigMaps
	list, err := getAllDescriptorConfigMaps(ctx, vr.configMapClient)
	if err != nil {
		return nil, cloudoguerrors.NewGenericError(fmt.Errorf("failed to list initial descriptor configmaps: %w", err))
	}

	persistenceContext, err := createCurrentPersistenceContext(ctx, list.Items)
	if err != nil {
		return nil, cloudoguerrors.NewGenericError(fmt.Errorf("failed to create persistence context for current dogu versions: %w", err))
	}

	retryWatcher, err := createRetryWatcher(ctx, vr, list.ResourceVersion)
	if err != nil {
		return nil, err
	}

	return startWatchInBackground(ctx, retryWatcher, persistenceContext), nil
}

func getWatchFunc(ctx context.Context, vr *doguVersionRegistry) func(options metav1.ListOptions) (watch.Interface, error) {
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		selector := getAllLocalDoguRegistriesSelector()
		options.LabelSelector = selector
		watchInterface, err := vr.configMapClient.Watch(ctx, options)
		if k8serrors.IsGone(err) {
			options.ResourceVersion = ""
			watchInterface, err = vr.configMapClient.Watch(ctx, options)
			if err != nil {
				return nil, fmt.Errorf("failed to create watch after IsGone: %w", err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("failed to create watch: %w", err)
		}

		return watchInterface, nil
	}

	return watchFunc
}

func createRetryWatcher(ctx context.Context, vr *doguVersionRegistry, resourceVersion string) (*toolsWatch.RetryWatcher, error) {
	watchFunc := getWatchFunc(ctx, vr)
	retryWatcher, err := toolsWatch.NewRetryWatcher(resourceVersion, &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		return nil, cloudoguerrors.NewGenericError(fmt.Errorf("failed to create watch for current dogu versions: %w", err))
	}
	return retryWatcher, nil
}

func throwAndLogWatchError(ctx context.Context, err error, resultChannel chan CurrentVersionsWatchResult) {
	logger := log.FromContext(ctx).WithName("DoguVersionRegistry.throwAndLogWatchError")
	logger.Error(err, errMsgWatch)
	resultChannel <- CurrentVersionsWatchResult{
		Err: cloudoguerrors.NewGenericError(err),
	}
}

func startWatchInBackground(ctx context.Context, watchInterface watch.Interface, persistenceContext map[SimpleDoguName]core.Version) <-chan CurrentVersionsWatchResult {
	logger := log.FromContext(ctx).WithName("DoguVersionRegistry.startWatchInBackground")
	currentVersionsWatchResult := make(chan CurrentVersionsWatchResult)

	go func() {
		defer close(currentVersionsWatchResult)
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

				handleEvent(ctx, event, persistenceContext, currentVersionsWatchResult)
			}
		}
	}()

	return currentVersionsWatchResult
}

func handleEvent(ctx context.Context, event watch.Event, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) {
	switch event.Type {
	case watch.Added:
		err := handleAddWatchEvent(ctx, event, persistenceContext, currentVersionsWatchResult)
		if err != nil {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to handle add watch event: %w", err), currentVersionsWatchResult)
		}
	case watch.Modified:
		err := handleModifiedWatchEvent(ctx, event, persistenceContext, currentVersionsWatchResult)
		if err != nil {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to handle modified watch event: %w", err), currentVersionsWatchResult)
		}
	case watch.Deleted:
		err := handleDeleteWatchEvent(ctx, event, persistenceContext, currentVersionsWatchResult)
		if err != nil {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to handle delete watch event: %w", err), currentVersionsWatchResult)
		}
	case watch.Error:
		status, ok := event.Object.(*metav1.Status)
		if !ok {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to cast event object to %T", metav1.Status{}), currentVersionsWatchResult)
		} else {
			throwAndLogWatchError(ctx, fmt.Errorf("watch event type is error: %q", status.String()), currentVersionsWatchResult)
		}
	}
}

func handleDeleteWatchEvent(ctx context.Context, event watch.Event, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) error {
	logger := log.FromContext(ctx).WithName("DoguVersionRegistry.handleDeleteWatchEvent")
	descriptorConfigMap, err := getDescriptorConfigMapFromEvent(event)
	if err != nil {
		return err
	}

	if !hasDoguDescriptorConfigMapCurrentKey(descriptorConfigMap) {
		// disabled dogus deleted. Do nothing
		logger.Info("dogu registry config map without current key was deleted. do nothing.")
		return nil
	}

	eventDoguVersion, err := getCurrentDoguVersionFromDoguDescriptorConfigMap(*descriptorConfigMap)
	if err != nil {
		return err
	}

	oldPersistenceContext := copyPersistenceContext(persistenceContext)
	delete(persistenceContext, eventDoguVersion.Name)

	fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{eventDoguVersion})
	return nil
}

func handleModifiedWatchEvent(ctx context.Context, event watch.Event, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) error {
	logger := log.FromContext(ctx).WithName("DoguVersionRegistry.handleModifiedWatchEvent")
	descriptorConfigMap, err := getDescriptorConfigMapFromEvent(event)
	if err != nil {
		return err
	}

	oldPersistenceContext := copyPersistenceContext(persistenceContext)

	// Skip process. Configmap was possible created empty and will get modified event on Enable.
	if !hasDoguDescriptorConfigMapCurrentKey(descriptorConfigMap) {
		// Dogu was disabled
		doguName := SimpleDoguName(descriptorConfigMap.Labels[doguNameLabelKey])
		version, ok := oldPersistenceContext[doguName]
		if !ok {
			// Dogu ist still disabled and cm got other updates than current deletion
			return nil
		}
		fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{{Name: doguName, Version: version}})
		delete(persistenceContext, doguName)
	} else {
		// Detect change
		eventDoguVersion, getErr := getCurrentDoguVersionFromDoguDescriptorConfigMap(*descriptorConfigMap)
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

func handleAddWatchEvent(ctx context.Context, event watch.Event, persistenceContext map[SimpleDoguName]core.Version, currentVersionsWatchResult chan CurrentVersionsWatchResult) error {
	logger := log.FromContext(ctx).WithName("DoguVersionRegistry.handleAddWatchEvent")
	descriptorConfigMap, err := getDescriptorConfigMapFromEvent(event)
	if err != nil {
		return err
	}

	// Skip process. Configmap was created empty.
	if !hasDoguDescriptorConfigMapCurrentKey(descriptorConfigMap) {
		logger.Info("dogu registry config map was created but without current key. do nothing.")
		return nil
	}

	eventDoguVersion, err := getCurrentDoguVersionFromDoguDescriptorConfigMap(*descriptorConfigMap)
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

func getDescriptorConfigMapFromEvent(event watch.Event) (*corev1.ConfigMap, error) {
	configMap, ok := event.Object.(*corev1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("failed to cast event object to %T. wrong type %T", corev1.ConfigMap{}, event.Object)
	}

	return configMap, nil
}

func getCurrentDoguVersionFromDoguDescriptorConfigMap(cm corev1.ConfigMap) (DoguVersion, error) {
	doguName, ok := cm.Labels[doguNameLabelKey]
	if !ok {
		return DoguVersion{}, fmt.Errorf("dogu descriptor configmap does not contain label %q", doguNameLabelKey)
	}

	currentVersion, ok := cm.Data[currentVersionKey]
	if !ok {
		return DoguVersion{}, fmt.Errorf("dogu descriptor configmap does not contain key %q", currentVersionKey)
	}

	version, err := core.ParseVersion(currentVersion)
	if err != nil {
		return DoguVersion{}, fmt.Errorf("error parsing version %q for dogu version registry %q", currentVersion, cm.Name)
	}

	return DoguVersion{Name: SimpleDoguName(doguName), Version: version}, nil
}

func hasDoguDescriptorConfigMapCurrentKey(cm *corev1.ConfigMap) bool {
	return hasDoguDescriptorConfigMapKey(cm, currentVersionKey)
}

func hasDoguDescriptorConfigMapKey(cm *corev1.ConfigMap, key string) bool {
	if cm != nil && cm.Data != nil {
		_, ok := cm.Data[key]
		return ok
	}

	return false
}

func createCurrentPersistenceContext(ctx context.Context, descriptorConfigMaps []corev1.ConfigMap) (map[SimpleDoguName]core.Version, error) {
	logger := log.FromContext(ctx).WithName("DoguVersionRegistry.createCurrentPersistenceContext")
	currentPersistenceContext := make(map[SimpleDoguName]core.Version)

	var multiErr []error
	for _, cm := range descriptorConfigMaps {
		versionStr, ok := cm.Data[currentVersionKey]
		if !ok {
			logger.Info("got dogu version registry config map without current key. skip create persistence context for it.")
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
