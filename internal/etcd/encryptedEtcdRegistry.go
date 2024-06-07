package etcd

import (
	"fmt"
	"github.com/cloudogu/cesapp-lib/keys"
)

type EncryptedRegistry struct {
	etcdRegistry ConfigurationContext
	encrypt      func(value string) (string, error)
	decrypt      func(value string) (string, error)
}

func NewEncryptedRegistry(reg ConfigurationContext, publicKey *keys.PublicKey, privateKey *keys.PrivateKey) *EncryptedRegistry {
	return &EncryptedRegistry{
		etcdRegistry: reg,
		encrypt:      publicKey.Encrypt,
		decrypt:      privateKey.Decrypt,
	}
}

func (e EncryptedRegistry) Set(key, value string) error {
	encryptedValue, err := e.encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt key %s: %w", key, err)
	}

	return e.etcdRegistry.Set(key, encryptedValue)
}

func (e EncryptedRegistry) SetWithLifetime(key, value string, timeToLiveInSeconds int) error {
	encryptedValue, err := e.encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt key %s: %w", key, err)
	}

	return e.etcdRegistry.SetWithLifetime(key, encryptedValue, timeToLiveInSeconds)
}

func (e EncryptedRegistry) Refresh(key string, timeToLiveInSeconds int) error {
	return e.etcdRegistry.Refresh(key, timeToLiveInSeconds)
}

func (e EncryptedRegistry) Get(key string) (string, error) {
	value, err := e.etcdRegistry.Get(key)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt key %s: %w", key, err)
	}

	return e.decrypt(value)
}

func (e EncryptedRegistry) GetAll() (map[string]string, error) {
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

func (e EncryptedRegistry) Delete(key string) error {
	return e.etcdRegistry.Delete(key)
}

func (e EncryptedRegistry) DeleteRecursive(key string) error {
	return e.etcdRegistry.DeleteRecursive(key)
}

func (e EncryptedRegistry) Exists(key string) (bool, error) {
	return e.etcdRegistry.Exists(key)
}

func (e EncryptedRegistry) RemoveAll() error {
	return e.etcdRegistry.RemoveAll()
}

func (e EncryptedRegistry) GetOrFalse(key string) (bool, string, error) {
	exists, value, err := e.etcdRegistry.GetOrFalse(key)
	if err != nil {
		return false, "", fmt.Errorf("failed to decrypt key %s: %w", key, err)
	}

	decryptedValue, err := e.decrypt(value)

	return exists, decryptedValue, err
}
