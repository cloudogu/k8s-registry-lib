package global

import (
	"fmt"
	"github.com/cloudogu/cesapp-lib/keys"
	"github.com/cloudogu/cesapp-lib/registry"
)

type encryptedEtcdRegistry struct {
	etcdRegistry registry.ConfigurationContext
	publicKey    *keys.PublicKey
	privateKey   *keys.PrivateKey
}

func newEncryptedEtcdRegistry(reg registry.ConfigurationContext, publicKey *keys.PublicKey, privateKey *keys.PrivateKey) *encryptedEtcdRegistry {
	return &encryptedEtcdRegistry{
		etcdRegistry: reg,
		publicKey:    publicKey,
		privateKey:   privateKey,
	}
}

func (e encryptedEtcdRegistry) encrypt(value string) (string, error) {
	return value, nil
}

func (e encryptedEtcdRegistry) decrypt(value string) (string, error) {
	return value, nil
}

func (e encryptedEtcdRegistry) Set(key, value string) error {
	encryptedValue, err := e.encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt key %s: %w", key, err)
	}

	return e.etcdRegistry.Set(key, encryptedValue)
}

func (e encryptedEtcdRegistry) SetWithLifetime(key, value string, timeToLiveInSeconds int) error {
	encryptedValue, err := e.encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt key %s: %w", key, err)
	}

	return e.etcdRegistry.SetWithLifetime(key, encryptedValue, timeToLiveInSeconds)
}

func (e encryptedEtcdRegistry) Refresh(key string, timeToLiveInSeconds int) error {
	return e.etcdRegistry.Refresh(key, timeToLiveInSeconds)
}

func (e encryptedEtcdRegistry) Get(key string) (string, error) {
	value, err := e.etcdRegistry.Get(key)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt key %s: %w", key, err)
	}

	return e.decrypt(value)
}

func (e encryptedEtcdRegistry) GetAll() (map[string]string, error) {
	all, err := e.etcdRegistry.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys: %w", err)
	}

	for key, value := range all {
		decryptedValue, err := e.decrypt(value)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt key %s: %w", key, err)
		}
		all[key] = decryptedValue
	}

	return all, nil
}

func (e encryptedEtcdRegistry) Delete(key string) error {
	return e.etcdRegistry.Delete(key)
}

func (e encryptedEtcdRegistry) DeleteRecursive(key string) error {
	return e.etcdRegistry.DeleteRecursive(key)
}

func (e encryptedEtcdRegistry) Exists(key string) (bool, error) {
	return e.etcdRegistry.Exists(key)
}

func (e encryptedEtcdRegistry) RemoveAll() error {
	return e.etcdRegistry.RemoveAll()
}

func (e encryptedEtcdRegistry) GetOrFalse(key string) (bool, string, error) {
	exists, value, err := e.etcdRegistry.GetOrFalse(key)
	if err != nil {
		return false, "", fmt.Errorf("failed to decrypt key %s: %w", key, err)
	}

	decryptedValue, err := e.decrypt(value)

	return exists, decryptedValue, err
}
