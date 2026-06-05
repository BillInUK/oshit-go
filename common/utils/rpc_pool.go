package utils

import (
	"fmt"
	"sync/atomic"

	"github.com/gagliardetto/solana-go/rpc"
)

// RPCEndpointConfig holds config for a single RPC endpoint.
type RPCEndpointConfig struct {
	Provider    string
	Endpoint    string
	APIKey      string
	WssEndpoint string
	WssAPIKey   string
	Weight      int
}

// BuildRPCURL constructs the full RPC URL based on provider conventions.
func BuildRPCURL(provider, endpoint, apiKey string) string {
	if apiKey == "" {
		return endpoint
	}
	switch provider {
	case "quicknode":
		return endpoint + "/" + apiKey
	case "helius":
		return endpoint + "/?api-key=" + apiKey
	default:
		return endpoint
	}
}

// BuildWssURL constructs the full WSS URL based on provider conventions.
func BuildWssURL(provider, endpoint, apiKey string) string {
	if apiKey == "" {
		return endpoint
	}
	switch provider {
	case "quicknode":
		return endpoint + "/" + apiKey
	default:
		return endpoint
	}
}

// RPCPool provides round-robin access to multiple Solana RPC clients.
type RPCPool struct {
	clients []*rpc.Client
	urls    []string
	index   atomic.Uint64
}

// NewRPCPool creates a pool from endpoint configs. Skips entries with weight <= 0.
func NewRPCPool(endpoints []RPCEndpointConfig) (*RPCPool, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no RPC endpoints configured")
	}
	pool := &RPCPool{}
	for _, ep := range endpoints {
		if ep.Weight <= 0 {
			continue
		}
		url := BuildRPCURL(ep.Provider, ep.Endpoint, ep.APIKey)
		pool.urls = append(pool.urls, url)
		pool.clients = append(pool.clients, rpc.New(url))
	}
	if len(pool.clients) == 0 {
		return nil, fmt.Errorf("all RPC endpoints are disabled (weight=0)")
	}
	return pool, nil
}

// Next returns the next RPC client in round-robin order.
func (p *RPCPool) Next() *rpc.Client {
	i := p.index.Add(1) - 1
	return p.clients[i%uint64(len(p.clients))]
}

// First returns the first RPC client (for backward compatibility).
func (p *RPCPool) First() *rpc.Client {
	return p.clients[0]
}

// FirstURL returns the first RPC URL.
func (p *RPCPool) FirstURL() string {
	return p.urls[0]
}

// Len returns the number of active endpoints.
func (p *RPCPool) Len() int {
	return len(p.clients)
}
