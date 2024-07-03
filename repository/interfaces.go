package repository

import (
	"context"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type generalConfigRepository interface {
	get(context.Context, configName) (config.Config, error)
	delete(context.Context, configName) error
	create(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)
	update(context.Context, configName, config.SimpleDoguName, config.Config) (config.Config, error)
	saveOrMerge(context.Context, configName, config.Config) (config.Config, error)
}
