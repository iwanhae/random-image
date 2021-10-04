package meta

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iwanhae/random-image/pkg/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const IMAGE_COLLECTION = "image"

type MongoDatabase struct {
	client            *mongo.Client
	db                string
	objectCollenction *mongo.Collection
}

func InitMongoDB(ctx context.Context, c config.RandomImageConfig) (*MongoDatabase, error) {
	mg, err := mongo.Connect(ctx, options.Client().ApplyURI(c.MongoURI))
	if err != nil {
		return nil, err
	}
	err = mg.Ping(ctx, readpref.Nearest())
	if err != nil {
		return nil, err
	}
	T := true
	index := mg.Database(c.MongoDatabase).Collection(IMAGE_COLLECTION).Indexes()

	// Creating index
	if _, err = index.CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "uuid", Value: -1}},
		Options: &options.IndexOptions{Unique: &T},
	}); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("fail to create index, but will ignore this")
	}
	if _, err = index.CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "s3_key", Value: -1}},
		Options: &options.IndexOptions{},
	}); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("fail to create index, but will ignore this")
	}
	if _, err = index.CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "path", Value: "text"}},
		Options: &options.IndexOptions{},
	}); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("fail to create index, but will ignore this")
	}
	// ignore the failiure

	return &MongoDatabase{
		client:            mg,
		db:                c.MongoDatabase,
		objectCollenction: mg.Database(c.MongoDatabase).Collection(IMAGE_COLLECTION),
	}, nil
}

// Create new one. UUID field will be ignored
func (m *MongoDatabase) Create(ctx context.Context, obj ObjectMeta) (id string, err error) {
	col := m.objectCollenction

	id = uuid.New().String()
	now := time.Now()

	obj.UUID = &id
	obj.CreatedAt = &now
	obj.UpdatedAt = &now

	if obj.Path == "" {
		return "", NewDatabaseError(fmt.Errorf("path can not be empty"), ValidationFailed)
	} else if strings.HasPrefix(obj.Path, "/") {
		return "", NewDatabaseError(
			fmt.Errorf("path should not start with %q, but give %q", "/", obj.Path),
			ValidationFailed)
	}

	if obj.S3Key != "" && path.Ext(obj.S3Key) == "" {
		return "", NewDatabaseError(fmt.Errorf("s3 key should contains extention info"), ValidationFailed)
	}

	_, err = col.InsertOne(ctx, obj)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", NewDatabaseError(err, DuplicatedKey)
		}
	}
	return id, nil
}

// Get one with UUID
func (m *MongoDatabase) Get(ctx context.Context, uuid string) (obj *ObjectMeta, err error) {
	col := m.objectCollenction
	result := col.FindOne(ctx, bson.D{{Key: "uuid", Value: uuid}}, options.FindOne())

	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, NewDatabaseError(err, ObjectNotFound)
		}
		return nil, NewDatabaseError(err, BackendError)
	}

	res := ObjectMeta{}
	if err = result.Decode(&res); err != nil {
		return nil, NewDatabaseError(err, BackendError)
	}

	return &res, nil
}

func (m *MongoDatabase) ListByPathPrefix(ctx context.Context, pathPrefix string) (obj []ObjectMeta, err error) {
	col := m.objectCollenction
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix += "/"
	}
	q := fmt.Sprintf("^%v", pathPrefix)

	cur, err := col.Find(ctx, bson.D{{Key: "path", Value: bson.D{{
		Key:   "$regex",
		Value: q,
	}}}})
	if err != nil {
		return nil, NewDatabaseError(err, BackendError)
	}
	err = cur.All(ctx, &obj)
	if err != nil {
		return nil, NewDatabaseError(err, BackendError)
	}
	return obj, nil
}

func (m *MongoDatabase) ListSamples(ctx context.Context, count int) (obj []ObjectMeta, err error) {
	col := m.objectCollenction
	cur, err := col.Aggregate(ctx, bson.A{bson.D{
		{
			Key: "$sample",
			Value: bson.D{
				{Key: "size", Value: count},
			},
		},
	}})
	if err != nil {
		return nil, NewDatabaseError(err, BackendError)
	}
	err = cur.All(ctx, &obj)
	if err != nil {
		return nil, NewDatabaseError(err, BackendError)
	}
	return obj, nil
}

func (m *MongoDatabase) Update(ctx context.Context, uuid string, obj *ObjectMeta) error {
	if obj == nil {
		return NewDatabaseError(fmt.Errorf("nil object received"), ValidationFailed)
	}

	q := bson.D{}
	if obj.Cache != nil {
		if obj.Cache.S3Key1024px != "" {
			q = append(q, bson.E{
				Key:   "resized_1024px",
				Value: obj.Cache.S3Key1024px,
			})
		}
		if obj.Cache.S3Key768px != "" {
			q = append(q, bson.E{
				Key:   "resized_768px",
				Value: obj.Cache.S3Key1024px,
			})
		}
		if obj.Cache.S3Key480px != "" {
			q = append(q, bson.E{
				Key:   "resized_480px",
				Value: obj.Cache.S3Key1024px,
			})
		}
	}
	if obj.Path != "" {
		q = append(q, bson.E{
			Key:   "path",
			Value: obj.Path,
		})
	}
	if obj.S3Key != "" {
		q = append(q, bson.E{
			Key:   "s3_key",
			Value: obj.S3Key,
		})
	}
	if len(obj.Tags) != 0 {
		q = append(q, bson.E{
			Key:   "tags",
			Value: obj.Tags,
		})
	}
	q = append(q, bson.E{
		Key:   "updated_at",
		Value: time.Now(),
	})

	col := m.objectCollenction
	_, err := col.UpdateOne(ctx, bson.D{{Key: "uuid", Value: uuid}}, bson.D{{Key: "$set", Value: q}}, options.MergeUpdateOptions())
	if err != nil {
		if errors.Is(err, mongo.ErrNilDocument) {
			return NewDatabaseError(err, ObjectNotFound)
		}
		return NewDatabaseError(err, BackendError)
	}
	return nil
}

func (m *MongoDatabase) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return NewDatabaseError(fmt.Errorf("none uuid can not be deleted"), ValidationFailed)
	}
	col := m.objectCollenction
	_, err := col.DeleteOne(ctx, bson.D{{Key: "uuid", Value: id}}, &options.DeleteOptions{})
	if err != nil {
		return NewDatabaseError(err, BackendError)
	}
	return nil
}

func (m *MongoDatabase) Stats(ctx context.Context) (*DatabaseStats, error) {
	col := m.objectCollenction
	c, err := col.CountDocuments(ctx, bson.D{})
	if err != nil {
		return nil, NewDatabaseError(err, BackendError)
	}
	return &DatabaseStats{
		ObjectCount: c,
	}, nil
}
