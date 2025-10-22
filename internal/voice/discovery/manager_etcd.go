package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdManager struct {
	cli    *clientv3.Client
	prefix string
	ttlS   int64
}

func NewEtcdManager(endpoints []string, prefix, username, password string) (*EtcdManager, error) {
	if prefix == "" {
		prefix = "/gochat/sfu"
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints, Username: username, Password: password})
	if err != nil {
		return nil, err
	}
	return &EtcdManager{cli: cli, prefix: prefix, ttlS: 15}, nil
}

func (m *EtcdManager) Register(ctx context.Context, region string, inst Instance) error {
	key := fmt.Sprintf("%s/%s/%s", m.prefix, region, inst.ID)
	val, _ := json.Marshal(inst)
	// create a short lease and attach it to the key
	lease, err := m.cli.Grant(ctx, m.ttlS)
	if err != nil {
		return err
	}
	_, err = m.cli.Put(ctx, key, string(val), clientv3.WithLease(lease.ID))
	return err
}

func (m *EtcdManager) List(ctx context.Context, region string) ([]Instance, error) {
	key := fmt.Sprintf("%s/%s/", m.prefix, region)
	resp, err := m.cli.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	out := make([]Instance, 0, len(resp.Kvs))
	now := time.Now().Unix()
	for _, kv := range resp.Kvs {
		var inst Instance
		if err := json.Unmarshal(kv.Value, &inst); err == nil {
			if now-inst.UpdatedAt <= 60 {
				out = append(out, inst)
			}
		}
	}
	return out, nil
}

// NewManager returns an etcd-backed manager when built with 'etcd' tag.
func NewManager(endpoints []string, prefix, username, password string) (Manager, error) {
	m, err := NewEtcdManager(endpoints, prefix, username, password)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Regions lists unique regions under the prefix.
func (m *EtcdManager) Regions(ctx context.Context) ([]string, error) {
	if m == nil || m.cli == nil {
		return nil, nil
	}
	// List all keys under prefix and extract the region segment: <prefix>/<region>/<id>
	resp, err := m.cli.Get(ctx, m.prefix+"/", clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return nil, err
	}
	uniq := make(map[string]struct{})
	for _, kv := range resp.Kvs {
		// kv.Key format: /prefix/region/id
		key := string(kv.Key)
		rest := key[len(m.prefix)+1:]
		if i := len(rest); i > 0 {
			// region is up to next '/'
			for j := 0; j < len(rest); j++ {
				if rest[j] == '/' {
					r := rest[:j]
					if r != "" {
						uniq[r] = struct{}{}
					}
					break
				}
			}
		}
	}
	out := make([]string, 0, len(uniq))
	for r := range uniq {
		out = append(out, r)
	}
	return out, nil
}
