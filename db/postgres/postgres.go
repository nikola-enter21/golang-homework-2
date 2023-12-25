package postgres

import (
	"awesomeProject1/db/model"
	"context"
	"database/sql"
	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

// NewPostgresDatabase creates a new PostgresDatabase instance
func NewPostgresDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (pdb *Database) SaveImage(ctx context.Context, image *model.Image) (*model.Image, error) {
	var inserted model.Image
	query := `
        INSERT INTO images (filename, alt_text, title, width, height, format, source_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING *
    `

	stmt, err := pdb.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, image.Filename, image.AltText, image.Title, image.Width, image.Height, image.Format, image.SourceURL).
		Scan(&inserted.Name, &inserted.Filename, &inserted.AltText, &inserted.Title, &inserted.Width, &inserted.Height, &inserted.Format, &inserted.SourceURL)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (pdb *Database) ListImages(ctx context.Context, filters *model.ListImagesFilters) ([]*model.Image, error) {
	query := `
        SELECT name, filename, alt_text, title, width, height, format, source_url
        FROM images
    `

	var args []interface{}
	if filters != nil && filters.Format != "" {
		query += " WHERE format = $1"
		args = append(args, filters.Format)
	}

	rows, err := pdb.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*model.Image

	for rows.Next() {
		img := &model.Image{}
		err := rows.Scan(&img.Name, &img.Filename, &img.AltText, &img.Title, &img.Width, &img.Height, &img.Format, &img.SourceURL)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

// Close closes the database connection
func (pdb *Database) Close() error {
	return pdb.db.Close()
}
