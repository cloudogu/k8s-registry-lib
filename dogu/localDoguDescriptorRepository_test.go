package dogu

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	"github.com/cloudogu/k8s-registry-lib/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

const (
	casVersionStr         = "7.0.5.1-1"
	ldapVersionStr        = "2.6.7-3"
	upgradeLdapVersionStr = "2.6.8-3"
)

var testCtx = context.Background()

func TestNewLocalDoguDescriptorRepository(t *testing.T) {
	// given
	configMapClientMock := newMockConfigMapClient(t)

	// when
	sut := NewLocalDoguDescriptorRepository(configMapClientMock)

	// then
	require.NotNil(t, sut)
	assert.Equal(t, configMapClientMock, sut.configMapClient)
}

func Test_localDoguDescriptorRepository_Add(t *testing.T) {
	casDogu := readCasDogu(t)
	expectedCasRegistryCm := &corev1.ConfigMap{Data: map[string]string{casVersionStr: readCasDoguStr(t)}}

	type args struct {
		ctx  context.Context
		name SimpleDoguName
		dogu *core.Dogu
	}
	tests := []struct {
		name              string
		configMapClientFn func(t *testing.T) configMapClient
		args              args
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "success with existent config",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(&corev1.ConfigMap{}, nil)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCm, metav1.UpdateOptions{}).Return(expectedCasRegistryCm, nil)

				return configMapClientMock
			},
			args:    args{ctx: testCtx, name: "cas", dogu: casDogu},
			wantErr: assert.NoError,
		},
		{
			name: "should create dogu descriptor config map if not existent",
			configMapClientFn: func(t *testing.T) configMapClient {
				cmToCreate := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: "dogu-spec-cas",
						Labels: map[string]string{
							"app":                   "ces",
							"dogu.name":             "cas",
							"k8s.cloudogu.com/type": "local-dogu-registry",
						},
					},
				}

				expectedUpdateCm := cmToCreate.DeepCopy()
				expectedUpdateCm.Data = map[string]string{casVersionStr: readCasDoguStr(t)}

				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(nil, apierrors.NewNotFound(schema.GroupResource{}, ""))
				configMapClientMock.EXPECT().Create(testCtx, cmToCreate, metav1.CreateOptions{}).Return(cmToCreate, nil)
				configMapClientMock.EXPECT().Update(testCtx, expectedUpdateCm, metav1.UpdateOptions{}).Return(expectedUpdateCm, nil)

				return configMapClientMock
			},
			args:    args{ctx: testCtx, name: "cas", dogu: casDogu},
			wantErr: assert.NoError,
		},
		{
			name: "should return error on not existent dogu descriptor configmap",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(&corev1.ConfigMap{}, assert.AnError)

				return configMapClientMock
			},
			args: args{ctx: testCtx, name: "cas", dogu: casDogu},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, errors.IsGenericError(err)) &&
					assert.ErrorContains(t, err, "failed to get dogu descriptor config map for dogu \"cas\"")
			},
		},
		{
			name: "should return error if the descriptor already exists",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(&corev1.ConfigMap{Data: map[string]string{casVersionStr: "exists"}}, nil)

				return configMapClientMock
			},
			args: args{ctx: testCtx, name: "cas", dogu: casDogu},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, errors.IsAlreadyExistsError(err)) &&
					assert.ErrorContains(t, err, "\"cas\" dogu descriptor already exists for version \"7.0.5.1-1\"")
			},
		},
		{
			name: "should return error on update configmap error",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(&corev1.ConfigMap{}, nil)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCm, metav1.UpdateOptions{}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			args: args{ctx: testCtx, name: "cas", dogu: casDogu},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err) &&
					assert.ErrorContains(t, err, "failed to update dogu descriptor configmap for dogu \"cas\"")
			},
		},
		{
			name: "should retry on conflict error",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(&corev1.ConfigMap{}, nil).Times(1)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(&corev1.ConfigMap{}, nil).Times(1)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCm, metav1.UpdateOptions{}).Return(nil, testConflictErr).Times(1)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCm, metav1.UpdateOptions{}).Return(expectedCasRegistryCm, nil).Times(1)

				return configMapClientMock
			},
			args:    args{ctx: testCtx, name: "cas", dogu: casDogu},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &localDoguDescriptorRepository{
				configMapClient: tt.configMapClientFn(t),
			}
			tt.wantErr(t, vr.Add(tt.args.ctx, tt.args.name, tt.args.dogu), fmt.Sprintf("Add(%v, %v, %v)", tt.args.ctx, tt.args.name, tt.args.dogu))
		})
	}
}

func Test_localDoguDescriptorRepository_DeleteAll(t *testing.T) {
	type args struct {
		ctx  context.Context
		name SimpleDoguName
	}
	tests := []struct {
		name              string
		configMapClientFn func(t *testing.T) configMapClient
		args              args
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Delete(testCtx, "dogu-spec-cas", metav1.DeleteOptions{}).Return(nil)

				return configMapClientMock
			},
			args: args{
				ctx:  testCtx,
				name: "cas",
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error on delete error",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Delete(testCtx, "dogu-spec-cas", metav1.DeleteOptions{}).Return(assert.AnError)

				return configMapClientMock
			},
			args: args{
				ctx:  testCtx,
				name: "cas",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, errors.IsGenericError(err)) &&
					assert.ErrorContains(t, err, "failed to delete dogu descriptor configmap for dogu \"cas\"")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &localDoguDescriptorRepository{
				configMapClient: tt.configMapClientFn(t),
			}
			tt.wantErr(t, vr.DeleteAll(tt.args.ctx, tt.args.name), fmt.Sprintf("DeleteAll(%v, %v)", tt.args.ctx, tt.args.name))
		})
	}
}

func Test_localDoguDescriptorRepository_Get(t *testing.T) {
	casVersion := parseVersionStr(t, casVersionStr)
	doguVersion := DoguVersion{
		Name:    "cas",
		Version: casVersion,
	}

	notFoundDoguVersion := DoguVersion{
		Name:    "cas",
		Version: parseVersionStr(t, "1.11.12-1"),
	}

	casDogu := readCasDogu(t)
	casRegistryCm := &corev1.ConfigMap{Data: map[string]string{casVersionStr: string(casBytes)}}
	invalidCasRegistryCm := &corev1.ConfigMap{Data: map[string]string{casVersionStr: "not valid"}}

	type args struct {
		ctx         context.Context
		doguVersion DoguVersion
	}
	tests := []struct {
		name              string
		configMapClientFn func(t *testing.T) configMapClient
		args              args
		want              *core.Dogu
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCm, nil)
				return configMapClientMock
			},
			args:    args{ctx: testCtx, doguVersion: doguVersion},
			want:    casDogu,
			wantErr: assert.NoError,
		},
		{
			name: "should return error on error getting dogu registry config map",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(nil, assert.AnError)
				return configMapClientMock
			},
			args: args{ctx: testCtx, doguVersion: doguVersion},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, errors.IsGenericError(err)) &&
					assert.ErrorContains(t, err, "failed to get dogu descriptor config map for dogu \"cas\"")
			},
		},
		{
			name: "should return error on if the key is not found",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCm, nil)
				return configMapClientMock
			},
			args: args{ctx: testCtx, doguVersion: notFoundDoguVersion},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, errors.IsNotFoundError(err)) &&
					assert.ErrorContains(t, err, "failed to get value for key \"1.11.12-1\" for dogu registry \"cas\"")
			},
		},
		{
			name: "should return error on invalid dogu descriptor",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(invalidCasRegistryCm, nil)
				return configMapClientMock
			},
			args: args{ctx: testCtx, doguVersion: doguVersion},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, errors.IsGenericError(err)) &&
					assert.ErrorContains(t, err, "failed to unmarshal descriptor for dogu \"cas\" with version \"7.0.5.1-1\"")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &localDoguDescriptorRepository{
				configMapClient: tt.configMapClientFn(t),
			}
			got, err := vr.Get(tt.args.ctx, tt.args.doguVersion)
			if !tt.wantErr(t, err, fmt.Sprintf("Get(%v, %v)", tt.args.ctx, tt.args.doguVersion)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.ctx, tt.args.doguVersion)
		})
	}
}

func Test_localDoguDescriptorRepository_GetAll(t *testing.T) {
	casVersion := parseVersionStr(t, casVersionStr)
	ldapVersion := parseVersionStr(t, ldapVersionStr)
	casDogu := readCasDogu(t)
	ldapDogu := readLdapDogu(t)
	casRegistryCm := &corev1.ConfigMap{Data: map[string]string{casVersionStr: string(casBytes)}}
	ldapRegistryCm := &corev1.ConfigMap{Data: map[string]string{ldapVersionStr: string(ldapBytes)}}
	casDoguVersion := DoguVersion{Name: SimpleDoguName(casDogu.GetSimpleName()), Version: casVersion}
	ldapDoguVersion := DoguVersion{Name: SimpleDoguName(ldapDogu.GetSimpleName()), Version: ldapVersion}
	notFoundCasDoguVersion := DoguVersion{Name: SimpleDoguName(casDogu.GetSimpleName()), Version: parseVersionStr(t, "1.222.11-1")}
	notFoundLdapDoguVersion := DoguVersion{Name: SimpleDoguName(ldapDogu.GetSimpleName()), Version: parseVersionStr(t, "1.222.11-1")}
	doguVersions := []DoguVersion{casDoguVersion, ldapDoguVersion}
	notFoundDoguVersions := []DoguVersion{notFoundCasDoguVersion, notFoundLdapDoguVersion}

	expectedDoguVersionMap := map[DoguVersion]*core.Dogu{casDoguVersion: casDogu, ldapDoguVersion: ldapDogu}

	invalidCasRegistryCm := &corev1.ConfigMap{Data: map[string]string{casVersionStr: "not valid"}}
	invalidLdapRegistryCm := &corev1.ConfigMap{Data: map[string]string{ldapVersionStr: "not valid"}}

	type args struct {
		ctx          context.Context
		doguVersions []DoguVersion
	}
	tests := []struct {
		name              string
		configMapClientFn func(t *testing.T) configMapClient
		args              args
		want              map[DoguVersion]*core.Dogu
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCm, nil)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-ldap", metav1.GetOptions{}).Return(ldapRegistryCm, nil)

				return configMapClientMock
			},
			args:    args{ctx: testCtx, doguVersions: doguVersions},
			want:    expectedDoguVersionMap,
			wantErr: assert.NoError,
		},
		{
			name: "should return multi error on error getting registry configmaps",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(nil, assert.AnError)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-ldap", metav1.GetOptions{}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			args: args{ctx: testCtx, doguVersions: doguVersions},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, errors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to get some dogu descriptors:") &&
					assert.ErrorContains(t, err, "failed to get dogu descriptor config map for dogu \"ldap\": assert.AnError general error for testing") &&
					assert.ErrorContains(t, err, "failed to get dogu descriptor config map for dogu \"cas\": assert.AnError general error for testing")
			},
		},
		{
			name: "should return multi error on invalid dogu descriptors",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(invalidCasRegistryCm, nil)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-ldap", metav1.GetOptions{}).Return(invalidLdapRegistryCm, nil)

				return configMapClientMock
			},
			args: args{ctx: testCtx, doguVersions: doguVersions},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, errors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to get some dogu descriptors:") &&
					assert.ErrorContains(t, err, "failed to unmarshal descriptor for dogu \"cas\" with version \"7.0.5.1-1\": invalid character 'o' in literal null (expecting 'u')") &&
					assert.ErrorContains(t, err, "failed to unmarshal descriptor for dogu \"ldap\" with version \"2.6.7-3\": invalid character 'o' in literal null (expecting 'u')")
			},
		},
		{
			name: "should return multi error on not existent dogu descriptors",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCm, nil)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-ldap", metav1.GetOptions{}).Return(ldapRegistryCm, nil)

				return configMapClientMock
			},
			args: args{ctx: testCtx, doguVersions: notFoundDoguVersions},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, errors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to get some dogu descriptors:") &&
					assert.ErrorContains(t, err, "did not find expected version \"1.222.11-1\" for dogu \"cas\" in dogu descriptor configmap") &&
					assert.ErrorContains(t, err, "did not find expected version \"1.222.11-1\" for dogu \"ldap\" in dogu descriptor configmap")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &localDoguDescriptorRepository{
				configMapClient: tt.configMapClientFn(t),
			}
			got, err := vr.GetAll(tt.args.ctx, tt.args.doguVersions)
			if !tt.wantErr(t, err, fmt.Sprintf("GetAll(%v, %v)", tt.args.ctx, tt.args.doguVersions)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetAll(%v, %v)", tt.args.ctx, tt.args.doguVersions)
		})
	}
}
