package repository

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/cloudogu/k8s-registry-lib/errors"
)

type configType int

const (
	globalConfigType configType = iota + 1
	doguConfigType
	sensitiveConfigType
)

func (t configType) String() string {
	switch t {
	case globalConfigType:
		return "global-config"
	case doguConfigType:
		return "dogu-config"
	case sensitiveConfigType:
		return "sensitive-config"
	default:
		return "unknown"
	}
}

const (
	appLabelKey      = "app"
	appLabelValueCes = "ces"
	typeLabelKey     = "k8s.cloudogu.com/type"
	doguNameLabelKey = "dogu.name"
)

const dataKeyName = "config.yaml"

type ConfigMapClient interface {
	corev1client.ConfigMapInterface
}

type clientData struct {
	dataStr string
	rawData any
}

type configMapClient struct {
	client ConfigMapClient
	labels labels.Set
}

var _ configClient = configMapClient{}

func createConfigMapClient(c ConfigMapClient, t configType) configMapClient {
	return configMapClient{
		client: c,
		labels: labels.Set{
			appLabelKey:  appLabelValueCes,
			typeLabelKey: t.String(),
		},
	}
}

func handleError(err error) error {
	if k8sErrs.IsNotFound(err) {
		return errors.NewNotFoundError(err)
	}

	if k8sErrs.IsConflict(err) {
		return errors.NewConflictError(err)
	}

	if k8sErrs.IsServerTimeout(err) || k8sErrs.IsTimeout(err) {
		return errors.NewConnectionError(err)
	}

	if k8sErrs.IsAlreadyExists(err) {
		return errors.NewAlreadyExistsError(err)
	}

	return errors.NewGenericError(err)
}

func (cmc configMapClient) Get(ctx context.Context, name string) (clientData, error) {
	cm, err := cmc.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return clientData{}, fmt.Errorf("unable to get config-map from cluster: %w", handleError(err))
	}

	dataStr, ok := cm.Data[dataKeyName]
	if !ok {
		return clientData{}, fmt.Errorf("could not find data for key %s", dataKeyName)
	}

	return clientData{
		dataStr: dataStr,
		rawData: cm,
	}, nil
}

// SingletonList gets a list of configmaps containing a single item. This is used for the config-watches, because they
// are operating on lists instead of single objects.
func (cmc configMapClient) SingletonList(ctx context.Context, name string) (clientData, string, error) {
	list, err := cmc.client.List(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	logger := log.FromContext(ctx).WithName("SingletonList")
	logger.Info(fmt.Sprintf("ResultList is %v and error %e", list, err))
	if err != nil {
		return clientData{}, "", fmt.Errorf("unable to list config-map from cluster: %w", handleError(err))
	}

	if len(list.Items) == 0 {
		return clientData{}, "", fmt.Errorf("could not find a configmap with the given name %s", name)
	}

	configMap := list.Items[0]
	dataStr, ok := configMap.Data[dataKeyName]
	if !ok {
		return clientData{}, "", fmt.Errorf("could not find data for key %s", dataKeyName)
	}

	return clientData{
		dataStr: dataStr,
		rawData: configMap,
	}, list.ResourceVersion, nil
}

func (cmc configMapClient) Delete(ctx context.Context, name string) error {
	if err := cmc.client.Delete(ctx, name, metav1.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("could not delete config-map in cluster: %w", handleError(err))
	}

	return nil
}

func (cmc configMapClient) createConfigMap(pCtx string, name string, doguName string, dataStr string) *v1.ConfigMap {
	if doguName != "" {
		cmc.labels[doguNameLabelKey] = doguName
	}

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          cmc.labels,
			ResourceVersion: pCtx,
		},
		Data: map[string]string{
			dataKeyName: dataStr,
		},
	}

	return configMap
}

func (cmc configMapClient) Create(ctx context.Context, name string, doguName string, dataStr string) (resourceVersionGetter, error) {
	configMap := cmc.createConfigMap("", name, doguName, dataStr)

	cm, err := cmc.client.Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not create configmap in cluster: %w", handleError(err))
	}

	return cm, nil
}

func (cmc configMapClient) Update(ctx context.Context, pCtx string, name string, doguName string, dataStr string) (resourceVersionGetter, error) {
	configMap := cmc.createConfigMap(pCtx, name, doguName, dataStr)

	updatedConfigMap, err := cmc.client.Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not update configmap in cluster: %w", handleError(err))
	}

	return updatedConfigMap, nil
}

func (cmc configMapClient) UpdateClientData(ctx context.Context, update clientData) (resourceVersionGetter, error) {
	cm, ok := update.rawData.(*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("configData cannot be used as configMap")
	}

	cm.Data[dataKeyName] = update.dataStr

	cm, err := cmc.client.Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not update configmap in cluster: %w", handleError(err))
	}

	return cm, nil
}

func (cmc configMapClient) Watch(ctx context.Context, name string, resourceVersion string) (<-chan clientWatchResult, error) {
	list, err := cmc.client.List(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	if err != nil {
		return nil, fmt.Errorf("failed to list single configmap %q for watch: %w", name, handleError(err))
	}
	logger := log.FromContext(ctx).WithName("Watch")
	logger.Info(fmt.Sprintf("Old ResourceVersion: %s New ResourceVersion: %s", list.ResourceVersion, resourceVersion))
	return watchWithClient(ctx, cmc.client, name, list.ResourceVersion)
}

type SecretClient interface {
	corev1client.SecretInterface
}

type secretClient struct {
	client SecretClient
	labels labels.Set
}

var _ configClient = secretClient{}

func createSecretClient(c SecretClient, t configType) secretClient {
	return secretClient{
		client: c,
		labels: labels.Set{
			appLabelKey:  appLabelValueCes,
			typeLabelKey: t.String(),
		},
	}
}

func (sc secretClient) Get(ctx context.Context, name string) (clientData, error) {
	secret, err := sc.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return clientData{}, fmt.Errorf("unable to get secret from cluster: %w", handleError(err))
	}

	dataBytes, ok := secret.Data[dataKeyName]
	if !ok {
		return clientData{}, fmt.Errorf("could not find data for key %s", dataKeyName)
	}

	return clientData{
		dataStr: string(dataBytes),
		rawData: secret,
	}, nil
}

// SingletonList gets a list of secrets containing a single item. This is used for the config-watches, because they
// are operating on lists instead of single objects.
func (sc secretClient) SingletonList(ctx context.Context, name string) (clientData, string, error) {
	list, err := sc.client.List(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	logger := log.FromContext(ctx).WithName("SingletonList")
	logger.Info(fmt.Sprintf("ResultList is %v and error %e", list, err))
	if err != nil {
		return clientData{}, "", fmt.Errorf("unable to list config-map from cluster: %w", handleError(err))
	}

	if len(list.Items) == 0 {
		return clientData{}, "", fmt.Errorf("could not find a configmap with the given name %s", name)
	}

	secret := list.Items[0]
	dataBytes, ok := secret.Data[dataKeyName]
	if !ok {
		return clientData{}, "", fmt.Errorf("could not find data for key %s", dataKeyName)
	}

	return clientData{
		dataStr: string(dataBytes),
		rawData: secret,
	}, list.ResourceVersion, nil
}

func (sc secretClient) Delete(ctx context.Context, name string) error {
	if err := sc.client.Delete(ctx, name, metav1.DeleteOptions{}); client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("could not delete secret in cluster: %w", handleError(err))
	}

	return nil
}

func (sc secretClient) createSecret(pCtx string, name string, doguName string, dataStr string) *v1.Secret {
	if doguName != "" {
		sc.labels[doguNameLabelKey] = doguName
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          sc.labels,
			ResourceVersion: pCtx,
		},
		StringData: map[string]string{
			dataKeyName: dataStr,
		},
	}

	return secret
}

func (sc secretClient) Create(ctx context.Context, name string, doguName string, dataStr string) (resourceVersionGetter, error) {
	secret := sc.createSecret("", name, doguName, dataStr)

	cm, err := sc.client.Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not create secret in cluster: %w", handleError(err))
	}

	return cm, nil
}

func (sc secretClient) Update(ctx context.Context, pCtx string, name string, doguName string, dataStr string) (resourceVersionGetter, error) {
	secret := sc.createSecret(pCtx, name, doguName, dataStr)

	updatedSecret, err := sc.client.Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not update secret in cluster: %w", handleError(err))
	}

	return updatedSecret, nil
}

func (sc secretClient) UpdateClientData(ctx context.Context, update clientData) (resourceVersionGetter, error) {
	secret, ok := update.rawData.(*v1.Secret)
	if !ok {
		return nil, fmt.Errorf("configData cannot be used as secret")
	}

	secret.StringData = map[string]string{
		dataKeyName: update.dataStr,
	}

	resource, err := sc.client.Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not update secret in cluster: %w", handleError(err))
	}

	return resource, nil
}

func (sc secretClient) Watch(ctx context.Context, name string, resourceVersion string) (<-chan clientWatchResult, error) {
	list, err := sc.client.List(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	if err != nil {
		return nil, fmt.Errorf("failed to list single configmap %q for watch: %w", name, handleError(err))
	}
	logger := log.FromContext(ctx).WithName("Watch")
	logger.Info(fmt.Sprintf("Old ResourceVersion: %s New ResourceVersion: %s", list.ResourceVersion, resourceVersion))
	return watchWithClient(ctx, sc.client, name, list.ResourceVersion)
}

type clientWatcher interface {
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type clientWatchResult struct {
	dataStr           string
	persistentContext string
	err               error
}

func watchWithClient(ctx context.Context, client clientWatcher, name, initialResourceVersion string) (<-chan clientWatchResult, error) {
	logger := log.FromContext(ctx).WithName("watchWithClient")

	watcher, err := createRetryWatcher(ctx, client, name, initialResourceVersion)
	if err != nil {
		return nil, err
	}

	resultChan := make(chan clientWatchResult)

	go func() {
		defer close(resultChan)
		for {
			select {
			case <-ctx.Done():
				watcher.Stop()
				return
			case event, open := <-watcher.ResultChan():
				if !open {
					logger.Info(fmt.Sprintf("watch for %q closed", name))

					return
				}

				var result clientWatchResult
				result = handleWatchEvent(name, event)
				resultChan <- result
			}
		}
	}()

	return resultChan, nil
}

func createRetryWatcher(ctx context.Context, client clientWatcher, name, initialResourceVersion string) (*toolsWatch.RetryWatcher, error) {
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		options.FieldSelector = fields.OneTermEqualSelector("metadata.name", name).String()
		watchInterface, err := client.Watch(ctx, options)
		if err != nil {
			return nil, err
		}

		return watchInterface, nil
	}
	watcher, err := toolsWatch.NewRetryWatcher(initialResourceVersion, &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		return nil, fmt.Errorf("could not watch '%s' in cluster: %w", name, handleError(err))
	}

	return watcher, nil
}

func handleWatchEvent(cfgName string, event watch.Event) clientWatchResult {
	if event.Type == watch.Error {
		var err error
		status, ok := event.Object.(*metav1.Status)
		if !ok {
			err = fmt.Errorf("error result in watcher for config '%s'", cfgName)
		} else {
			err = fmt.Errorf("watch event type is error: %q", status.String())
		}

		return clientWatchResult{
			dataStr:           "",
			persistentContext: "",
			err:               err,
		}
	}

	switch r := event.Object.(type) {
	case *v1.Secret:
		dataBytes, ok := r.Data[dataKeyName]
		if !ok {
			return clientWatchResult{
				dataStr:           "",
				persistentContext: "",
				err:               fmt.Errorf("could not find data for key %s in secret %s", dataKeyName, cfgName),
			}
		}

		return clientWatchResult{
			dataStr:           string(dataBytes),
			persistentContext: r.GetResourceVersion(),
			err:               nil,
		}
	case *v1.ConfigMap:
		dataString, ok := r.Data[dataKeyName]
		if !ok {
			return clientWatchResult{
				dataStr:           "",
				persistentContext: "",
				err:               fmt.Errorf("could not find data for key %s in configmap %s", dataKeyName, cfgName),
			}
		}

		return clientWatchResult{
			dataStr:           dataString,
			persistentContext: r.GetResourceVersion(),
			err:               nil,
		}
	default:
		return clientWatchResult{
			dataStr:           "",
			persistentContext: "",
			err:               fmt.Errorf("unsupported type in watch %T", r),
		}
	}
}
