package dogu

import (
	"context"
	"fmt"
	cloudoguerrors "github.com/cloudogu/k8s-registry-lib/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
	"time"
)

const (
	versionRegistryLabelSelector = "app=ces,dogu.name,k8s.cloudogu.com/type=local-dogu-registry"
)

var (
	casVersionRegistryLabelMap  = map[string]string{"app": "ces", "dogu.name": "cas", "k8s.cloudogu.com/type": "local-dogu-registry"}
	ldapVersionRegistryLabelMap = map[string]string{"app": "ces", "dogu.name": "ldap", "k8s.cloudogu.com/type": "local-dogu-registry"}
	testConflictErr             = apierrors.NewConflict(schema.GroupResource{}, "conflict", assert.AnError)
)

func TestNewDoguVersionRegistry(t *testing.T) {
	// given
	configMapClientMock := newMockConfigMapClient(t)

	// when
	sut := NewDoguVersionRegistry(configMapClientMock)

	// then
	require.NotNil(t, sut)
	assert.NotNil(t, sut.configMapClient)
}

func Test_versionRegistry_GetCurrent(t *testing.T) {
	expectedDoguVersion := DoguVersion{
		Name:    "cas",
		Version: parseVersionStr(t, casVersionStr),
	}
	casRegistryCm := &corev1.ConfigMap{Data: map[string]string{"current": casVersionStr}}

	type args struct {
		ctx  context.Context
		name SimpleDoguName
	}
	casArgs := args{
		ctx:  testCtx,
		name: "cas",
	}
	tests := []struct {
		name              string
		configMapClientFn func(t *testing.T) configMapClient
		args              args
		want              DoguVersion
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCm, nil)

				return configMapClientMock
			},
			args:    casArgs,
			want:    expectedDoguVersion,
			wantErr: assert.NoError,
		},
		{
			name: "should return error on dogu registry get error",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			args: casArgs,
			want: DoguVersion{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to get dogu descriptor config map for dogu \"cas\"", i)
			},
		},
		{
			name: "should return error if no current key is defined in dogu registry",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(&corev1.ConfigMap{}, nil)

				return configMapClientMock
			},
			args: casArgs,
			want: DoguVersion{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsNotFoundError(err), i) &&
					assert.ErrorContains(t, err, "failed to get value for key \"current\" for dogu registry \"cas\"", i)
			},
		},
		{
			name: "should return error if on invalid current dogu version",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(&corev1.ConfigMap{Data: map[string]string{"current": "abc"}}, nil)

				return configMapClientMock
			},
			args: casArgs,
			want: DoguVersion{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to parse version \"abc\" for dogu \"cas\"", i)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &doguVersionRegistry{
				configMapClient: tt.configMapClientFn(t),
			}
			got, err := vr.GetCurrent(tt.args.ctx, tt.args.name)
			if !tt.wantErr(t, err, fmt.Sprintf("GetCurrent(%v, %v)", tt.args.ctx, tt.args.name)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetCurrent(%v, %v)", tt.args.ctx, tt.args.name)
		})
	}
}

func Test_versionRegistry_GetCurrentOfAll(t *testing.T) {
	expectedCasDoguVersion := DoguVersion{
		Name:    "cas",
		Version: parseVersionStr(t, casVersionStr),
	}
	expectedLdapDoguVersion := DoguVersion{
		Name:    "ldap",
		Version: parseVersionStr(t, ldapVersionStr),
	}
	casRegistryCm := &corev1.ConfigMap{Data: map[string]string{"current": casVersionStr}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
	ldapRegistryCm := &corev1.ConfigMap{Data: map[string]string{"current": ldapVersionStr}, ObjectMeta: metav1.ObjectMeta{Labels: ldapVersionRegistryLabelMap}}
	registryCmList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*casRegistryCm, *ldapRegistryCm}}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name              string
		configMapClientFn func(t *testing.T) configMapClient
		args              args
		want              []DoguVersion
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().List(testCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(registryCmList, nil)

				return configMapClientMock
			},
			args:    args{ctx: testCtx},
			want:    []DoguVersion{expectedCasDoguVersion, expectedLdapDoguVersion},
			wantErr: assert.NoError,
		},
		{
			name: "should success if a dogu is not enabled",
			configMapClientFn: func(t *testing.T) configMapClient {
				notEnabledCasCm := &corev1.ConfigMap{Data: map[string]string{}}
				notEnabledRegistryCmList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*notEnabledCasCm, *ldapRegistryCm}}
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().List(testCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(notEnabledRegistryCmList, nil)

				return configMapClientMock
			},
			args:    args{ctx: testCtx},
			want:    []DoguVersion{expectedLdapDoguVersion},
			wantErr: assert.NoError,
		},
		{
			name: "should return error on error getting all dogu descriptor configmaps",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().List(testCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			args: args{ctx: testCtx},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to get all cluster native local dogu registries")
			},
		},
		{
			name: "should return multi error on error parsing versions",
			configMapClientFn: func(t *testing.T) configMapClient {
				invalidCasCm := &corev1.ConfigMap{Data: map[string]string{"current": "abc"}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
				invalidLdapCm := &corev1.ConfigMap{Data: map[string]string{"current": "abcd"}, ObjectMeta: metav1.ObjectMeta{Labels: ldapVersionRegistryLabelMap}}
				invalidRegistryCmList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*invalidCasCm, *invalidLdapCm}}
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().List(testCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(invalidRegistryCmList, nil)

				return configMapClientMock
			},
			args: args{ctx: testCtx},
			want: []DoguVersion{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to get some dogu versions: failed to parse version \"abc\" for dogu \"cas\": failed to parse major version abc: strconv.Atoi: parsing \"abc\": invalid syntax\nfailed to parse version \"abcd\" for dogu \"ldap\": failed to parse major version abcd: strconv.Atoi: parsing \"abcd\": invalid syntax")
			},
		},
		{
			name: "should return multi error for invalid versions and dogu versions for valid versions",
			configMapClientFn: func(t *testing.T) configMapClient {
				invalidCasCm := &corev1.ConfigMap{Data: map[string]string{"current": "abc"}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
				validLdapCm := &corev1.ConfigMap{Data: map[string]string{"current": "1.0.0"}, ObjectMeta: metav1.ObjectMeta{Labels: ldapVersionRegistryLabelMap}}
				validRegistryCmList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*invalidCasCm, *validLdapCm}}
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().List(testCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(validRegistryCmList, nil)

				return configMapClientMock
			},
			args: args{ctx: testCtx},
			want: []DoguVersion{
				{
					Name:    "ldap",
					Version: parseVersionStr(t, "1.0.0"),
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to get some dogu versions: failed to parse version \"abc\" for dogu \"cas\": failed to parse major version abc: strconv.Atoi: parsing \"abc\": invalid syntax")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &doguVersionRegistry{
				configMapClient: tt.configMapClientFn(t),
			}
			got, err := vr.GetCurrentOfAll(tt.args.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("GetCurrentOfAll(%v)", tt.args.ctx)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetCurrentOfAll(%v)", tt.args.ctx)
		})
	}
}

func Test_versionRegistry_IsEnabled(t *testing.T) {
	casRegistryCmWithCurrent := &corev1.ConfigMap{Data: map[string]string{"current": casVersionStr}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
	casRegistryCmWithOutCurrent := &corev1.ConfigMap{Data: map[string]string{casVersionStr: readCasDoguStr(t)}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
	type args struct {
		ctx         context.Context
		doguVersion DoguVersion
	}
	tests := []struct {
		name              string
		configMapClientFn func(t *testing.T) configMapClient
		args              args
		want              bool
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "should return true if the current key matches",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCmWithCurrent, nil)

				return configMapClientMock
			},
			args:    args{ctx: testCtx, doguVersion: DoguVersion{"cas", parseVersionStr(t, casVersionStr)}},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "should return false if the current key does not exist",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCmWithOutCurrent, nil)

				return configMapClientMock
			},
			args:    args{ctx: testCtx, doguVersion: DoguVersion{"cas", parseVersionStr(t, casVersionStr)}},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "should return false if the current key does not match",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCmWithCurrent, nil)

				return configMapClientMock
			},
			args:    args{ctx: testCtx, doguVersion: DoguVersion{"cas", parseVersionStr(t, "7.0.5.1-2")}},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "should return error on error getting registry",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			args: args{ctx: testCtx, doguVersion: DoguVersion{"cas", parseVersionStr(t, casVersionStr)}},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to get dogu descriptor config map for dogu \"cas\"")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &doguVersionRegistry{
				configMapClient: tt.configMapClientFn(t),
			}
			got, err := vr.IsEnabled(tt.args.ctx, tt.args.doguVersion)
			if !tt.wantErr(t, err, fmt.Sprintf("IsEnabled(%v, %v)", tt.args.ctx, tt.args.doguVersion)) {
				return
			}
			assert.Equalf(t, tt.want, got, "IsEnabled(%v, %v)", tt.args.ctx, tt.args.doguVersion)
		})
	}
}

func Test_versionRegistry_Enable(t *testing.T) {
	casRegistryCmWithoutCurrent := &corev1.ConfigMap{Data: map[string]string{casVersionStr: readCasDoguStr(t)}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
	expectedCasRegistryCmWithoutCurrent := &corev1.ConfigMap{Data: map[string]string{casVersionStr: readCasDoguStr(t), "current": casVersionStr}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}

	type args struct {
		ctx         context.Context
		doguVersion DoguVersion
	}
	casArgs := args{
		ctx:         testCtx,
		doguVersion: DoguVersion{Name: "cas", Version: parseVersionStr(t, casVersionStr)},
	}
	tests := []struct {
		name              string
		configMapClientFn func(t *testing.T) configMapClient
		args              args
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "success with existent registry",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCmWithoutCurrent, nil)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCmWithoutCurrent, metav1.UpdateOptions{}).Return(expectedCasRegistryCmWithoutCurrent, nil)

				return configMapClientMock
			},
			args:    casArgs,
			wantErr: assert.NoError,
		},
		{
			name: "should success with conflict error on retry",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCmWithoutCurrent, nil).Times(2)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCmWithoutCurrent, metav1.UpdateOptions{}).Return(nil, testConflictErr).Times(1)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCmWithoutCurrent, metav1.UpdateOptions{}).Return(casRegistryCmWithoutCurrent, nil).Times(1)

				return configMapClientMock
			},
			args:    casArgs,
			wantErr: assert.NoError,
		},
		{
			name: "should return error on error getting registry",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			args: casArgs,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to enable dogu \"cas\" with version \"7.0.5.1-1\": failed to get dogu descriptor config map for dogu \"cas\"")
			},
		},
		{
			name: "should return error if the descriptor of the specified version is not found",
			configMapClientFn: func(t *testing.T) configMapClient {
				casRegistryCmWithOutSpec := &corev1.ConfigMap{Data: map[string]string{}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCmWithOutSpec, nil)

				return configMapClientMock
			},
			args: casArgs,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to enable dogu \"cas\" with version \"7.0.5.1-1\": dogu descriptor is not available")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &doguVersionRegistry{
				configMapClient: tt.configMapClientFn(t),
			}
			tt.wantErr(t, vr.Enable(tt.args.ctx, tt.args.doguVersion), fmt.Sprintf("Enable(%v, %v)", tt.args.ctx, tt.args.doguVersion))
		})
	}
}

func Test_versionRegistry_WatchAllCurrent(t *testing.T) {
	tests := []struct {
		name                string
		configMapClientFn   func(t *testing.T) configMapClient
		configMapInformerFn func(t *testing.T) configMapInformer
		expectFn            func(t *testing.T, watch <-chan CurrentVersionsWatchResult)
		wantErr             assert.ErrorAssertionFunc
	}{
		{
			name: "should return error on error getting initial dogu descriptor configmaps",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().List(mock.Anything, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			configMapInformerFn: func(t *testing.T) configMapInformer {
				return newMockConfigMapInformer(t)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i) && assert.True(t, cloudoguerrors.IsGenericError(err), "error is not a generic error", i) &&
					assert.ErrorContains(t, err, "failed to get initial state for watch: failed to get all cluster native local dogu registries", i)
			},
			expectFn: func(t *testing.T, watch <-chan CurrentVersionsWatchResult) {
				assert.Nil(t, watch)
			},
		},
		{
			name: "should return error on error creating initial persistence context because of invalid current versions",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				invalidCasCm := &corev1.ConfigMap{Data: map[string]string{"current": "abc"}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
				invalidRegistryList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*invalidCasCm}}
				configMapClientMock.EXPECT().List(mock.Anything, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(invalidRegistryList, nil)

				return configMapClientMock
			},
			configMapInformerFn: func(t *testing.T) configMapInformer {
				return newMockConfigMapInformer(t)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i) && assert.True(t, cloudoguerrors.IsGenericError(err), "error is not a generic error", i) &&
					assert.ErrorContains(t, err, "error during persistence context creation for watch: failed to parse version \"abc\" for dogu \"cas\"")
			},
			expectFn: func(t *testing.T, watch <-chan CurrentVersionsWatchResult) {
				assert.Nil(t, watch)
			},
		},
		{
			name: "should return error when failing to add event handler",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				casCm := &corev1.ConfigMap{Data: map[string]string{"current": "1.2.3-4"}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
				registryList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*casCm}}
				configMapClientMock.EXPECT().List(mock.Anything, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(registryList, nil)

				return configMapClientMock
			},
			configMapInformerFn: func(t *testing.T) configMapInformer {
				sInformer := newMockSharedInformer(t)
				sInformer.EXPECT().AddEventHandler(mock.Anything).Return(nil, assert.AnError)
				cmInformer := newMockConfigMapInformer(t)
				cmInformer.EXPECT().Informer().Return(sInformer)
				return cmInformer
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i) && assert.True(t, cloudoguerrors.IsGenericError(err), "error is not a generic error", i) &&
					assert.ErrorContains(t, err, "failed to start watch: failed to add event handler for current versions")
			},
			expectFn: func(t *testing.T, watch <-chan CurrentVersionsWatchResult) {
				assert.Nil(t, watch)
			},
		},
		{
			name: "should successfully run informer and close channel",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				casCm := &corev1.ConfigMap{Data: map[string]string{"current": "1.2.3-4"}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
				registryList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*casCm}}
				configMapClientMock.EXPECT().List(mock.Anything, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(registryList, nil)

				return configMapClientMock
			},
			configMapInformerFn: func(t *testing.T) configMapInformer {
				sInformer := newMockSharedInformer(t)
				sInformer.EXPECT().AddEventHandler(mock.Anything).Return(nil, nil)
				sInformer.EXPECT().Run(mock.Anything)
				cmInformer := newMockConfigMapInformer(t)
				cmInformer.EXPECT().Informer().Return(sInformer)
				return cmInformer
			},
			wantErr: assert.NoError,
			expectFn: func(t *testing.T, watch <-chan CurrentVersionsWatchResult) {
				assert.NotNil(t, watch)
				_, open := <-watch
				assert.False(t, open, "channel is still open but should be closed")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := &doguVersionRegistry{
				configMapClient:   tt.configMapClientFn(t),
				configMapInformer: tt.configMapInformerFn(t),
			}
			ctx, cancel := context.WithTimeout(testCtx, time.Second*5)
			defer cancel()
			got, err := vr.WatchAllCurrent(ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("WatchAllCurrent(%v)", ctx)) {
				return
			}

			if tt.expectFn != nil {
				tt.expectFn(t, got)
			}
		})
	}
}
