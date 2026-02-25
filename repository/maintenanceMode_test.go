package repository

import (
	"context"
	"testing"

	"github.com/cloudogu/ces-commons-lib/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testNamespace = "test-namespace"

var testCtx = context.Background()

func TestNewMaintenanceModeAdapter(t *testing.T) {
	adapter := NewMaintenanceModeAdapter("k8s-service-discovery", nil, testNamespace)
	assert.NotEmpty(t, adapter)
}

func TestMaintenanceModeAdapter_GetStatus(t *testing.T) {
	tests := []struct {
		name     string
		clientFn func(t *testing.T) k8sClient
		want1    MaintenanceModeDescription
		want2    bool
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "fail to get maintenance config",
			clientFn: func(t *testing.T) k8sClient {
				mck := newMockK8sClient(t)
				mck.EXPECT().Get(testCtx, types.NamespacedName{Name: "maintenance", Namespace: testNamespace}, &corev1.ConfigMap{}).
					Return(assert.AnError)
				return mck
			},
			want1: MaintenanceModeDescription{},
			want2: false,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, assert.AnError, i) &&
					errors.IsGenericError(err) &&
					assert.ErrorContains(t, err, "failed to get config for maintenance mode")
			},
		},
		{
			name: "succeed with false if not found",
			clientFn: func(t *testing.T) k8sClient {
				return fake.NewClientBuilder().Build()
			},
			want1:   MaintenanceModeDescription{},
			want2:   false,
			wantErr: assert.NoError,
		},
		{
			name: "succeed with active",
			clientFn: func(t *testing.T) k8sClient {
				config := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespace,
						Name:      "maintenance",
					},
					Data: map[string]string{
						"active": "TRuE",
						"title":  "Backup",
						"text":   "Backup in progress",
					},
				}
				return fake.NewClientBuilder().WithObjects(config).Build()
			},
			want1: MaintenanceModeDescription{
				Title: "Backup",
				Text:  "Backup in progress",
			},
			want2:   true,
			wantErr: assert.NoError,
		},
		{
			name: "succeed with inactive",
			clientFn: func(t *testing.T) k8sClient {
				config := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespace,
						Name:      "maintenance",
					},
					Data: map[string]string{
						"active": "notTRuE",
						"title":  "Backup",
						"text":   "Backup in progress",
					},
				}
				return fake.NewClientBuilder().WithObjects(config).Build()
			},
			want1:   MaintenanceModeDescription{},
			want2:   false,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mma := &MaintenanceModeAdapter{
				client:    tt.clientFn(t),
				namespace: testNamespace,
			}
			got1, got2, err := mma.GetStatus(testCtx)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want1, got1)
			assert.Equal(t, tt.want2, got2)
		})
	}
}

func TestMaintenanceModeAdapter_Activate(t *testing.T) {
	tests := []struct {
		name     string
		clientFn func(t *testing.T) k8sClient
		content  MaintenanceModeDescription
		force    bool
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "fail to get maintenance config-map",
			clientFn: func(t *testing.T) k8sClient {
				mck := newMockK8sClient(t)
				mck.EXPECT().Get(testCtx, types.NamespacedName{Name: "maintenance", Namespace: testNamespace}, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      MaintenanceConfigMapName,
						Namespace: testNamespace,
					},
				}).
					Return(assert.AnError)
				return mck
			},
			content: MaintenanceModeDescription{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, assert.AnError, i) &&
					errors.IsGenericError(err) &&
					assert.ErrorContains(t, err, "could not maintenance config-map")
			},
		},
		{
			name: "fail with conflict",
			clientFn: func(t *testing.T) k8sClient {
				config := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespace,
						Name:      "maintenance",
					},
					Data: map[string]string{
						"active": "true",
						"holder": "k8s-ces-control",
					},
				}
				return fake.NewClientBuilder().WithObjects(config).Build()
			},
			content: MaintenanceModeDescription{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return errors.IsConflictError(err) &&
					assert.ErrorContains(t, err, "maintenance mode is already activated by another owner: k8s-ces-control")
			},
		},
		{
			name: "no conflict with force",
			clientFn: func(t *testing.T) k8sClient {
				config := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespace,
						Name:      "maintenance",
					},
					Data: map[string]string{
						"active": "true",
						"holder": "k8s-ces-control",
					},
				}
				return fake.NewClientBuilder().WithObjects(config).Build()
			},
			content: MaintenanceModeDescription{
				Title: "Backup",
				Text:  "Backup in progress",
			},
			force:   true,
			wantErr: assert.NoError,
		},
		{
			name: "fail to update",
			clientFn: func(t *testing.T) k8sClient {
				mck := newMockK8sClient(t)
				mck.EXPECT().Get(testCtx, types.NamespacedName{Name: "maintenance", Namespace: testNamespace}, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      MaintenanceConfigMapName,
						Namespace: testNamespace,
					},
				}).
					Run(func(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) {
						obj.(*corev1.ConfigMap).Data = map[string]string{
							"active": "true",
							"holder": "k8s-service-discovery",
							"title":  "Restore",
							"text":   "A backup is currently being restored",
						}
					}).
					Return(nil)
				mck.EXPECT().Update(testCtx, mock.AnythingOfType("*v1.ConfigMap")).
					Return(assert.AnError)
				return mck
			},
			content: MaintenanceModeDescription{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, assert.AnError, i) &&
					errors.IsGenericError(err) &&
					assert.ErrorContains(t, err, "could not update maintenance config-map")
			},
		},
		{
			name: "succeed to update",
			clientFn: func(t *testing.T) k8sClient {
				config := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespace,
						Name:      "maintenance",
					},
					Data: map[string]string{
						"active": "true",
						"holder": "k8s-service-discovery",
						"title":  "Backup",
						"text":   "Backup in progress",
					},
				}
				return fake.NewClientBuilder().WithObjects(config).Build()
			},
			content: MaintenanceModeDescription{
				Title: "Backup",
				Text:  "Backup in progress",
			},
			wantErr: assert.NoError,
		},
		{
			name: "fail to create",
			clientFn: func(t *testing.T) k8sClient {
				mck := newMockK8sClient(t)
				mck.EXPECT().Get(testCtx, types.NamespacedName{Name: "maintenance", Namespace: testNamespace}, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      MaintenanceConfigMapName,
						Namespace: testNamespace,
					},
				}).
					Return(&apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}})
				mck.EXPECT().Create(testCtx, mock.AnythingOfType("*v1.ConfigMap")).
					Return(assert.AnError)
				return mck
			},
			content: MaintenanceModeDescription{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, assert.AnError, i) &&
					errors.IsGenericError(err) &&
					assert.ErrorContains(t, err, "could not create maintenance config-map")
			},
		},
		{
			name: "succeed to create",
			clientFn: func(t *testing.T) k8sClient {
				return fake.NewClientBuilder().Build()
			},
			content: MaintenanceModeDescription{
				Title: "Backup",
				Text:  "Backup in progress",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mma := &MaintenanceModeAdapter{
				owner:     "k8s-service-discovery",
				client:    tt.clientFn(t),
				namespace: testNamespace,
			}
			tt.wantErr(t, mma.Activate(testCtx, tt.content, tt.force))
		})
	}
}

func TestMaintenanceModeAdapter_Deactivate(t *testing.T) {
	tests := []struct {
		name     string
		force    bool
		clientFn func(t *testing.T) k8sClient
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "fail to get maintenance config-map",
			clientFn: func(t *testing.T) k8sClient {
				mck := newMockK8sClient(t)
				mck.EXPECT().Get(testCtx, types.NamespacedName{Name: "maintenance", Namespace: testNamespace}, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      MaintenanceConfigMapName,
						Namespace: testNamespace,
					},
				}).
					Return(assert.AnError)
				return mck
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, assert.AnError, i) &&
					errors.IsGenericError(err) &&
					assert.ErrorContains(t, err, "could not maintenance config-map")
			},
		},
		{
			name: "fail with conflict",
			clientFn: func(t *testing.T) k8sClient {
				config := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespace,
						Name:      "maintenance",
					},
					Data: map[string]string{
						"active": "true",
						"holder": "k8s-ces-control",
					},
				}
				return fake.NewClientBuilder().WithObjects(config).Build()
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return errors.IsConflictError(err) &&
					assert.ErrorContains(t, err, "maintenance mode is already activated by another owner: k8s-ces-control")
			},
		},
		{
			name:  "no conflict with force",
			force: true,
			clientFn: func(t *testing.T) k8sClient {
				config := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespace,
						Name:      "maintenance",
					},
					Data: map[string]string{
						"active": "true",
						"holder": "k8s-ces-control",
					},
				}
				return fake.NewClientBuilder().WithObjects(config).Build()
			},
			wantErr: assert.NoError,
		},
		{
			name: "fail to update",
			clientFn: func(t *testing.T) k8sClient {
				mck := newMockK8sClient(t)
				mck.EXPECT().Get(testCtx, types.NamespacedName{Name: "maintenance", Namespace: testNamespace}, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      MaintenanceConfigMapName,
						Namespace: testNamespace,
					},
				}).
					Run(func(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) {
						obj.(*corev1.ConfigMap).Data = map[string]string{
							"active": "true",
							"holder": "k8s-service-discovery",
							"title":  "Restore",
							"text":   "A backup is currently being restored",
						}
					}).
					Return(nil)
				mck.EXPECT().Update(testCtx, mock.AnythingOfType("*v1.ConfigMap")).
					Return(assert.AnError)
				return mck
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, assert.AnError, i) &&
					errors.IsGenericError(err) &&
					assert.ErrorContains(t, err, "could not update maintenance config-map")
			},
		},
		{
			name: "succeed to update",
			clientFn: func(t *testing.T) k8sClient {
				config := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespace,
						Name:      "maintenance",
					},
					Data: map[string]string{
						"active": "true",
						"holder": "k8s-service-discovery",
						"title":  "Backup",
						"text":   "Backup in progress",
					},
				}
				return fake.NewClientBuilder().WithObjects(config).Build()
			},
			wantErr: assert.NoError,
		},
		{
			name: "fail to create",
			clientFn: func(t *testing.T) k8sClient {
				mck := newMockK8sClient(t)
				mck.EXPECT().Get(testCtx, types.NamespacedName{Name: "maintenance", Namespace: testNamespace}, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      MaintenanceConfigMapName,
						Namespace: testNamespace,
					},
				}).
					Return(&apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}})
				mck.EXPECT().Create(testCtx, mock.AnythingOfType("*v1.ConfigMap")).
					Return(assert.AnError)
				return mck
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, assert.AnError, i) &&
					errors.IsGenericError(err) &&
					assert.ErrorContains(t, err, "could not create maintenance config-map")
			},
		},
		{
			name: "succeed to create",
			clientFn: func(t *testing.T) k8sClient {
				return fake.NewClientBuilder().Build()
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mma := &MaintenanceModeAdapter{
				owner:     "k8s-service-discovery",
				client:    tt.clientFn(t),
				namespace: testNamespace,
			}
			tt.wantErr(t, mma.Deactivate(testCtx, tt.force))
		})
	}
}
