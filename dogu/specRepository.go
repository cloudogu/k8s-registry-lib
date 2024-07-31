package dogu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
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
		return &core.Dogu{}, err
	}

	versionStr := doguVersion.Version.Raw
	doguStr, ok := specConfigMap.Data[versionStr]
	if !ok {
		return &core.Dogu{}, getDoguRegistryKeyNotFoundError(versionStr, doguName)
	}

	return unmarshalDoguStr(doguStr, doguName, doguVersion.Version)
}

func unmarshalDoguStr(doguStr string, doguName SimpleDoguName, doguVersion core.Version) (*core.Dogu, error) {
	dogu := &core.Dogu{}
	err := json.Unmarshal([]byte(doguStr), dogu)
	if err != nil {
		return &core.Dogu{}, fmt.Errorf("failed to unmarshal spec for dogu %q with version %q: %w", doguName, doguVersion.Raw, err)
	}

	return dogu, nil
}

func (vr *specRepository) GetAll(ctx context.Context, doguVersions []DoguVersion) (map[DoguVersion]*core.Dogu, error) {
	// TODO get only versions defined in doguVersions parameter
	registryList, err := getAllSpecConfigMaps(ctx, vr.configMapClient)
	if err != nil {
		return nil, err
	}

	var errs []error
	allDogus := map[DoguVersion]*core.Dogu{}
	for _, localRegistry := range registryList.Items {
		for versionStr, doguStr := range localRegistry.Data {
			if versionStr == currentVersionKey {
				continue
			}

			doguName := localRegistry.Labels[doguNameLabelKey]
			simpleDoguName := SimpleDoguName(doguName)
			parsedVersion, parseErr := parseDoguVersion(versionStr, simpleDoguName)
			if parseErr != nil {
				errs = append(errs, parseErr)
				continue
			}

			dogu, unmarshalErr := unmarshalDoguStr(doguStr, simpleDoguName, parsedVersion)
			if unmarshalErr != nil {
				errs = append(errs, unmarshalErr)
				continue
			}

			doguVersion := DoguVersion{
				Name:    simpleDoguName,
				Version: parsedVersion,
			}
			allDogus[doguVersion] = dogu
		}
	}

	err = errors.Join(errs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get some dogu specs: %w", err)
	}

	return allDogus, nil
}

func (vr *specRepository) Add(ctx context.Context, SimpleDoguName, dogu *core.Dogu) error {
	return nil
}

func (vr *specRepository) DeleteAll(ctx context.Context, name SimpleDoguName) error {
	return nil
}
