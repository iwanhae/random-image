package meta

import (
	"context"
	"fmt"
	"time"
)

type DatabaseStats struct {
	ObjectCount      int64
	ContentTypeStats map[string]int
}

type ResizedObject struct {
	S3Key480px  string `bson:"resized_480px"`
	S3Key768px  string `bson:"resized_768px"`
	S3Key1024px string `bson:"resized_1024px"`
}

type ObjectMeta struct {
	UUID      *string        `json:"uuid,omitempty" bson:"uuid"`
	Path      string         `json:"path" bson:"path"`
	Tags      []string       `json:"tags" bson:"tags"`
	S3Key     string         `json:"s3_key" bson:"s3_key"`
	Cache     *ResizedObject `bson:",omitempty,inline"`
	CreatedAt *time.Time     `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at,omitempty" bson:"updated_at"`
}

type DatabaseErrorType string

type DatabaseError struct {
	cause  error
	reason DatabaseErrorType
}

func (d *DatabaseError) Unwrap() error {
	return d.cause
}

func (d *DatabaseError) Error() string {
	return fmt.Sprintf("db failed: %s: %s", d.reason, d.cause.Error())
}

func NewDatabaseError(cause error, reason DatabaseErrorType) error {
	return &DatabaseError{cause: cause, reason: reason}
}

const (
	DuplicatedKey    DatabaseErrorType = "duplicated key can not be inserted"
	ValidationFailed DatabaseErrorType = "validation failed"
	ObjectNotFound   DatabaseErrorType = "object with given uuid not found"
	BackendError     DatabaseErrorType = "backend db emit error"
)

type Database interface {
	// Create new one. UUID field will be ignored
	Create(ctx context.Context, obj ObjectMeta) (uuid string, err error)

	// Get one with UUID
	Get(ctx context.Context, uuid string) (obj *ObjectMeta, err error)
	ListByPathPrefix(ctx context.Context, group string) (obj []ObjectMeta, err error)
	ListSamples(ctx context.Context, count int) (obj []ObjectMeta, err error)

	Update(ctx context.Context, uuid string, obj *ObjectMeta) error

	Delete(ctx context.Context, uuid string) error

	Stats(ctx context.Context) (*DatabaseStats, error)
}
