package store

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type ObjectMeta struct {
	ID  string
	Key string
}

type DatabaseStats struct {
	ObjectCount      int
	ContentTypeStats map[string]int
}

type Database struct {
	mu         sync.RWMutex
	keyStore   map[string]*ObjectMeta
	idStore    map[string]*ObjectMeta
	groupStore map[string][]*ObjectMeta
}

func NewDatabse() *Database {
	return &Database{
		mu:         sync.RWMutex{},
		keyStore:   make(map[string]*ObjectMeta),
		idStore:    make(map[string]*ObjectMeta),
		groupStore: make(map[string][]*ObjectMeta),
	}
}
func (d *Database) PutBatchObjectMeta(ctx context.Context, objs []minio.ObjectInfo) (ids []string, err error) {
	newthings := []minio.ObjectInfo{}
	d.mu.RLock()
	for _, obj := range objs {
		v, ok := d.keyStore[obj.Key]
		if ok {
			ids = append(ids, v.ID)
		} else {
			newthings = append(newthings, obj)
		}
	}
	d.mu.RUnlock()

	d.mu.Lock()
	for _, obj := range newthings {
		id := uuid.New().String()
		ob := ObjectMeta{
			Key: obj.Key,
			ID:  id,
		}
		d.idStore[id] = &ob
		d.keyStore[obj.Key] = &ob
		group := path.Dir(obj.Key)
		d.groupStore[group] = append(d.groupStore[group], &ob)
		ids = append(ids, id)
	}
	d.mu.Unlock()
	return
}
func (d *Database) PutObjectMeta(ctx context.Context, obj minio.ObjectInfo) (id string, err error) {
	// Read Lock
	d.mu.RLock()
	v, ok := d.keyStore[obj.Key]
	d.mu.RUnlock()
	if ok {
		return v.ID, nil
	} else {
		id = uuid.New().String()
		ob := ObjectMeta{
			Key: obj.Key,
			ID:  id,
		}

		// Write Lock
		d.mu.Lock()
		d.idStore[id] = &ob
		d.keyStore[obj.Key] = &ob
		group := path.Dir(obj.Key)
		d.groupStore[group] = append(d.groupStore[group], &ob)
		d.mu.Unlock()
		return id, nil
	}
}

func (d *Database) ListObjectMetaByGroup(ctx context.Context, group string) (obj []*ObjectMeta, err error) {
	d.mu.RLock()
	v, ok := d.groupStore[group]
	d.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("group not found")
	}
	return v, nil
}

func (d *Database) GetObjectMeta(ctx context.Context, id string) (obj *ObjectMeta, err error) {
	d.mu.RLock()
	v, ok := d.idStore[id]
	d.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("object not found")
	}
	return v, nil
}

func (d *Database) SampleObject(ctx context.Context, count int) (obj []*ObjectMeta, err error) {
	d.mu.RLock()
	for _, v := range d.idStore {
		count -= 1
		if count < 0 {
			break
		}
		obj = append(obj, v)
	}
	d.mu.RUnlock()
	return
}

func (d *Database) Stats(ctx context.Context) (DatabaseStats, error) {
	stats := make(map[string]int)
	d.mu.RLock()
	defer d.mu.RUnlock()
	for k := range d.keyStore {
		stats[path.Ext(k)] += 1
	}
	return DatabaseStats{
		ObjectCount:      len(d.idStore),
		ContentTypeStats: stats,
	}, nil
}
