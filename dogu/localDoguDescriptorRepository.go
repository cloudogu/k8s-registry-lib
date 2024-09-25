package dogu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	cloudoguerrors "github.com/cloudogu/k8s-registry-lib/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

type localDoguDescriptorRepository struct {
	configMapClient configMapClient
}

func NewLocalDoguDescriptorRepository(configMapClient configMapClient) *localDoguDescriptorRepository {
	return &localDoguDescriptorRepository{
		configMapClient: configMapClient,
	}
}

func (lddr *localDoguDescriptorRepository) Get(ctx context.Context, doguVersion DoguVersion) (*core.Dogu, error) {
	doguName := doguVersion.Name
	descriptorConfigMap, err := getDescriptorConfigMapForDogu(ctx, lddr.configMapClient, doguName)
	if err != nil {
		return nil, handleK8sError(err)
	}

	versionStr := doguVersion.Version.Raw
	doguStr, ok := descriptorConfigMap.Data[versionStr]
	if !ok {
		return nil, getDoguRegistryKeyNotFoundError(versionStr, doguName)
	}

	result, err := unmarshalDoguJsonStr(doguStr, doguName, versionStr)
	if err != nil {
		return nil, cloudoguerrors.NewGenericError(err)
	}
	return result, err
}

func unmarshalDoguJsonStr(doguStr string, doguName SimpleDoguName, doguVersion string) (*core.Dogu, error) {
	dogu := &core.Dogu{}
	err := json.Unmarshal([]byte(doguStr), dogu)
	if err != nil {
		return &core.Dogu{}, fmt.Errorf("failed to unmarshal descriptor for dogu %q with version %q: %w", doguName, doguVersion, err)
	}

	return dogu, nil
}

func (lddr *localDoguDescriptorRepository) GetAll(ctx context.Context, doguVersions []DoguVersion) (map[DoguVersion]*core.Dogu, error) {
	allDogus := make(map[DoguVersion]*core.Dogu, len(doguVersions))
	versionsByDogu := map[SimpleDoguName][]DoguVersion{}
	for _, doguVersion := range doguVersions {
		if versionsByDogu[doguVersion.Name] == nil {
			versionsByDogu[doguVersion.Name] = []DoguVersion{}
		}
		versionsByDogu[doguVersion.Name] = append(versionsByDogu[doguVersion.Name], doguVersion)
	}

	var multiErr []error
	for doguName, versions := range versionsByDogu {
		doguDescriptorConfigMap, err := getDescriptorConfigMapForDogu(ctx, lddr.configMapClient, doguName)
		if err != nil {
			multiErr = append(multiErr, err)
			continue
		}
		for _, doguVersion := range versions {
			doguStr, ok := doguDescriptorConfigMap.Data[doguVersion.Version.Raw]
			if !ok {
				multiErr = append(multiErr, fmt.Errorf("did not find expected version %q for dogu %q in dogu descriptor configmap", doguVersion.Version.Raw, doguName))
				continue
			}

			dogu, unmarshalErr := unmarshalDoguJsonStr(doguStr, doguName, doguVersion.Version.Raw)
			if unmarshalErr != nil {
				multiErr = append(multiErr, unmarshalErr)
				continue
			}

			allDogus[doguVersion] = dogu
		}
	}

	err := errors.Join(multiErr...)
	if err != nil {
		return nil, cloudoguerrors.NewGenericError(fmt.Errorf("failed to get some dogu descriptors: %w", err))
	}

	return allDogus, nil
}

func (lddr *localDoguDescriptorRepository) Add(ctx context.Context, name SimpleDoguName, dogu *core.Dogu) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		doguDescriptorConfigMap, err := getOrCreateDescriptorConfigMapForDogu(ctx, lddr.configMapClient, name)
		if err != nil {
			return err
		}

		if doguDescriptorConfigMap.Data == nil {
			doguDescriptorConfigMap.Data = map[string]string{}
		}

		_, alreadyExists := doguDescriptorConfigMap.Data[dogu.Version]
		if alreadyExists {
			return cloudoguerrors.NewAlreadyExistsError(fmt.Errorf("%q dogu descriptor already exists for version %q", name, dogu.Version))
		}

		doguBytes, err := json.Marshal(dogu)
		if err != nil {
			return cloudoguerrors.NewGenericError(fmt.Errorf("failed to marshal dogu %v: %w", dogu, err))
		}

		doguDescriptorConfigMap.Data[dogu.Version] = string(doguBytes)

		_, err = lddr.configMapClient.Update(ctx, doguDescriptorConfigMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update dogu descriptor configmap for dogu %q: %w", name, err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func getOrCreateDescriptorConfigMapForDogu(ctx context.Context, configMapClient configMapClient, simpleDoguName SimpleDoguName) (*corev1.ConfigMap, error) {
	descriptorConfigMap, err := getDescriptorConfigMapForDogu(ctx, configMapClient, simpleDoguName)
	if err != nil {
		if cloudoguerrors.IsNotFoundError(err) {
			var createErr error
			descriptorConfigMap, createErr = createDescriptorConfigMapForDogu(ctx, configMapClient, simpleDoguName)
			if createErr != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return descriptorConfigMap, nil
}

func createDescriptorConfigMapForDogu(ctx context.Context, configMapClient configMapClient, simpleDoguName SimpleDoguName) (*corev1.ConfigMap, error) {
	descriptorConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: getDescriptorConfigMapName(simpleDoguName),
			Labels: map[string]string{
				appLabelKey:      appLabelValueCes,
				doguNameLabelKey: string(simpleDoguName),
				typeLabelKey:     typeLabelValueLocalDoguRegistry,
			},
		},
	}

	_, createErr := configMapClient.Create(ctx, descriptorConfigMap, metav1.CreateOptions{})
	if createErr != nil {
		return nil, fmt.Errorf("failed to create local registry config map for dogu %q: %w", simpleDoguName, createErr)
	}

	return descriptorConfigMap, nil
}

func getDescriptorConfigMapName(simpleDoguName SimpleDoguName) string {
	return fmt.Sprintf("dogu-spec-%s", simpleDoguName)
}

func (lddr *localDoguDescriptorRepository) DeleteAll(ctx context.Context, name SimpleDoguName) error {
	err := lddr.configMapClient.Delete(ctx, getDescriptorConfigMapName(name), metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete dogu descriptor configmap for dogu %q: %w", name, handleK8sError(err))
	}

	return nil
}
