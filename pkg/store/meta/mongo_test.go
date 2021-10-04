package meta_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/iwanhae/random-image/pkg/config"
	"github.com/iwanhae/random-image/pkg/store/meta"
	"github.com/stretchr/testify/assert"
)

var (
	ValidObjectMetas = []meta.ObjectMeta{
		{Path: "test/1", S3Key: "some/key1.jpg"},
		{Path: "test/2", S3Key: "some/key2.jpeg"},
		{Path: "test/3", S3Key: "some/key3.svg"},
		{Path: "test/4", S3Key: "some/key4.jpg"},
		{Path: "test/5", S3Key: "some/key5.jpg"},
		{Path: "test/6", S3Key: "some/key6.jpg"},
		{Path: "test/7", S3Key: "some/key7.jpg"},
		{Path: "test/8", S3Key: "some/key8.jpg"},
		{Path: "test/9", S3Key: "some/key9.jpg"},
		{Path: "test/10", S3Key: "some/key10.jpg"},
		{Path: "test2/10", S3Key: "some/key11.jpg"},
	}
	InvalidObjestMetas = []meta.ObjectMeta{
		{Path: "", S3Key: "invalid.key"},
		{Path: "/invalid", S3Key: ""},
		{Path: "/invalid", S3Key: "invalid.key"},
	}
)

func TestMongoDatabase(t *testing.T) {
	c := config.Init()
	ctx := context.Background()
	db, err := meta.InitMongoDB(ctx, c)
	if !assert.NoError(t, err, "fail to init mongodb") {
		return
	}

	////////////////
	// Test Invalid Create
	for _, v := range InvalidObjestMetas {
		_, err := db.Create(ctx, v)
		if err == nil {
			assert.FailNowf(t, "there must be error", "%v should not be inserted", v)
		}
	}

	// Insert Sample Data
	tmp := make(map[string]meta.ObjectMeta)
	for _, v := range ValidObjectMetas {
		id, err := db.Create(ctx, v)
		if !assert.NoError(t, err, "fail to insert document", id, v) {
			return
		}
		_, err = uuid.Parse(id)
		if !assert.NoError(t, err, "invalid id returned", id, v) {
			return
		}
		tmp[id] = v
	}

	// Delete Inserted Sample data
	defer func() {
		for id, _ := range tmp {
			err := db.Delete(ctx, id)
			if !assert.NoError(t, err, "fail to fetch document") {
				return
			}
		}
	}()

	// Test Just Get
	for id, v := range tmp {
		received, err := db.Get(ctx, id)
		if !assert.NoError(t, err, "fail to fetch document") {
			return
		}
		assert.Equal(t, received.Path, v.Path)
		assert.Equal(t, received.S3Key, v.S3Key)
	}

	// Test Update and Get
	for id, v := range tmp {
		newPath := fmt.Sprintf("some_path/%s", v.S3Key)
		err = db.Update(ctx, id, &meta.ObjectMeta{
			S3Key: newPath,
			Tags:  []string{"system:hello-world"},
		})
		if !assert.NoError(t, err, "fail to fetch document") {
			return
		}
		received, err := db.Get(ctx, id)
		if !assert.NoError(t, err, "fail to fetch document") {
			return
		}
		assert.NotEmpty(t, received.Tags)
		assert.Equal(t, received.Path, v.Path)
		assert.Equal(t, received.S3Key, newPath)
		assert.NotEqual(t, received.CreatedAt, received.UpdatedAt)
	}

	// Test List by Prefix
	{
		objs, err := db.ListByPathPrefix(ctx, "test")
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, 10, len(objs))
		for _, v := range objs {
			if strings.HasPrefix(v.Path, "test2") {
				assert.FailNow(t, "test2 Prefix should not be exists", v.Path)
			}
		}
	}
	// Test Sampling 5 data
	{
		objs, err := db.ListSamples(ctx, 5)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, 5, len(objs))

		for i, a := range objs {
			assert.NotEmpty(t, a.Path)
			assert.NotEmpty(t, a.S3Key)
			for j, b := range objs {
				if a.UUID == b.UUID && i != j {
					assert.Fail(t, "duplicated sample detected", a, b)
				}
			}
		}
	}

	// Count
	{
		stats, err := db.Stats(ctx)
		if !assert.NoError(t, err) {
			return
		}
		assert.NotZero(t, stats.ObjectCount)
	}
}
