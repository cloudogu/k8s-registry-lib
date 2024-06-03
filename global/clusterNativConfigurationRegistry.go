package global

import "context"

type clusterNativeConfigRegistry struct {
	prefix string
}

func (c clusterNativeConfigRegistry) Set(ctx context.Context, key, value string) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) SetWithLifetime(ctx context.Context, key, value string, timeToLiveInSeconds int) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) Refresh(ctx context.Context, key string, timeToLiveInSeconds int) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) Get(ctx context.Context, key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) GetAll(ctx context.Context) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) Delete(ctx context.Context, key string) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) DeleteRecursive(ctx context.Context, key string) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) Exists(ctx context.Context, key string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) RemoveAll(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) GetOrFalse(ctx context.Context, key string) (bool, string, error) {
	//TODO implement me
	panic("implement me")
}
