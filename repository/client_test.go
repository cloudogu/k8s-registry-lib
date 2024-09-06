package repository

import (
	"context"
	"errors"
	liberrors "github.com/cloudogu/k8s-registry-lib/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"testing"
	"time"
)

type testcase int

const (
	validReturn testcase = iota
	configDataKeyNotFound
	returnNotFound
	returnOtherError
	notCalled
	noItems
)

const resourceVersion = "testVersion"

func TestConfigType_String(t *testing.T) {
	tests := []struct {
		name      string
		input     configType
		expOutput string
	}{
		{"global config", globalConfigType, "global-config"},
		{"dogu config", doguConfigType, "dogu-config"},
		{"sensitive config", sensitiveConfigType, "sensitive-config"},
		{"unknown config", 0, "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expOutput, tc.input.String())
		})
	}
}

func TestCreateConfigMapClient(t *testing.T) {
	tests := []struct {
		name string
		m    ConfigMapClient
		in   configType
	}{
		{"global client", NewMockConfigMapClient(t), globalConfigType},
		{"dogu client", NewMockConfigMapClient(t), doguConfigType},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := createConfigMapClient(tc.m, tc.in)

			assert.NotNil(t, c)
			assert.NotNil(t, c.client)

			assert.Equal(t, appLabelValueCes, c.labels.Get(appLabelKey))
			assert.Equal(t, tc.in.String(), c.labels.Get(typeLabelKey))
		})
	}

}

func TestCreateSecretClient(t *testing.T) {
	tests := []struct {
		name string
		m    SecretClient
		in   configType
	}{
		{"secret client", NewMockSecretClient(t), sensitiveConfigType},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := createSecretClient(tc.m, tc.in)

			assert.NotNil(t, c)
			assert.NotNil(t, c.client)

			assert.Equal(t, appLabelValueCes, c.labels.Get(appLabelKey))
			assert.Equal(t, tc.in.String(), c.labels.Get(typeLabelKey))
		})
	}

}

func TestConfigMapClient_createConfigMap(t *testing.T) {
	t.Run("create with doguName", func(t *testing.T) {
		client := configMapClient{
			client: nil,
			labels: make(labels.Set),
		}

		cm := client.createConfigMap(resourceVersion, "test-config", "test", "testValue")
		assert.Equal(t, "test-config", cm.GetName())
		assert.Equal(t, "test", cm.GetLabels()[doguNameLabelKey])
		assert.Equal(t, "testValue", cm.Data[dataKeyName])
		assert.Equal(t, resourceVersion, cm.GetResourceVersion())
	})

	t.Run("create without doguName", func(t *testing.T) {
		client := configMapClient{
			client: nil,
			labels: make(labels.Set),
		}

		cm := client.createConfigMap(resourceVersion, "test-config", "", "testValue")
		assert.Equal(t, "test-config", cm.GetName())
		_, ok := cm.GetLabels()[doguNameLabelKey]
		assert.False(t, ok)
		assert.Equal(t, "testValue", cm.Data[dataKeyName])
		assert.Equal(t, resourceVersion, cm.GetResourceVersion())
	})
}

func TestConfigMapClient_Get(t *testing.T) {
	applyTestCase := func(m *MockConfigMapClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&v1.ConfigMap{
				Data: map[string]string{dataKeyName: "testString"},
			}, nil)
		case configDataKeyNotFound:
			m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&v1.ConfigMap{}, nil)
		case returnNotFound:
			m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))
		case returnOtherError:
			m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		default:
		}
	}

	tests := []struct {
		name   string
		tc     testcase
		xErr   bool
		valErr func(error) bool
	}{
		{
			name:   "Get",
			tc:     validReturn,
			xErr:   false,
			valErr: nil,
		},
		{
			name:   "No config found in configMap",
			tc:     configDataKeyNotFound,
			xErr:   true,
			valErr: nil,
		},
		{
			name:   "Return Error: Not Found",
			tc:     returnNotFound,
			xErr:   true,
			valErr: liberrors.IsNotFoundError,
		},
		{
			name:   "Return Error",
			tc:     returnOtherError,
			xErr:   true,
			valErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTestCase(m, tc.tc)

			client := configMapClient{
				client: m,
			}

			_, err := client.Get(context.TODO(), "")
			assert.Equal(t, tc.xErr, err != nil)

			if tc.valErr != nil {
				assert.True(t, tc.valErr(err))
			}
		})
	}
}

func TestConfigmapClient_GetWithListResourceVersion(t *testing.T) {
	applyTestCase := func(m *MockConfigMapClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(&v1.ConfigMapList{
				ListMeta: metav1.ListMeta{
					ResourceVersion: "resourceVersion",
				},
				Items: []v1.ConfigMap{
					{
						Data: map[string]string{dataKeyName: "testString"},
					},
				},
			}, nil)
		case configDataKeyNotFound:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(&v1.ConfigMapList{
				Items: []v1.ConfigMap{{Data: map[string]string{}}},
			}, nil)
		case returnNotFound:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))
		case returnOtherError:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		case noItems:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(&v1.ConfigMapList{}, nil)
		default:
		}
	}

	tests := []struct {
		name   string
		tc     testcase
		xErr   bool
		valErr func(error) bool
	}{
		{
			name:   "List",
			tc:     validReturn,
			xErr:   false,
			valErr: nil,
		},
		{
			name:   "No config found in secret",
			tc:     configDataKeyNotFound,
			xErr:   true,
			valErr: nil,
		},
		{
			name:   "Return Error: Not Found",
			tc:     returnNotFound,
			xErr:   true,
			valErr: liberrors.IsNotFoundError,
		},
		{
			name:   "Return Error",
			tc:     returnOtherError,
			xErr:   true,
			valErr: nil,
		},
		{
			name:   "No items in secret list",
			tc:     noItems,
			xErr:   true,
			valErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTestCase(m, tc.tc)

			client := configMapClient{
				client: m,
			}

			_, resourceVersion, err := client.GetWithListResourceVersion(context.TODO(), "")
			assert.Equal(t, tc.xErr, err != nil)
			assert.Equal(t, tc.xErr, len(resourceVersion) == 0)

			if tc.valErr != nil {
				assert.True(t, tc.valErr(err))
			}
		})
	}
}

func TestConfigMapClient_Delete(t *testing.T) {
	applyTestCase := func(m *MockConfigMapClient, tc testcase) {
		switch tc {
		case returnNotFound:
			m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))
		case returnOtherError:
			m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("testErr"))
		default:
			m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		}
	}

	tests := []struct {
		name string
		tc   testcase
		xErr bool
	}{
		{
			name: "Delete",
			tc:   validReturn,
			xErr: false,
		},
		{
			name: "Return Error: Not Found",
			tc:   returnNotFound,
			xErr: false,
		},
		{
			name: "Return Error",
			tc:   returnOtherError,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTestCase(m, tc.tc)

			client := configMapClient{
				client: m,
			}

			err := client.Delete(context.TODO(), "")
			assert.Equal(t, tc.xErr, err != nil)
		})
	}
}

func TestConfigMapClient_Create(t *testing.T) {
	applyTestCase := func(m *MockConfigMapClient, tc testcase) {
		switch tc {
		case returnOtherError:
			m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		default:
			m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string]string{dataKeyName: "testString"},
			}, nil)
		}
	}

	tests := []struct {
		name string
		tc   testcase
		xErr bool
	}{
		{
			name: "Create",
			tc:   validReturn,
			xErr: false,
		},
		{
			name: "Return Error",
			tc:   returnOtherError,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTestCase(m, tc.tc)

			client := configMapClient{
				client: m,
			}

			cm, err := client.Create(context.TODO(), "", "", "")
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				assert.NotNil(t, cm)
				assert.Equal(t, resourceVersion, cm.GetResourceVersion())
			}
		})
	}
}

func TestConfigMapClient_Update(t *testing.T) {
	applyTestCase := func(m *MockConfigMapClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string]string{dataKeyName: "testString"},
			}, nil)
		case returnOtherError:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		default:
		}
	}

	tests := []struct {
		name string
		tc   testcase
		xErr bool
	}{
		{
			name: "Update",
			tc:   validReturn,
			xErr: false,
		},
		{
			name: "Return Error",
			tc:   returnOtherError,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTestCase(m, tc.tc)

			client := configMapClient{
				client: m,
			}

			cm, err := client.Update(context.TODO(), "", "", "", "")
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				assert.NotNil(t, cm)
				assert.Equal(t, resourceVersion, cm.GetResourceVersion())
			}
		})
	}
}

func TestConfigMapClient_UpdateClientData(t *testing.T) {
	applyTestCase := func(m *MockConfigMapClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string]string{dataKeyName: "testString"},
			}, nil)
		case returnOtherError:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		default:
		}
	}

	tests := []struct {
		name string
		tc   testcase
		cd   clientData
		xErr bool
	}{
		{
			name: "UpdateClientData",
			tc:   validReturn,
			cd: clientData{
				dataStr: "testUpdate",
				rawData: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
					Data:       map[string]string{dataKeyName: "testString"},
				},
			},
			xErr: false,
		},
		{
			name: "client data with wrong raw entries",
			tc:   notCalled,
			cd: clientData{
				dataStr: "testData",
				rawData: &v1.Secret{},
			},
			xErr: true,
		},
		{
			name: "Return Error",
			tc:   returnOtherError,
			cd: clientData{
				dataStr: "testData",
				rawData: &v1.ConfigMap{
					Data: map[string]string{dataKeyName: "testString"},
				},
			},
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockConfigMapClient(t)
			applyTestCase(m, tc.tc)

			client := configMapClient{
				client: m,
			}

			cm, err := client.UpdateClientData(context.TODO(), tc.cd)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				assert.NotNil(t, cm)
				assert.Equal(t, resourceVersion, cm.GetResourceVersion())
			}
		})
	}
}

func TestSecretClient_createSecret(t *testing.T) {
	t.Run("create with doguName", func(t *testing.T) {
		client := secretClient{
			client: nil,
			labels: make(labels.Set),
		}

		s := client.createSecret(resourceVersion, "test-config", "test", "testValue")
		assert.Equal(t, "test-config", s.GetName())
		assert.Equal(t, "test", s.GetLabels()[doguNameLabelKey])
		assert.Equal(t, "testValue", s.StringData[dataKeyName])
		assert.Equal(t, resourceVersion, s.GetResourceVersion())
	})

	t.Run("create without doguName", func(t *testing.T) {
		client := secretClient{
			client: nil,
			labels: make(labels.Set),
		}

		s := client.createSecret(resourceVersion, "test-config", "", "testValue")
		assert.Equal(t, "test-config", s.GetName())
		_, ok := s.GetLabels()[doguNameLabelKey]
		assert.False(t, ok)
		assert.Equal(t, "testValue", s.StringData[dataKeyName])
		assert.Equal(t, resourceVersion, s.GetResourceVersion())
	})
}

func TestSecretClient_Get(t *testing.T) {
	applyTestCase := func(m *MockSecretClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&v1.Secret{
				Data: map[string][]byte{dataKeyName: []byte("testString")},
			}, nil)
		case configDataKeyNotFound:
			m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&v1.Secret{}, nil)
		case returnNotFound:
			m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))
		case returnOtherError:
			m.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		default:
		}
	}

	tests := []struct {
		name   string
		tc     testcase
		xErr   bool
		valErr func(error) bool
	}{
		{
			name:   "Get",
			tc:     validReturn,
			xErr:   false,
			valErr: nil,
		},
		{
			name:   "No config found in secret",
			tc:     configDataKeyNotFound,
			xErr:   true,
			valErr: nil,
		},
		{
			name:   "Return Error: Not Found",
			tc:     returnNotFound,
			xErr:   true,
			valErr: liberrors.IsNotFoundError,
		},
		{
			name:   "Return Error",
			tc:     returnOtherError,
			xErr:   true,
			valErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockSecretClient(t)
			applyTestCase(m, tc.tc)

			client := secretClient{
				client: m,
			}

			_, err := client.Get(context.TODO(), "")
			assert.Equal(t, tc.xErr, err != nil)

			if tc.valErr != nil {
				assert.True(t, tc.valErr(err))
			}
		})
	}
}

func TestSecretClient_GetWithListResourceVersion(t *testing.T) {
	applyTestCase := func(m *MockSecretClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(&v1.SecretList{
				ListMeta: metav1.ListMeta{
					ResourceVersion: "resourceVersion",
				},
				Items: []v1.Secret{
					{
						Data: map[string][]byte{dataKeyName: []byte("testString")},
					},
				},
			}, nil)
		case configDataKeyNotFound:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(&v1.SecretList{
				Items: []v1.Secret{{Data: map[string][]byte{}}},
			}, nil)
		case returnNotFound:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))
		case returnOtherError:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		case noItems:
			m.EXPECT().List(mock.Anything, mock.Anything).Return(&v1.SecretList{}, nil)
		default:
		}
	}

	tests := []struct {
		name   string
		tc     testcase
		xErr   bool
		valErr func(error) bool
	}{
		{
			name:   "List",
			tc:     validReturn,
			xErr:   false,
			valErr: nil,
		},
		{
			name:   "No config found in secret",
			tc:     configDataKeyNotFound,
			xErr:   true,
			valErr: nil,
		},
		{
			name:   "Return Error: Not Found",
			tc:     returnNotFound,
			xErr:   true,
			valErr: liberrors.IsNotFoundError,
		},
		{
			name:   "Return Error",
			tc:     returnOtherError,
			xErr:   true,
			valErr: nil,
		},
		{
			name:   "No items in secret list",
			tc:     noItems,
			xErr:   true,
			valErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockSecretClient(t)
			applyTestCase(m, tc.tc)

			client := secretClient{
				client: m,
			}

			_, resourceVersion, err := client.GetWithListResourceVersion(context.TODO(), "")
			assert.Equal(t, tc.xErr, err != nil)
			assert.Equal(t, tc.xErr, len(resourceVersion) == 0)

			if tc.valErr != nil {
				assert.True(t, tc.valErr(err))
			}
		})
	}
}

func TestSecretClient_Delete(t *testing.T) {
	applyTestCase := func(m *MockSecretClient, tc testcase) {
		switch tc {
		case returnNotFound:
			m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))
		case returnOtherError:
			m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("testErr"))
		default:
			m.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		}
	}

	tests := []struct {
		name string
		tc   testcase
		xErr bool
	}{
		{
			name: "Delete",
			tc:   validReturn,
			xErr: false,
		},
		{
			name: "Return Error: Not Found",
			tc:   returnNotFound,
			xErr: false,
		},
		{
			name: "Return Error",
			tc:   returnOtherError,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockSecretClient(t)
			applyTestCase(m, tc.tc)

			client := secretClient{
				client: m,
			}

			err := client.Delete(context.TODO(), "")
			assert.Equal(t, tc.xErr, err != nil)
		})
	}
}

func TestSecretClient_Create(t *testing.T) {
	applyTestCase := func(m *MockSecretClient, tc testcase) {
		switch tc {
		case returnOtherError:
			m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		default:
			m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string][]byte{dataKeyName: []byte("testString")},
			}, nil)
		}
	}

	tests := []struct {
		name string
		tc   testcase
		xErr bool
	}{
		{
			name: "Create",
			tc:   validReturn,
			xErr: false,
		},
		{
			name: "Return Error",
			tc:   returnOtherError,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockSecretClient(t)
			applyTestCase(m, tc.tc)

			client := secretClient{
				client: m,
			}

			s, err := client.Create(context.TODO(), "", "", "")
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				assert.NotNil(t, s)
				assert.Equal(t, resourceVersion, s.GetResourceVersion())
			}
		})
	}
}

func TestSecretClient_Update(t *testing.T) {
	applyTestCase := func(m *MockSecretClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string][]byte{dataKeyName: []byte("testString")},
			}, nil)
		case returnOtherError:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		default:
		}
	}

	tests := []struct {
		name string
		tc   testcase
		xErr bool
	}{
		{
			name: "Update",
			tc:   validReturn,
			xErr: false,
		},
		{
			name: "Return Error",
			tc:   returnOtherError,
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockSecretClient(t)
			applyTestCase(m, tc.tc)

			client := secretClient{
				client: m,
			}

			cm, err := client.Update(context.TODO(), "", "", "", "")
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				assert.NotNil(t, cm)
				assert.Equal(t, resourceVersion, cm.GetResourceVersion())
			}
		})
	}
}

func TestSecretClient_UpdateClientData(t *testing.T) {
	applyTestCase := func(m *MockSecretClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string][]byte{dataKeyName: []byte("testString")},
			}, nil)
		case returnOtherError:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("testErr"))
		default:
		}
	}

	tests := []struct {
		name string
		tc   testcase
		cd   clientData
		xErr bool
	}{
		{
			name: "UpdateClientData",
			tc:   validReturn,
			cd: clientData{
				dataStr: "testUpdate",
				rawData: &v1.Secret{},
			},
			xErr: false,
		},
		{
			name: "client data with wrong raw entries",
			tc:   notCalled,
			cd: clientData{
				dataStr: "testData",
				rawData: &v1.ConfigMap{},
			},
			xErr: true,
		},
		{
			name: "Return Error",
			tc:   returnOtherError,
			cd: clientData{
				dataStr: "testData",
				rawData: &v1.Secret{},
			},
			xErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMockSecretClient(t)
			applyTestCase(m, tc.tc)

			client := secretClient{
				client: m,
			}

			s, err := client.UpdateClientData(context.TODO(), tc.cd)
			assert.Equal(t, tc.xErr, err != nil)

			if err == nil {
				assert.NotNil(t, s)
				assert.Equal(t, resourceVersion, s.GetResourceVersion())
			}
		})
	}
}

func Test_watchWithClient(t *testing.T) {
	listOptions := metav1.SingleObject(metav1.ObjectMeta{Name: "dogu-config"})
	listOptions.ResourceVersion = resourceVersion
	listOptions.AllowWatchBookmarks = true

	t.Run("should watch with client", func(t *testing.T) {
		ctx := context.Background()
		fakeWatcher := watch.NewFake()

		mockWatcher := newMockClientWatcher(t)
		mockWatcher.EXPECT().Watch(ctx, listOptions).Return(fakeWatcher, nil)

		watchChan, err := watchWithClient(ctx, mockWatcher, "dogu-config", resourceVersion)
		require.NoError(t, err)
		require.NotNil(t, watchChan)

		cancel := make(chan bool, 1)

		go func() {
			fakeWatcher.Modify(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string][]byte{dataKeyName: []byte("test-data-secret")},
			})
			fakeWatcher.Modify(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string]string{dataKeyName: "test-data-configmap"},
			})
			fakeWatcher.Modify(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
			})
		}()

		go func() {
			i := 0
			for result := range watchChan {
				if i == 0 {
					assert.NoError(t, result.err)
					assert.Equal(t, resourceVersion, result.persistentContext)
					assert.Equal(t, "test-data-secret", result.dataStr)
				}

				if i == 1 {
					assert.NoError(t, result.err)
					assert.Equal(t, resourceVersion, result.persistentContext)
					assert.Equal(t, "test-data-configmap", result.dataStr)
				}

				if i == 2 {
					assert.Error(t, result.err)
					assert.ErrorContains(t, result.err, "unsupported type in watch *v1.Pod")
					cancel <- true
				}

				i++
			}
		}()

		select {
		case <-cancel:
			fakeWatcher.Stop()
		case <-time.After(5 * time.Second):
			fakeWatcher.Stop()
			t.Errorf("did not reach third event in time")
		}
	})

	t.Run("should write error in result for missing data", func(t *testing.T) {
		ctx := context.Background()
		fakeWatcher := watch.NewFake()

		mockWatcher := newMockClientWatcher(t)
		mockWatcher.EXPECT().Watch(ctx, listOptions).Return(fakeWatcher, nil)

		watchChan, err := watchWithClient(ctx, mockWatcher, "dogu-config", resourceVersion)
		require.NoError(t, err)
		require.NotNil(t, watchChan)

		cancel := make(chan bool, 1)

		go func() {
			fakeWatcher.Modify(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string][]byte{"foo": []byte("test-data-secret")},
			})
			fakeWatcher.Modify(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string]string{"foo": "test-data-configmap"},
			})
		}()

		go func() {
			i := 0
			for result := range watchChan {
				if i == 0 {
					assert.Error(t, result.err)
					assert.ErrorContains(t, result.err, "could not find data for key config.yaml in secret dogu-config")
				}

				if i == 1 {
					assert.Error(t, result.err)
					assert.ErrorContains(t, result.err, "could not find data for key config.yaml in configmap dogu-config")
					cancel <- true
				}

				i++
			}
		}()

		select {
		case <-cancel:
			fakeWatcher.Stop()
		case <-time.After(5 * time.Second):
			fakeWatcher.Stop()
			t.Errorf("did not reach the second event in time")
		}
	})

	t.Run("should write error in result for error in watch-channel", func(t *testing.T) {
		ctx := context.Background()
		fakeWatcher := watch.NewFake()

		mockWatcher := newMockClientWatcher(t)
		mockWatcher.EXPECT().Watch(ctx, listOptions).Return(fakeWatcher, nil)

		watchChan, err := watchWithClient(ctx, mockWatcher, "dogu-config", resourceVersion)
		require.NoError(t, err)
		require.NotNil(t, watchChan)

		cancel := make(chan bool, 1)

		go func() {
			fakeWatcher.Error(&metav1.Status{})
			// Wait until channel is closed due to error to reset accordingly afterwards
			for !fakeWatcher.IsStopped() {
				time.Sleep(1 * time.Millisecond)
			}
			fakeWatcher.Reset()

			fakeWatcher.Modify(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: resourceVersion},
				Data:       map[string]string{dataKeyName: "test-data-configmap"},
			})
		}()

		go func() {
			for result := range watchChan {
				assert.NoError(t, result.err)
				assert.Equal(t, resourceVersion, result.persistentContext)
				assert.Equal(t, "test-data-configmap", result.dataStr)
				cancel <- true
			}
		}()

		select {
		case <-cancel:
			fakeWatcher.Stop()
		case <-time.After(5 * time.Second):
			fakeWatcher.Stop()
			t.Errorf("did not reach the event in time")
		}
	})

	t.Run("should stop watch on context-cancel", func(t *testing.T) {
		ctx := context.Background()
		fakeWatcher := watch.NewFake()
		cancelCtx, cancelCtxFunc := context.WithCancel(ctx)

		mockWatcher := newMockClientWatcher(t)
		mockWatcher.EXPECT().Watch(cancelCtx, listOptions).Return(fakeWatcher, nil)

		watchChan, err := watchWithClient(cancelCtx, mockWatcher, "dogu-config", resourceVersion)
		require.NoError(t, err)
		require.NotNil(t, watchChan)

		isCanceled := false
		go func() {
			for range watchChan {
				isCanceled = false
			}
			isCanceled = true
		}()

		cancelCtxFunc()

		select {
		// wait for watcher to stopped
		case <-time.After(200 * time.Millisecond):
			assert.True(t, isCanceled)
			assert.True(t, fakeWatcher.IsStopped())
		}
	})

	t.Run("should return error for error when starting watch", func(t *testing.T) {
		ctx := context.Background()
		mockWatcher := newMockClientWatcher(t)

		_, err := watchWithClient(ctx, mockWatcher, "dogu-config", "")

		require.Error(t, err)
		assert.ErrorContains(t, err, "could not watch 'dogu-config' in cluster:")
	})
}

func Test_handleWatchEvent(t *testing.T) {
	t.Run("should return status error for error of type status", func(t *testing.T) {
		//given
		event := watch.Event{
			Type:   watch.Error,
			Object: &metav1.Status{Status: metav1.StatusFailure},
		}

		//when
		watchResult := handleWatchEvent("testName", event)

		//then
		assert.Empty(t, watchResult.dataStr)
		assert.Empty(t, watchResult.persistentContext)
		assert.Error(t, watchResult.err)
		assert.ErrorContains(t, watchResult.err, "watch event type is error:")
		assert.ErrorContains(t, watchResult.err, metav1.StatusFailure)
	})
	t.Run("should return generic error for error of type other than status", func(t *testing.T) {
		//given
		event := watch.Event{
			Type:   watch.Error,
			Object: &metav1.List{},
		}

		//when
		watchResult := handleWatchEvent("testName", event)

		//then
		assert.Empty(t, watchResult.dataStr)
		assert.Empty(t, watchResult.persistentContext)
		assert.Error(t, watchResult.err)
		assert.ErrorContains(t, watchResult.err, "error result in watcher for config 'testName'")
	})
}

func Test_secretClient_Watch(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error for error when starting watch", func(t *testing.T) {
		mockClient := NewMockSecretClient(t)
		client := secretClient{
			client: mockClient,
		}

		_, err := client.Watch(ctx, "dogu-config", "")

		require.Error(t, err)
		assert.ErrorContains(t, err, "could not watch 'dogu-config' in cluster:")
	})
}

func Test_configMapClient_Watch(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error for error when starting watch", func(t *testing.T) {
		mockClient := NewMockConfigMapClient(t)
		client := configMapClient{
			client: mockClient,
		}

		_, err := client.Watch(ctx, "dogu-config", "")

		require.Error(t, err)
		assert.ErrorContains(t, err, "could not watch 'dogu-config' in cluster:")
	})
}

func Test_handleError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		valErr func(error) bool
	}{
		{
			name:   "NotFoundErr",
			err:    k8serrors.NewNotFound(schema.GroupResource{}, ""),
			valErr: liberrors.IsNotFoundError,
		},
		{
			name:   "ConflictErr",
			err:    k8serrors.NewConflict(schema.GroupResource{}, "", assert.AnError),
			valErr: liberrors.IsConflictError,
		},
		{
			name:   "ServerTimeoutErr",
			err:    k8serrors.NewServerTimeout(schema.GroupResource{}, "", 0),
			valErr: liberrors.IsConnectionError,
		},
		{
			name:   "TimeoutErr",
			err:    k8serrors.NewTimeoutError("", 0),
			valErr: liberrors.IsConnectionError,
		},
		{
			name:   "AlreadyExistsErr",
			err:    k8serrors.NewAlreadyExists(schema.GroupResource{}, ""),
			valErr: liberrors.IsAlreadyExistsError,
		},
		{
			name:   "InternetErr",
			err:    k8serrors.NewInternalError(assert.AnError),
			valErr: liberrors.IsGenericError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := handleError(tc.err)
			assert.True(t, tc.valErr(err))
		})
	}
}
