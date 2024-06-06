package global

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewEncryptedEtcdRegistry(t *testing.T) {
	reg := newEncryptedEtcdRegistry(nil, nil, nil)
	assert.NotNil(t, reg)
}

func mockRegistry(t *testing.T, errEnc error, errDec error) (*encryptedEtcdRegistry, *mockEtcdConfigContext) {
	t.Helper()
	configContext := &mockEtcdConfigContext{}
	return &encryptedEtcdRegistry{
		etcdRegistry: configContext,
		encrypt: func(value string) (string, error) {
			return fmt.Sprintf("e(%s)", value), errEnc
		},
		decrypt: func(value string) (string, error) {
			return fmt.Sprintf("d(%s)", value), errDec
		},
	}, configContext
}

func TestGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Get("mykey").Return("myval", nil)
		get, err := reg.Get("mykey")
		require.Nil(t, err)
		assert.Equal(t, "d(myval)", get)
	})
	t.Run("fail", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Get("mykey").Return("myval", errors.New("testerror"))
		_, err := reg.Get("mykey")
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
		require.Contains(t, err.Error(), "failed to decrypt key mykey")
	})
}
func TestSet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Set("mykey", "e(myval)").Return(nil)
		err := reg.Set("mykey", "myval")
		require.Nil(t, err)
	})
	t.Run("fail", func(t *testing.T) {
		reg, _ := mockRegistry(t, errors.New("testerror"), nil)
		err := reg.Set("mykey", "myval")
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
		require.Contains(t, err.Error(), "failed to encrypt key mykey")
	})
}

func TestSetWithLifetime(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().SetWithLifetime("mykey", "e(myval)", 60).Return(nil)
		err := reg.SetWithLifetime("mykey", "myval", 60)
		require.Nil(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		reg, _ := mockRegistry(t, errors.New("testerror"), nil)
		err := reg.SetWithLifetime("mykey", "myval", 60)
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
		require.Contains(t, err.Error(), "failed to encrypt key mykey")
	})
}

func TestRefresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Refresh("mykey", 60).Return(nil)
		err := reg.Refresh("mykey", 60)
		require.Nil(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Refresh("mykey", 60).Return(errors.New("testerror"))
		err := reg.Refresh("mykey", 60)
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
	})
}

func TestGetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().GetAll().Return(map[string]string{
			"key1": "val1",
			"key2": "val2",
		}, nil)

		all, err := reg.GetAll()
		require.Nil(t, err)
		assert.Equal(t, map[string]string{
			"key1": "d(val1)",
			"key2": "d(val2)",
		}, all)
	})

	t.Run("fail to get all", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().GetAll().Return(nil, errors.New("testerror"))
		_, err := reg.GetAll()
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
		require.Contains(t, err.Error(), "failed to get all keys")
	})

	t.Run("fail to decrypt", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, errors.New("decrypt error"))
		con.EXPECT().GetAll().Return(map[string]string{
			"key1": "e(val1)",
			"key2": "e(val2)",
		}, nil)

		_, err := reg.GetAll()
		require.Error(t, err)
		require.Contains(t, err.Error(), "decrypt error")
		require.Contains(t, err.Error(), "failed to decrypt key key")
	})
}

func TestDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Delete("mykey").Return(nil)
		err := reg.Delete("mykey")
		require.Nil(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Delete("mykey").Return(errors.New("testerror"))
		err := reg.Delete("mykey")
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
	})
}

func TestDeleteRecursive(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().DeleteRecursive("mykey").Return(nil)
		err := reg.DeleteRecursive("mykey")
		require.Nil(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().DeleteRecursive("mykey").Return(errors.New("testerror"))
		err := reg.DeleteRecursive("mykey")
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
	})
}

func TestExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Exists("mykey").Return(true, nil)
		exists, err := reg.Exists("mykey")
		require.Nil(t, err)
		assert.True(t, exists)
	})

	t.Run("not exists", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Exists("mykey").Return(false, nil)
		exists, err := reg.Exists("mykey")
		require.Nil(t, err)
		assert.False(t, exists)
	})

	t.Run("fail", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().Exists("mykey").Return(false, errors.New("testerror"))
		_, err := reg.Exists("mykey")
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
	})
}

func TestRemoveAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().RemoveAll().Return(nil)
		err := reg.RemoveAll()
		require.Nil(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().RemoveAll().Return(errors.New("testerror"))
		err := reg.RemoveAll()
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
	})
}

func TestGetOrFalse(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().GetOrFalse("mykey").Return(true, "myval", nil)
		exists, value, err := reg.GetOrFalse("mykey")
		require.Nil(t, err)
		assert.True(t, exists)
		assert.Equal(t, "d(myval)", value)
	})

	t.Run("not exists", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().GetOrFalse("mykey").Return(false, "myval", nil)
		exists, value, err := reg.GetOrFalse("mykey")
		require.Nil(t, err)
		assert.False(t, exists)
		assert.Equal(t, "d(myval)", value)
	})

	t.Run("fail", func(t *testing.T) {
		reg, con := mockRegistry(t, nil, nil)
		con.EXPECT().GetOrFalse("mykey").Return(false, "", errors.New("testerror"))
		_, _, err := reg.GetOrFalse("mykey")
		require.Error(t, err)
		require.Contains(t, err.Error(), "testerror")
		require.Contains(t, err.Error(), "failed to decrypt key mykey")
	})
}
