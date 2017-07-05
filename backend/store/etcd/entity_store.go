package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	entityPathPrefix = "entities"
)

func getEntityPath(entity *types.Entity) string {
	return path.Join(etcdRoot, entityPathPrefix, entity.Organization, entity.ID)
}

func getEntitiesPath(ctx context.Context, id string) string {
	var org string

	// Determine the organization
	if value := ctx.Value(types.OrganizationKey); value != nil {
		org = value.(string)
	} else {
		org = ""
	}

	return path.Join(etcdRoot, entityPathPrefix, org, id)
}

func (s *etcdStore) DeleteEntity(e *types.Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}
	_, err := s.kvc.Delete(context.TODO(), getEntityPath(e))
	return err
}

func (s *etcdStore) DeleteEntityByID(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("must specify id")
	}

	_, err := s.kvc.Delete(context.TODO(), getEntitiesPath(ctx, id))
	return err
}

func (s *etcdStore) GetEntityByID(ctx context.Context, id string) (*types.Entity, error) {
	if id == "" {
		return nil, errors.New("must specify id")
	}

	resp, err := s.kvc.Get(context.TODO(), getEntitiesPath(ctx, id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, nil
	}
	entity := &types.Entity{}
	err = json.Unmarshal(resp.Kvs[0].Value, entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// GetEntities takes an optional org argument, an empty string will return
// all entities.
func (s *etcdStore) GetEntities(ctx context.Context) ([]*types.Entity, error) {
	resp, err := s.kvc.Get(context.TODO(), getEntitiesPath(ctx, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.Entity{}, nil
	}

	earr := make([]*types.Entity, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		entity := &types.Entity{}
		err = json.Unmarshal(kv.Value, entity)
		if err != nil {
			return nil, err
		}
		earr[i] = entity
	}

	return earr, nil
}

func (s *etcdStore) UpdateEntity(e *types.Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}

	eStr, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.kvc.Put(context.TODO(), getEntityPath(e), string(eStr))
	return err
}
