package dogu

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	cloudoguerrors "github.com/cloudogu/k8s-registry-lib/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
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

type doguVersionRegistry struct {
	configMapClient   configMapClient
	configMapInformer configMapInformer
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
	informer := vr.configMapInformer.Informer()

	currentVersionsWatchResult := make(chan CurrentVersionsWatchResult)

	go func() {
		newCtx, cancelFunc := context.WithCancel(ctx)
		defer cancelFunc()

		// Fetch all descriptor ConfigMaps
		list, err := getAllDescriptorConfigMaps(newCtx, vr.configMapClient)
		if err != nil {
			throwAndLogWatchError(newCtx, err, currentVersionsWatchResult)
			return
		}
		persistenceContext, err := createCurrentPersistenceContext(newCtx, list.Items)
		if err != nil {
			throwAndLogWatchError(newCtx, fmt.Errorf("error during persistence context creation. watch is still active: %w", err), currentVersionsWatchResult)
		}

		err = waitForWatchEvents(newCtx, informer, persistenceContext, currentVersionsWatchResult)
		if err != nil {
			throwAndLogWatchError(newCtx, err, currentVersionsWatchResult)
		}
	}()

	return currentVersionsWatchResult, nil
}

func throwAndLogWatchError(ctx context.Context, err error, resultChannel chan CurrentVersionsWatchResult) {
	logger := log.FromContext(ctx).WithName("DoguVersionRegistry.throwAndLogWatchError")
	logger.Error(err, errMsgWatch)
	resultChannel <- CurrentVersionsWatchResult{
		Err: cloudoguerrors.NewGenericError(err),
	}
}

func waitForWatchEvents(
	ctx context.Context,
	informer sharedInformer,
	persistenceContext map[SimpleDoguName]core.Version,
	currentVersionsWatchResult chan CurrentVersionsWatchResult,
) error {
	selectorString := getAllLocalDoguRegistriesSelector()
	selector, err := labels.Parse(selectorString)
	if err != nil {
		return fmt.Errorf("failed to parse label selector for dogu registries: %s: %w", selectorString, err)
	}

	_, err = informer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			configMap, err := toConfigMap(obj)
			if err != nil {
				return false
			}

			return selector.Matches(labels.Set(configMap.Labels))

			// TODO IDEA: use filter function to check if configmap contains current key
			// that way, we'll only have to fire the appropriate watch events in the handler
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    handleAdd(ctx, persistenceContext, currentVersionsWatchResult),
			UpdateFunc: handleUpdate(ctx, persistenceContext, currentVersionsWatchResult),
			DeleteFunc: handleDelete(ctx, persistenceContext, currentVersionsWatchResult),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add event handler for current versions: %w", err)
	}

	informer.Run(ctx.Done())
	return nil
}

func handleDelete(
	ctx context.Context,
	persistenceContext map[SimpleDoguName]core.Version,
	currentVersionsWatchResult chan CurrentVersionsWatchResult,
) func(obj interface{}) {
	return func(obj interface{}) {
		logger := log.FromContext(ctx).WithName("DoguVersionRegistry.handleDelete")
		descriptorConfigMap, err := toConfigMap(obj)
		if err != nil {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to handle delete watch event: %w", err), currentVersionsWatchResult)
			return
		}

		if !hasDoguDescriptorConfigMapCurrentKey(descriptorConfigMap) {
			// disabled dogus deleted. Do nothing
			logger.Info("dogu registry config map without current key was deleted. do nothing.")
			return
		}

		eventDoguVersion, err := getCurrentDoguVersionFromDoguDescriptorConfigMap(*descriptorConfigMap)
		if err != nil {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to handle delete watch event: %w", err), currentVersionsWatchResult)
			return
		}

		oldPersistenceContext := copyPersistenceContext(persistenceContext)
		delete(persistenceContext, eventDoguVersion.Name)

		fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{eventDoguVersion})
	}
}

func handleUpdate(
	ctx context.Context,
	persistenceContext map[SimpleDoguName]core.Version,
	currentVersionsWatchResult chan CurrentVersionsWatchResult,
) func(oldObj, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		logger := log.FromContext(ctx).WithName("DoguVersionRegistry.handleUpdate")
		descriptorConfigMap, err := toConfigMap(newObj)
		if err != nil {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to handle update watch event: %w", err), currentVersionsWatchResult)
			return
		}

		oldPersistenceContext := copyPersistenceContext(persistenceContext)

		// Skip process. Configmap was possible created empty and will get modified event on Enable.
		if !hasDoguDescriptorConfigMapCurrentKey(descriptorConfigMap) {
			// Dogu was disabled
			doguName := SimpleDoguName(descriptorConfigMap.Labels[doguNameLabelKey])
			version, ok := oldPersistenceContext[doguName]
			if !ok {
				// Dogu ist still disabled and cm got other updates than current deletion
				return
			}
			fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{{Name: doguName, Version: version}})
			delete(persistenceContext, doguName)
		} else {
			// TODO maybe use both old and new object to determine the change?
			// Detect change
			eventDoguVersion, getErr := getCurrentDoguVersionFromDoguDescriptorConfigMap(*descriptorConfigMap)
			if getErr != nil {
				throwAndLogWatchError(ctx, fmt.Errorf("failed to handle update watch event: %w", getErr), currentVersionsWatchResult)
				return
			}

			version, ok := persistenceContext[eventDoguVersion.Name]
			if ok && version.IsEqualTo(eventDoguVersion.Version) {
				logger.Info("current versions %s for dogu %s from persistent context and modified event are equal", eventDoguVersion.Version.Raw, eventDoguVersion.Name)
				return
			}

			persistenceContext[eventDoguVersion.Name] = eventDoguVersion.Version
			fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{eventDoguVersion})
		}
	}
}

func handleAdd(
	ctx context.Context,
	persistenceContext map[SimpleDoguName]core.Version,
	currentVersionsWatchResult chan CurrentVersionsWatchResult,
) func(obj interface{}) {
	return func(obj interface{}) {
		logger := log.FromContext(ctx).WithName("DoguVersionRegistry.handleAdd")
		descriptorConfigMap, err := toConfigMap(obj)
		if err != nil {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to handle add watch event: %w", err), currentVersionsWatchResult)
			return
		}

		// Skip process. Configmap was created empty.
		if !hasDoguDescriptorConfigMapCurrentKey(descriptorConfigMap) {
			logger.Info("dogu registry config map was created but without current key. do nothing.")
			return
		}

		eventDoguVersion, err := getCurrentDoguVersionFromDoguDescriptorConfigMap(*descriptorConfigMap)
		if err != nil {
			throwAndLogWatchError(ctx, fmt.Errorf("failed to handle add watch event: %w", err), currentVersionsWatchResult)
			return
		}

		oldPersistenceContext := copyPersistenceContext(persistenceContext)
		persistenceContext[eventDoguVersion.Name] = eventDoguVersion.Version
		fireWatchResult(currentVersionsWatchResult, oldPersistenceContext, persistenceContext, []DoguVersion{eventDoguVersion})
	}
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

func toConfigMap(obj interface{}) (*corev1.ConfigMap, error) {
	configMap, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("failed to cast event object to %T. wrong type %T", corev1.ConfigMap{}, obj)
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
