package dogu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	cloudoguerrors "github.com/cloudogu/k8s-registry-lib/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	specConfigMap, err := getSpecConfigMapForDogu(ctx, lddr.configMapClient, doguName)
	if err != nil {
		return nil, cloudoguerrors.NewGenericError(err)
	}

	versionStr := doguVersion.Version.Raw
	doguStr, ok := specConfigMap.Data[versionStr]
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
		return &core.Dogu{}, fmt.Errorf("failed to unmarshal spec for dogu %q with version %q: %w", doguName, doguVersion, err)
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
		doguSpecConfigMap, err := getSpecConfigMapForDogu(ctx, lddr.configMapClient, doguName)
		if err != nil {
			multiErr = append(multiErr, err)
			continue
		}
		for _, doguVersion := range versions {
			doguStr, ok := doguSpecConfigMap.Data[doguVersion.Version.Raw]
			if !ok {
				multiErr = append(multiErr, fmt.Errorf("did not find expected version %q for dogu %q in dogu spec configmap", doguVersion.Version.Raw, doguName))
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
		return nil, cloudoguerrors.NewGenericError(fmt.Errorf("failed to get some dogu specs: %w", err))
	}

	return allDogus, nil
}

func (lddr *localDoguDescriptorRepository) Add(ctx context.Context, name SimpleDoguName, dogu *core.Dogu) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		doguSpecConfigMap, err := getOrCreateSpecConfigMapForDogu(ctx, lddr.configMapClient, name)
		if err != nil {
			return cloudoguerrors.NewGenericError(err)
		}

		if doguSpecConfigMap.Data == nil {
			doguSpecConfigMap.Data = map[string]string{}
		}

		_, alreadyExists := doguSpecConfigMap.Data[dogu.Version]
		if alreadyExists {
			return cloudoguerrors.NewAlreadyExistsError(fmt.Errorf("%q dogu spec already exists for version %q", name, dogu.Version))
		}

		doguBytes, err := json.Marshal(dogu)
		if err != nil {
			return cloudoguerrors.NewGenericError(fmt.Errorf("failed to marshal dogu %v: %w", dogu, err))
		}

		doguSpecConfigMap.Data[dogu.Version] = string(doguBytes)

		_, err = lddr.configMapClient.Update(ctx, doguSpecConfigMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update dogu spec configmap for dogu %q: %w", name, err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
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
			return nil, err
		}
	}

	return specConfigMap, nil
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
		return nil, fmt.Errorf("failed to create local registry config map for dogu %q: %w", simpleDoguName, createErr)
	}

	return specConfigMap, nil
}

func getSpecConfigMapName(simpleDoguName SimpleDoguName) string {
	return fmt.Sprintf("dogu-spec-%s", simpleDoguName)
}

func (lddr *localDoguDescriptorRepository) DeleteAll(ctx context.Context, name SimpleDoguName) error {
	err := lddr.configMapClient.Delete(ctx, getSpecConfigMapName(name), metav1.DeleteOptions{})
	if err != nil {
		return cloudoguerrors.NewGenericError(fmt.Errorf("failed to delete dogu spec configmap for dogu %q: %w", name, err))
	}

	return nil
}
