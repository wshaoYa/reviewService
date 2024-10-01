package server

import (
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
	"reviewService/internal/conf"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar)

// NewRegistrar 服务注册
func NewRegistrar(c *conf.Consul) registry.Registrar {
	consulCfg := api.DefaultConfig()
	consulCfg.Address = c.GetAddress()
	consulCfg.Scheme = c.GetScheme()

	cli, err := api.NewClient(consulCfg)
	if err != nil {
		panic(err)
	}

	return consul.New(cli)
}
