package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	informerCore "k8s.io/client-go/informers/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cloudogu/k8s-registry-lib/config"
	regErrs "github.com/cloudogu/k8s-registry-lib/errors"
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

type ConfigMapInformer interface {
	informerCore.ConfigMapInformer
}

type sharedInformer interface {
	cache.SharedIndexInformer
}

type watchKind string

const (
	configMapWatchKind watchKind = "configmap"
	secretWatchKind    watchKind = "secret"
)

type clientData struct {
	dataStr string
	rawData any
}

type configMapClient struct {
	client   ConfigMapClient
	informer ConfigMapInformer
	labels   labels.Set
}

var _ configClient = configMapClient{}

func createConfigMapClient(c ConfigMapClient, i ConfigMapInformer, t configType) configMapClient {
	return configMapClient{
		client:   c,
		informer: i,
		labels: labels.Set{
			appLabelKey:  appLabelValueCes,
			typeLabelKey: t.String(),
		},
	}
}

func handleError(err error) error {
	if k8sErrs.IsNotFound(err) {
		return regErrs.NewNotFoundError(err)
	}

	if k8sErrs.IsConflict(err) {
		return regErrs.NewConflictError(err)
	}

	if k8sErrs.IsServerTimeout(err) || k8sErrs.IsTimeout(err) {
		return regErrs.NewConnectionError(err)
	}

	if k8sErrs.IsAlreadyExists(err) {
		return regErrs.NewAlreadyExistsError(err)
	}

	return fmt.Errorf("%v", err)
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

func (cmc configMapClient) Watch(ctx context.Context, name string) (<-chan clientWatchResult, error) {
	return registerEventHandler(ctx, cmc.informer.Informer(), configMapWatchKind, name)
}

type SecretClient interface {
	corev1client.SecretInterface
}

type SecretInformer interface {
	informerCore.SecretInformer
}

type secretClient struct {
	client   SecretClient
	informer SecretInformer
	labels   labels.Set
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

func (sc secretClient) Watch(ctx context.Context, name string) (<-chan clientWatchResult, error) {
	return registerEventHandler(ctx, sc.informer.Informer(), secretWatchKind, name)
}

func registerEventHandler(ctx context.Context, informer sharedInformer, kind watchKind, name string) (<-chan clientWatchResult, error) {
	watchCh := make(chan clientWatchResult)
	_, err := informer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			var isType bool
			var cast metav1.Object
			switch kind {
			case secretWatchKind:
				cast, isType = obj.(*v1.Secret)
			case configMapWatchKind:
				cast, isType = obj.(*v1.ConfigMap)
			default:
				return false
			}

			return isType && cast.GetName() == name
		},
		Handler: cache.ResourceEventHandlerFuncs{
			UpdateFunc: updateHandler(kind, watchCh, name),
			DeleteFunc: deleteHandler(kind, watchCh, name),
		}})
	if err != nil {
		return nil, fmt.Errorf("failed to register event handler for kind %T: %w", kind, err)
	}

	go func() {
		informer.Run(ctx.Done())
	}()

	return watchCh, nil
}

func updateHandler(kind watchKind, watchCh chan clientWatchResult, name string) func(prevObj interface{}, newObj interface{}) {
	return func(prevObj, newObj interface{}) {
		switch kind {
		case secretWatchKind:
			watchCh <- getSecretWatchResult(prevObj, newObj, name)
		case configMapWatchKind:
			watchCh <- getConfigMapDataStrings(prevObj, newObj, name)
		default:
			watchCh <- clientWatchResult{err: regErrs.NewGenericError(fmt.Errorf("unknown watch kind: %s", kind))}
		}
	}
}

func deleteHandler(kind watchKind, watchCh chan clientWatchResult, name string) func(obj interface{}) {
	return func(obj interface{}) {
		watchCh <- clientWatchResult{
			err: regErrs.NewNotFoundError(fmt.Errorf("subject of watch (%s %s) was deleted", kind, name)),
		}
	}
}

func getSecretWatchResult(prevObj, newObj interface{}, name string) clientWatchResult {
	var errs []error
	prevSecret, ok := prevObj.(*v1.Secret)
	if !ok {
		errs = append(errs, fmt.Errorf("previous object is not of type secret"))
	}

	newSecret, ok := newObj.(*v1.Secret)
	if !ok {
		errs = append(errs, fmt.Errorf("updated object is not of type secret"))
	}

	prevDataBytes, ok := prevSecret.Data[dataKeyName]
	if !ok {
		errs = append(errs, fmt.Errorf("could not find data for key %q in previous state of secret %q", dataKeyName, name))
	}

	newDataBytes, ok := newSecret.Data[dataKeyName]
	if !ok {
		errs = append(errs, fmt.Errorf("could not find data for key %q in updated state of secret %q", dataKeyName, name))
	}

	return clientWatchResult{
		prevConfig: configRaw{
			data:           string(prevDataBytes),
			persistenceCtx: prevSecret.ResourceVersion,
		},
		newConfig: configRaw{
			data:           string(newDataBytes),
			persistenceCtx: newSecret.ResourceVersion,
		},
		err: regErrs.NewGenericError(errors.Join(errs...)),
	}
}

func getConfigMapDataStrings(prevObj, newObj interface{}, name string) clientWatchResult {
	var errs []error
	prevConfigMap, ok := prevObj.(*v1.ConfigMap)
	if !ok {
		errs = append(errs, fmt.Errorf("previous object is not of type configmap"))
	}

	newConfigMap, ok := newObj.(*v1.ConfigMap)
	if !ok {
		errs = append(errs, fmt.Errorf("updated object is not of type configmap"))
	}

	prevDataString, ok := prevConfigMap.Data[dataKeyName]
	if !ok {
		errs = append(errs, fmt.Errorf("could not find data for key %q in previous state of configmap %q", dataKeyName, name))
	}

	newDataString, ok := newConfigMap.Data[dataKeyName]
	if !ok {
		errs = append(errs, fmt.Errorf("could not find data for key %q in updated state of configmap %q", dataKeyName, name))
	}

	return clientWatchResult{
		prevConfig: configRaw{
			data:           prevDataString,
			persistenceCtx: prevConfigMap.ResourceVersion,
		},
		newConfig: configRaw{
			data:           newDataString,
			persistenceCtx: newConfigMap.ResourceVersion,
		},
		err: regErrs.NewGenericError(errors.Join(errs...)),
	}
}

type configRaw struct {
	data           string
	persistenceCtx string
}

func (c configRaw) toConfig(converter config.Converter) (config.Config, error) {
	reader := strings.NewReader(c.data)

	cfgData, err := converter.Read(reader)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not convert client data to config data: %w", err)
	}

	return config.CreateConfig(cfgData, config.WithPersistenceContext(c.persistenceCtx)), nil
}

type clientWatchResult struct {
	prevConfig configRaw
	newConfig  configRaw
	err        error
}
