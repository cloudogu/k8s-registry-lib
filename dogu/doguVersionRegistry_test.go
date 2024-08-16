package dogu

import (
	"context"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	cloudoguerrors "github.com/cloudogu/k8s-registry-lib/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
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
					assert.ErrorContains(t, err, "failed to get dogu spec config map for dogu \"cas\"", i)
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
			name: "should return error on error getting all dogu spec configmaps",
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
			name: "should return true if the current key exists",
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
			name: "should return error on error getting registry",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			args: args{ctx: testCtx, doguVersion: DoguVersion{"cas", parseVersionStr(t, casVersionStr)}},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to get dogu spec config map for dogu \"cas\"")
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
	casRegistryCmWithOutCurrent := &corev1.ConfigMap{Data: map[string]string{casVersionStr: readCasDoguStr(t)}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
	expectedCasRegistryCmWithOutCurrent := &corev1.ConfigMap{Data: map[string]string{casVersionStr: readCasDoguStr(t), "current": casVersionStr}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}

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
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCmWithOutCurrent, nil)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCmWithOutCurrent, metav1.UpdateOptions{}).Return(expectedCasRegistryCmWithOutCurrent, nil)

				return configMapClientMock
			},
			args:    casArgs,
			wantErr: assert.NoError,
		},
		{
			name: "should success with conflict error on retry",
			configMapClientFn: func(t *testing.T) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCmWithOutCurrent, nil).Times(2)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCmWithOutCurrent, metav1.UpdateOptions{}).Return(nil, testConflictErr).Times(1)
				configMapClientMock.EXPECT().Update(testCtx, expectedCasRegistryCmWithOutCurrent, metav1.UpdateOptions{}).Return(casRegistryCmWithOutCurrent, nil).Times(1)

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
					assert.ErrorContains(t, err, "failed to enable dogu \"cas\" with version \"7.0.5.1-1\": failed to get dogu spec config map for dogu \"cas\"")
			},
		},
		{
			name: "should return error if the spec of the specified version is not found",
			configMapClientFn: func(t *testing.T) configMapClient {
				casRegistryCmWithOutSpec := &corev1.ConfigMap{Data: map[string]string{}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Get(testCtx, "dogu-spec-cas", metav1.GetOptions{}).Return(casRegistryCmWithOutSpec, nil)

				return configMapClientMock
			},
			args: casArgs,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err), i) &&
					assert.ErrorContains(t, err, "failed to enable dogu \"cas\" with version \"7.0.5.1-1\": dogu spec is not available")
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

type mockWatchInterface struct {
	channel chan watch.Event
}

func NewMockWatchInterface() *mockWatchInterface {
	channel := make(chan watch.Event)

	return &mockWatchInterface{
		channel: channel,
	}
}

func (mwi mockWatchInterface) Stop() {
}

func (mwi mockWatchInterface) ResultChan() <-chan watch.Event {
	return mwi.channel
}

func Test_versionRegistry_WatchAllCurrent(t *testing.T) {
	addCancelCtx, addCancelFunc := context.WithCancel(context.Background())
	emptyAddCancelCtx, emptyAddCancelFunc := context.WithCancel(context.Background())
	modifyCancelCtx, modifyCancelFunc := context.WithCancel(context.Background())
	deleteCancelCtx, deleteCancelFunc := context.WithCancel(context.Background())
	errorCancelCtx, errorCancelFunc := context.WithCancel(context.Background())
	ldapRegistryCm := &corev1.ConfigMap{Data: map[string]string{"current": ldapVersionStr}, ObjectMeta: metav1.ObjectMeta{Labels: ldapVersionRegistryLabelMap}}
	initialDoguVersionCtx := map[SimpleDoguName]core.Version{"ldap": parseVersionStr(t, ldapVersionStr)}
	casRegistryCm := &corev1.ConfigMap{Data: map[string]string{"current": casVersionStr}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
	registryCmList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*ldapRegistryCm}}
	emptyLdapRegistryCm := &corev1.ConfigMap{Data: map[string]string{}, ObjectMeta: metav1.ObjectMeta{Labels: ldapVersionRegistryLabelMap}}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name              string
		configMapClientFn func(t *testing.T, watchInterface *mockWatchInterface) configMapClient
		args              args
		eventMockFn       func(watchInterface *mockWatchInterface)
		expectFn          func(t *testing.T, watch CurrentVersionsWatch)
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name: "should return error on watch error",
			configMapClientFn: func(t *testing.T, watchInterface *mockWatchInterface) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Watch(context.Background(), metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			args: args{context.Background()},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.True(t, cloudoguerrors.IsGenericError(err)) &&
					assert.ErrorContains(t, err, "failed to create watches for selector \"app=ces,dogu.name,k8s.cloudogu.com/type=local-dogu-registry\"")
			},
			expectFn: func(t *testing.T, watch CurrentVersionsWatch) {},
		},
		{
			name: "should return error on error getting initial dogu spec configmaps",
			configMapClientFn: func(t *testing.T, watchInterface *mockWatchInterface) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Watch(testCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(watchInterface, nil)
				configMapClientMock.EXPECT().List(testCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(nil, assert.AnError)

				return configMapClientMock
			},
			args:    args{testCtx},
			wantErr: assert.NoError,
			expectFn: func(t *testing.T, watch CurrentVersionsWatch) {
				channel := watch.ResultChan

				result := <-channel
				require.True(t, cloudoguerrors.IsGenericError(result.Err))
				assert.ErrorContains(t, result.Err, "failed to get all cluster native local dogu registries")
			},
		},
		{
			name: "should return error on error creating initial persistence context because of invalid current versions",
			configMapClientFn: func(t *testing.T, watchInterface *mockWatchInterface) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				invalidCasCm := &corev1.ConfigMap{Data: map[string]string{"current": "abc"}, ObjectMeta: metav1.ObjectMeta{Labels: casVersionRegistryLabelMap}}
				invalidRegistryList := &corev1.ConfigMapList{Items: []corev1.ConfigMap{*invalidCasCm}}
				configMapClientMock.EXPECT().Watch(testCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(watchInterface, nil)
				configMapClientMock.EXPECT().List(testCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(invalidRegistryList, nil)

				return configMapClientMock
			},
			args:    args{testCtx},
			wantErr: assert.NoError,
			expectFn: func(t *testing.T, watch CurrentVersionsWatch) {
				channel := watch.ResultChan

				result := <-channel
				require.True(t, cloudoguerrors.IsGenericError(result.Err))
				assert.ErrorContains(t, result.Err, "failed to create persistent context: failed to parse version \"abc\" for dogu \"cas\"")
			},
		},
		{
			name: "should throw event with dogu version objects on add event",
			configMapClientFn: func(t *testing.T, watchInterface *mockWatchInterface) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Watch(addCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(watchInterface, nil)
				configMapClientMock.EXPECT().List(addCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(registryCmList, nil)

				return configMapClientMock
			},
			args: args{ctx: addCancelCtx},
			eventMockFn: func(watchInterface *mockWatchInterface) {
				event := watch.Event{
					Type:   watch.Added,
					Object: casRegistryCm,
				}

				watchInterface.channel <- event
			},
			expectFn: func(t *testing.T, watch CurrentVersionsWatch) {
				channel := watch.ResultChan

				result := <-channel
				require.NoError(t, result.Err)
				assert.Equal(t, initialDoguVersionCtx, result.PrevVersions)
				casVersion := parseVersionStr(t, casVersionStr)
				assert.Equal(t, result.Versions, map[SimpleDoguName]core.Version{"ldap": parseVersionStr(t, ldapVersionStr), "cas": casVersion})
				assert.Equal(t, []DoguVersion{{Name: "cas", Version: casVersion}}, result.Diff)

				addCancelFunc()
			},
			wantErr: assert.NoError,
		},
		{
			name: "should throw no event with add event without current key",
			configMapClientFn: func(t *testing.T, watchInterface *mockWatchInterface) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Watch(emptyAddCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(watchInterface, nil)
				configMapClientMock.EXPECT().List(emptyAddCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(registryCmList, nil)

				return configMapClientMock
			},
			args: args{ctx: emptyAddCancelCtx},
			eventMockFn: func(watchInterface *mockWatchInterface) {
				event := watch.Event{
					Type:   watch.Added,
					Object: emptyLdapRegistryCm,
				}

				watchInterface.channel <- event

				// We have to send two events because is not possible to check if no event is thrown.
				event = watch.Event{
					Type:   watch.Added,
					Object: casRegistryCm,
				}

				watchInterface.channel <- event
			},
			expectFn: func(t *testing.T, watch CurrentVersionsWatch) {
				channel := watch.ResultChan

				result := <-channel
				require.NoError(t, result.Err)
				assert.Equal(t, initialDoguVersionCtx, result.PrevVersions)
				casVersion := parseVersionStr(t, casVersionStr)
				assert.Equal(t, result.Versions, map[SimpleDoguName]core.Version{"ldap": parseVersionStr(t, ldapVersionStr), "cas": casVersion})
				assert.Equal(t, []DoguVersion{{Name: "cas", Version: casVersion}}, result.Diff)

				emptyAddCancelFunc()
			},
			wantErr: assert.NoError,
		},
		{
			name: "should throw event with dogu version objects on modified event",
			configMapClientFn: func(t *testing.T, watchInterface *mockWatchInterface) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Watch(modifyCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(watchInterface, nil)
				configMapClientMock.EXPECT().List(modifyCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(registryCmList, nil)

				return configMapClientMock
			},
			args: args{ctx: modifyCancelCtx},
			eventMockFn: func(watchInterface *mockWatchInterface) {
				event := watch.Event{
					Type:   watch.Modified,
					Object: &corev1.ConfigMap{Data: map[string]string{"current": upgradeLdapVersionStr}, ObjectMeta: metav1.ObjectMeta{Labels: ldapVersionRegistryLabelMap}},
				}

				watchInterface.channel <- event
			},
			expectFn: func(t *testing.T, watch CurrentVersionsWatch) {
				channel := watch.ResultChan

				result := <-channel
				require.NoError(t, result.Err)
				assert.Equal(t, initialDoguVersionCtx, result.PrevVersions)
				upgradedLdapVersion := parseVersionStr(t, upgradeLdapVersionStr)
				assert.Equal(t, result.Versions, map[SimpleDoguName]core.Version{"ldap": upgradedLdapVersion})
				assert.Equal(t, []DoguVersion{{Name: "ldap", Version: upgradedLdapVersion}}, result.Diff)

				modifyCancelFunc()
			},
			wantErr: assert.NoError,
		},
		{
			name: "should throw event with dogu version objects on delete event",
			configMapClientFn: func(t *testing.T, watchInterface *mockWatchInterface) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Watch(deleteCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(watchInterface, nil)
				configMapClientMock.EXPECT().List(deleteCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(registryCmList, nil)

				return configMapClientMock
			},
			args: args{ctx: deleteCancelCtx},
			eventMockFn: func(watchInterface *mockWatchInterface) {
				event := watch.Event{
					Type:   watch.Deleted,
					Object: &corev1.ConfigMap{Data: map[string]string{"current": ldapVersionStr}, ObjectMeta: metav1.ObjectMeta{Labels: ldapVersionRegistryLabelMap}},
				}

				watchInterface.channel <- event
			},
			expectFn: func(t *testing.T, watch CurrentVersionsWatch) {
				channel := watch.ResultChan

				result := <-channel
				require.NoError(t, result.Err)
				assert.Equal(t, initialDoguVersionCtx, result.PrevVersions)
				ldapVersion := parseVersionStr(t, ldapVersionStr)
				assert.Equal(t, result.Versions, map[SimpleDoguName]core.Version{})
				assert.Equal(t, []DoguVersion{{Name: "ldap", Version: ldapVersion}}, result.Diff)

				deleteCancelFunc()
			},
			wantErr: assert.NoError,
		},
		{
			name: "should return error on error event",
			configMapClientFn: func(t *testing.T, watchInterface *mockWatchInterface) configMapClient {
				configMapClientMock := newMockConfigMapClient(t)
				configMapClientMock.EXPECT().Watch(errorCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(watchInterface, nil)
				configMapClientMock.EXPECT().List(errorCancelCtx, metav1.ListOptions{LabelSelector: versionRegistryLabelSelector}).Return(registryCmList, nil)

				return configMapClientMock
			},
			args: args{ctx: errorCancelCtx},
			eventMockFn: func(watchInterface *mockWatchInterface) {
				event := watch.Event{
					Type:   watch.Error,
					Object: &metav1.Status{Status: "123", Message: "message"},
				}

				watchInterface.channel <- event
				errorCancelFunc()
			},
			expectFn: func(t *testing.T, watch CurrentVersionsWatch) {
				channel := watch.ResultChan

				result := <-channel
				require.True(t, cloudoguerrors.IsGenericError(result.Err))
				assert.ErrorContains(t, result.Err, "watch event type is error: \"&Status{ListMeta:ListMeta{SelfLink:,ResourceVersion:,Continue:,RemainingItemCount:nil,},Status:123,Message:message,Reason:,Details:nil,Code:0,}\"")
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watchInterface := NewMockWatchInterface()

			vr := &doguVersionRegistry{
				configMapClient: tt.configMapClientFn(t, watchInterface),
			}
			got, err := vr.WatchAllCurrent(tt.args.ctx)
			if !tt.wantErr(t, err, fmt.Sprintf("WatchAllCurrent(%v)", tt.args.ctx)) {
				return
			}

			if tt.eventMockFn != nil {
				go func() {
					timer := time.NewTimer(time.Second)
					<-timer.C
					tt.eventMockFn(watchInterface)
				}()
			}

			if tt.expectFn != nil {
				tt.expectFn(t, got)
			}
		})
	}
}
