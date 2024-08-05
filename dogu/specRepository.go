package dogu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	cloudoguerrors "github.com/cloudogu/k8s-registry-lib/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type specRepository struct {
	configMapClient configMapClient
}

func NewSpecRepository(configMapClient configMapClient) *specRepository {
	return &specRepository{
		configMapClient: configMapClient,
	}
}

func (vr *specRepository) Get(ctx context.Context, doguVersion DoguVersion) (*core.Dogu, error) {
	doguName := doguVersion.Name
	specConfigMap, err := getSpecConfigMapForDogu(ctx, vr.configMapClient, doguName)
	if err != nil {
		return &core.Dogu{}, cloudoguerrors.NewGenericError(err)
	}

	versionStr := doguVersion.Version.Raw
	doguStr, ok := specConfigMap.Data[versionStr]
	if !ok {
		return &core.Dogu{}, getDoguRegistryKeyNotFoundError(versionStr, doguName)
	}

	result, err := unmarshalDoguJsonStr(doguStr, doguName, doguVersion.Version.Raw)
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

func (vr *specRepository) GetAll(ctx context.Context, doguVersions []DoguVersion) (map[DoguVersion]*core.Dogu, error) {
	var allDogus map[DoguVersion]*core.Dogu
	var versionsByDogu map[SimpleDoguName][]DoguVersion
	for _, doguVersion := range doguVersions {
		versionsByDogu[doguVersion.Name] = append(versionsByDogu[doguVersion.Name], doguVersion)
	}

	var multiErr []error
	for doguName, versions := range versionsByDogu {
		doguSpecConfigMap, err := getSpecConfigMapForDogu(ctx, vr.configMapClient, doguName)
		if err != nil {
			multiErr = append(multiErr, err)
			continue
		}
		for _, doguVersion := range versions {
			doguStr, ok := doguSpecConfigMap.Data[doguVersion.Version.Raw]
			if !ok {
				multiErr = append(multiErr, fmt.Errorf("did not find expected version %s for dogu %s in dogu spec configmap", doguVersion.Version.Raw, doguName))
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

func (vr *specRepository) Add(ctx context.Context, name SimpleDoguName, dogu *core.Dogu) error {
	doguSpecConfigMap, err := getOrCreateSpecConfigMapForDogu(ctx, vr.configMapClient, name)
	if err != nil {
		return cloudoguerrors.NewGenericError(err)
	}

	_, ok := doguSpecConfigMap.Data[dogu.Version]
	if ok {
		return cloudoguerrors.NewAlreadyExistsError(fmt.Errorf("dogu spec already exists for version %s", dogu.Version))
	}

	doguBytes, err := json.Marshal(dogu)
	if err != nil {
		return cloudoguerrors.NewGenericError(fmt.Errorf("failed to marshal dogu %v: %w", dogu, err))
	}

	doguSpecConfigMap.Data[dogu.Version] = string(doguBytes)

	// TODO retry
	_, err = vr.configMapClient.Update(ctx, doguSpecConfigMap, v1.UpdateOptions{})
	if err != nil {
		return cloudoguerrors.NewGenericError(fmt.Errorf("failed to update dogu spec configmap %s: %w", doguSpecConfigMap.Name, err))
	}

	return nil
}

func (vr *specRepository) DeleteAll(ctx context.Context, name SimpleDoguName) error {
	err := vr.configMapClient.Delete(ctx, string(name), v1.DeleteOptions{})
	if err != nil {
		return cloudoguerrors.NewGenericError(fmt.Errorf("failed to delete dogu spec configmap for dogu %s: %w", name, err))
	}

	return nil
}
