package global

type clusterNativeConfigRegistry struct {
	prefix string
}

func (c clusterNativeConfigRegistry) Set(key, value string) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) SetWithLifetime(key, value string, timeToLiveInSeconds int) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) Refresh(key string, timeToLiveInSeconds int) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) Get(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) GetAll() (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) Delete(key string) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) DeleteRecursive(key string) error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) Exists(key string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) RemoveAll() error {
	//TODO implement me
	panic("implement me")
}

func (c clusterNativeConfigRegistry) GetOrFalse(key string) (bool, string, error) {
	//TODO implement me
	panic("implement me")
}
