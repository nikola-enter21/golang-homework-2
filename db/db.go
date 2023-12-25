package db

import (
	"awesomeProject1/db/model"
	"golang.org/x/net/context"
)

// DB is the interface for database operations
type DB interface {
	SaveImage(context.Context, *model.Image) (*model.Image, error)
	ListImages(ctx context.Context, filters *model.ListImagesFilters) ([]*model.Image, error)
	Close() error
}
