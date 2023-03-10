package main

import (
	"context"
	"os"
	"sync"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/sdk/plugin"
)

type backend struct {
	framework.Backend
	initOnce                sync.Once
	thingWeNeedToInitialize int
}

func (b *backend) init() {
	// Initialize here
	b.thingWeNeedToInitialize = 42
}

func (b *backend) doSomething(ctx context.Context, request *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// Initialization actually triggered here
	b.initOnce.Do(b.init)

	return &logical.Response{
		Data: map[string]interface{}{
			"hello":                   "world",
			"thingWeNeedToInitialize": b.thingWeNeedToInitialize,
		},
	}, nil
}

func factory(context.Context, *logical.BackendConfig) (logical.Backend, error) {
	var b backend
	b.Backend = framework.Backend{
		Paths: []*framework.Path{
			{
				Pattern: "something",
				Operations: map[logical.Operation]framework.OperationHandler{
					logical.ReadOperation: &framework.PathOperation{
						Callback: b.doSomething,
					},
				},
			},
		},
		BackendType: logical.TypeLogical,
	}

	// Don't initialize more than necessary here

	return &b, nil
}

func main() {
	var apiClientMeta api.PluginAPIClientMeta

	err := apiClientMeta.FlagSet().Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	err = plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: factory,
		TLSProviderFunc:    api.VaultPluginTLSProvider(apiClientMeta.GetTLSConfig()),
	})
	if err != nil {
		panic(err)
	}
}
