package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/cloudogu/k8s-registry-lib/errors"
)

const registryKeyMaintenance = "maintenance"

// MaintenanceModeDescription contains data that gets displayed when the maintenance mode is active.
type MaintenanceModeDescription struct {
	Title string
	Text  string
}

type MaintenanceModeAdapter struct {
	owner            string
	globalConfigRepo *GlobalConfigRepository
}

// NewMaintenanceModeAdapter creates a new adapter to handel the maintenance mode
func NewMaintenanceModeAdapter(owner string, client ConfigMapClient) *MaintenanceModeAdapter {
	return &MaintenanceModeAdapter{
		owner:            owner,
		globalConfigRepo: NewGlobalConfigRepository(client),
	}
}

type maintenanceConfigObject struct {
	Title  string `json:"title"`
	Text   string `json:"text"`
	Holder string `json:"holder,omitempty"`
}

func newMaintenanceConfigObject(owner string, description MaintenanceModeDescription) *maintenanceConfigObject {
	return &maintenanceConfigObject{
		Title:  description.Title,
		Text:   description.Text,
		Holder: owner,
	}
}

// Activate enables the maintenance mode and blocks the execution until the maintenance mode is activated.
// You can set timeouts via the go context.
// ConflictError if another component already activated the maintenance mode
// ConnectionError at any connection issues
// Generic Error at any other issue
func (mma *MaintenanceModeAdapter) Activate(ctx context.Context, content MaintenanceModeDescription) error {
	globalConfig, err := mma.globalConfigRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("could not get contents of global config-map for activating maintenance mode: %w", handleError(err))
	}

	if rawValue, isActive := globalConfig.Get(registryKeyMaintenance); !isActive {
		return mma.setMaintenanceModeInConfig(ctx, globalConfig, content)
	} else {
		return mma.checkForConflict(rawValue)
	}
}

func (mma *MaintenanceModeAdapter) setMaintenanceModeInConfig(ctx context.Context, globalConfig config.GlobalConfig, content MaintenanceModeDescription) error {
	configObject := newMaintenanceConfigObject(mma.owner, content)
	jsonBytes, err := json.Marshal(configObject)
	if err != nil {
		return errors.NewGenericError(fmt.Errorf("failed to serialize maintenance mode object: %w", err))
	}

	updatedConfig, err := globalConfig.Set(registryKeyMaintenance, config.Value(jsonBytes))
	if err != nil {
		return errors.NewGenericError(fmt.Errorf("failed to set maintenance mode registry key: %w", err))
	}

	_, err = mma.globalConfigRepo.Update(ctx, config.GlobalConfig{Config: updatedConfig})
	if err != nil {
		return fmt.Errorf("could not update global config-map for activating maintenance mode: %w", handleError(err))
	}
	return nil
}

func (mma *MaintenanceModeAdapter) checkForConflict(rawValue config.Value) error {
	var value maintenanceConfigObject
	err := json.Unmarshal([]byte(rawValue), &value)
	if err != nil {
		return errors.NewGenericError(fmt.Errorf("failed to parse json of maintenance mode object: %w", err))
	}
	if value.Holder != mma.owner {
		return errors.NewConflictError(fmt.Errorf("maintenance mode %s is already activated by another owner: %s", rawValue, value.Holder))
	}
	return nil
}

// Deactivate disables the maintenance mode if it is active.
// ConflictError if another component activated the maintenance mode
// ConnectionError at any connection issues
// Generic Error at any other issue
func (mma *MaintenanceModeAdapter) Deactivate(ctx context.Context) error {
	globalConfig, err := mma.globalConfigRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("could not get contents of global config-map for deactivating maintenance mode: %w", handleError(err))
	}

	if rawValue, isActive := globalConfig.Get(registryKeyMaintenance); isActive {
		err = mma.checkForConflict(rawValue)
		if err != nil {
			return err
		}

		updatedConfig := globalConfig.Delete(registryKeyMaintenance)
		_, err = mma.globalConfigRepo.Update(ctx, config.GlobalConfig{Config: updatedConfig})
		if err != nil {
			return fmt.Errorf("could not update global config-map for activating maintenance mode: %w", handleError(err))
		}
	}

	return nil
}
