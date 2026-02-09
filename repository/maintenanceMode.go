package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudogu/ces-commons-lib/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	MaintenanceConfigMapName = "maintenance"
	maintenanceActiveKey     = "active"
	maintenanceTitleKey      = "title"
	maintenanceTextKey       = "text"
	maintenanceHolderKey     = "holder"
	maintenanceActiveTrue    = "true"
)

// MaintenanceModeDescription contains data that gets displayed when the maintenance mode is active.
type MaintenanceModeDescription struct {
	Title string
	Text  string
}

type MaintenanceModeAdapter struct {
	owner     string
	client    k8sClient
	namespace string
}

// NewMaintenanceModeAdapter creates a new adapter to handel the maintenance mode
func NewMaintenanceModeAdapter(owner string, client client.Client, namespace string) *MaintenanceModeAdapter {
	return &MaintenanceModeAdapter{
		owner:     owner,
		client:    client,
		namespace: namespace,
	}
}

// IsActive checks if the maintenance mode is active.
func (mma *MaintenanceModeAdapter) IsActive(ctx context.Context) (bool, error) {
	maintenanceConfig := &corev1.ConfigMap{}
	err := mma.client.Get(ctx, types.NamespacedName{Name: MaintenanceConfigMapName, Namespace: mma.namespace}, maintenanceConfig)
	if k8sErrs.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to get config for maintenance mode: %w", handleError(err))
	}

	return IsMaintenanceModeActive(maintenanceConfig), nil
}

func IsMaintenanceModeActive(config *corev1.ConfigMap) bool {
	activeString, ok := config.Data[maintenanceActiveKey]
	return ok && strings.ToLower(strings.TrimSpace(activeString)) == maintenanceActiveTrue
}

// Activate enables the maintenance mode and blocks the execution until the maintenance mode is activated.
// You can set timeouts via the go context.
// ConflictError if another component already activated the maintenance mode
// ConnectionError at any connection issues
// Generic Error at any other issue
func (mma *MaintenanceModeAdapter) Activate(ctx context.Context, content MaintenanceModeDescription) error {
	newConfig := newActiveMaintenanceConfig(mma.owner, content)
	return mma.setMaintenanceMode(ctx, newConfig)
}

// Deactivate disables the maintenance mode if it is active.
// ConflictError if another component activated the maintenance mode
// ConnectionError at any connection issues
// Generic Error at any other issue
func (mma *MaintenanceModeAdapter) Deactivate(ctx context.Context) error {
	newConfig := newInactiveMaintenanceConfig()
	return mma.setMaintenanceMode(ctx, newConfig)
}

func (mma *MaintenanceModeAdapter) setMaintenanceMode(ctx context.Context, config *maintenanceConfig) error {
	maintenanceConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MaintenanceConfigMapName,
			Namespace: mma.namespace,
		},
	}
	err := mma.client.Get(ctx, types.NamespacedName{Name: MaintenanceConfigMapName, Namespace: mma.namespace}, maintenanceConfigMap)
	if client.IgnoreNotFound(err) != nil {
		return fmt.Errorf("could not maintenance config-map: %w", handleError(err))
	}

	shouldCreate := k8sErrs.IsNotFound(err)

	if !shouldCreate {
		existingConfig := newMaintenanceConfigFromConfigMap(maintenanceConfigMap)
		err = mma.checkForConflict(existingConfig)
		if err != nil {
			return err
		}
	}

	config.setInConfigMap(maintenanceConfigMap)
	return mma.updateMaintenanceConfigMap(ctx, maintenanceConfigMap, shouldCreate)
}

func (mma *MaintenanceModeAdapter) updateMaintenanceConfigMap(ctx context.Context, configMap *corev1.ConfigMap, shouldCreate bool) error {
	if shouldCreate {
		err := mma.client.Create(ctx, configMap)
		if err != nil {
			return fmt.Errorf("could not create maintenance config-map: %w", handleError(err))
		}
	} else {
		err := mma.client.Update(ctx, configMap)
		if err != nil {
			return fmt.Errorf("could not update maintenance config-map: %w", handleError(err))
		}
	}

	return nil
}

func (mma *MaintenanceModeAdapter) checkForConflict(config *maintenanceConfig) error {
	if config.holder != mma.owner {
		return errors.NewConflictError(fmt.Errorf("maintenance mode is already activated by another owner: %s", config.holder))
	}
	return nil
}

type maintenanceConfig struct {
	active bool
	title  string
	text   string
	holder string
}

func newActiveMaintenanceConfig(owner string, description MaintenanceModeDescription) *maintenanceConfig {
	return &maintenanceConfig{
		active: true,
		title:  description.Title,
		text:   description.Text,
		holder: owner,
	}
}

func newInactiveMaintenanceConfig() *maintenanceConfig {
	return &maintenanceConfig{active: false}
}

func newMaintenanceConfigFromConfigMap(configMap *corev1.ConfigMap) *maintenanceConfig {
	return &maintenanceConfig{
		active: configMap.Data[maintenanceActiveKey] == maintenanceActiveTrue,
		title:  configMap.Data[maintenanceTitleKey],
		text:   configMap.Data[maintenanceTextKey],
		holder: configMap.Data[maintenanceHolderKey],
	}
}

func (m *maintenanceConfig) setInConfigMap(configMap *corev1.ConfigMap) {
	if configMap.Data == nil {
		configMap.Data = map[string]string{}
	}

	configMap.Data[maintenanceActiveKey] = strconv.FormatBool(m.active)

	if m.active {
		configMap.Data[maintenanceTitleKey] = m.title
		configMap.Data[maintenanceTextKey] = m.text
		configMap.Data[maintenanceHolderKey] = m.holder
		return
	}

	delete(configMap.Data, maintenanceTitleKey)
	delete(configMap.Data, maintenanceTextKey)
	delete(configMap.Data, maintenanceHolderKey)
}
