package registry

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

type testcase int

const (
	validReturn testcase = iota
	configDataKeyNotFound
	returnNotFound
	returnOtherError
	notCalled
)

func TestConfigType_String(t *testing.T) {
	tests := []struct {
		name      string
		input     configType
		expOutput string
	}{
		{"global config", globalConfigType, "global-config"},
		{"dogu config", doguConfigType, "dogu-config"},
		{"sensitive config", sensitiveConfigType, "sensitive-config"},
		{"unknown config", unknown, "unknown"},
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
		name      string
		tc        testcase
		xErr      bool
		xErrValue error
	}{
		{
			name:      "Get",
			tc:        validReturn,
			xErr:      false,
			xErrValue: nil,
		},
		{
			name:      "No config found in configMap",
			tc:        configDataKeyNotFound,
			xErr:      true,
			xErrValue: nil,
		},
		{
			name:      "Return Error: Not Found",
			tc:        returnNotFound,
			xErr:      true,
			xErrValue: ErrConfigNotFound,
		},
		{
			name:      "Return Error",
			tc:        returnOtherError,
			xErr:      true,
			xErrValue: nil,
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

			if tc.xErrValue != nil {
				assert.True(t, errors.Is(err, tc.xErrValue))
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
				Data: map[string]string{dataKeyName: "testString"},
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

			err := client.Create(context.TODO(), "", "")
			assert.Equal(t, tc.xErr, err != nil)
		})
	}
}

func TestConfigMapClient_Update(t *testing.T) {
	applyTestCase := func(m *MockConfigMapClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(&v1.ConfigMap{
				Data: map[string]string{dataKeyName: "testString"},
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
			name: "Update",
			tc:   validReturn,
			cd: clientData{
				dataStr: "testUpdate",
				rawData: &v1.ConfigMap{
					Data: map[string]string{dataKeyName: "testString"},
				},
			},
			xErr: false,
		},
		{
			name: "client data with wrong raw Data",
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

			err := client.Update(context.TODO(), tc.cd)
			assert.Equal(t, tc.xErr, err != nil)
		})
	}
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
		name      string
		tc        testcase
		xErr      bool
		xErrValue error
	}{
		{
			name:      "Get",
			tc:        validReturn,
			xErr:      false,
			xErrValue: nil,
		},
		{
			name:      "No config found in secret",
			tc:        configDataKeyNotFound,
			xErr:      true,
			xErrValue: nil,
		},
		{
			name:      "Return Error: Not Found",
			tc:        returnNotFound,
			xErr:      true,
			xErrValue: ErrConfigNotFound,
		},
		{
			name:      "Return Error",
			tc:        returnOtherError,
			xErr:      true,
			xErrValue: nil,
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

			if tc.xErrValue != nil {
				assert.True(t, errors.Is(err, tc.xErrValue))
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
				Data: map[string][]byte{dataKeyName: []byte("testString")},
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

			err := client.Create(context.TODO(), "", "")
			assert.Equal(t, tc.xErr, err != nil)
		})
	}
}

func TestSecretClient_Update(t *testing.T) {
	applyTestCase := func(m *MockSecretClient, tc testcase) {
		switch tc {
		case validReturn:
			m.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(&v1.Secret{
				Data: map[string][]byte{dataKeyName: []byte("testString")},
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
			name: "Update",
			tc:   validReturn,
			cd: clientData{
				dataStr: "testUpdate",
				rawData: &v1.Secret{},
			},
			xErr: false,
		},
		{
			name: "client data with wrong raw Data",
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

			err := client.Update(context.TODO(), tc.cd)
			assert.Equal(t, tc.xErr, err != nil)
		})
	}
}
