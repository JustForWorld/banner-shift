package postgresql

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(user, password, host, dbname string, port int) (*Storage, error) {
	const op = "storage.postgresql.New"

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, fmt.Errorf("1: %s: %w", op, err)
	}

	// create banner table
	stmtBanner, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS banner (
			id SERIAL PRIMARY KEY,
			content JSONB,
			is_active BOOLEAN,
			feature_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("create banner: %s: %w", op, err)
	}

	_, err = stmtBanner.Exec()
	if err != nil {
		return nil, fmt.Errorf("exec banner: %s: %w", op, err)
	}

	// create banner_tag table
	stmtBannerTag, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS banner_tag (
			banner_id INTEGER,
			tag_id INTEGER,
			feature_id INTEGER
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("create banner_tag: %s: %w", op, err)
	}

	_, err = stmtBannerTag.Exec()
	if err != nil {
		return nil, fmt.Errorf("exec banner_tag: %s: %w", op, err)
	}

	// create tag table
	stmtTag, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS tag (
			id SERIAL PRIMARY KEY
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("create tag: %s: %w", op, err)
	}

	_, err = stmtTag.Exec()
	if err != nil {
		return nil, fmt.Errorf("exec tag: %s: %w", op, err)
	}

	// create feature table
	stmtFeature, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS feature (
			id SERIAL PRIMARY KEY
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("create feature: %s: %w", op, err)
	}

	_, err = stmtFeature.Exec()
	if err != nil {
		return nil, fmt.Errorf("exec feature: %s: %w", op, err)
	}

	// create relations between banner < banner_tag tables
	stmtRelBanner, err := db.Prepare(`
		ALTER TABLE banner_tag ADD FOREIGN KEY (banner_id) REFERENCES banner (id);
	`)
	if err != nil {
		return nil, fmt.Errorf("create relation between banner < banner_tag: %s: %w", op, err)
	}

	_, err = stmtRelBanner.Exec()
	if err != nil {
		return nil, fmt.Errorf("exec relation between banner < banner_tag: %s: %w", op, err)
	}

	// create relations between banner > feature tables
	stmtRelFeature, err := db.Prepare(`
		ALTER TABLE banner ADD FOREIGN KEY (feature_id) REFERENCES feature (id);
	`)
	if err != nil {
		return nil, fmt.Errorf("create relation between banner > feature: %s: %w", op, err)
	}

	_, err = stmtRelFeature.Exec()
	if err != nil {
		return nil, fmt.Errorf("exec relation between banner > feature: %s: %w", op, err)
	}

	// create relations between tag < banner_tag tables
	stmtRelTag, err := db.Prepare(`
		ALTER TABLE banner_tag ADD FOREIGN KEY (tag_id) REFERENCES tag (id);
	`)
	if err != nil {
		return nil, fmt.Errorf("create relation between tag < banner_tag: %s: %w", op, err)
	}

	_, err = stmtRelTag.Exec()
	if err != nil {
		return nil, fmt.Errorf("exec relation between tag < banner_tag: %s: %w", op, err)
	}

	// create UNIQUE CONSTAINT with tag_id and feature_id
	stmtUniqueTagFeature, err := db.Prepare(`
		ALTER TABLE banner_tag ADD CONSTRAINT unique_tag_feuture UNIQUE (tag_id, feature_id);
	`)
	if err != nil {
		return nil, fmt.Errorf("create UNIQUE CONSTAINT with tag_id and feature_id: %s: %w", op, err)
	}

	_, err = stmtUniqueTagFeature.Exec()
	if err != nil {
		return nil, fmt.Errorf("exec UNIQUE CONSTAINT with tag_id and feature_id: %s: %w", op, err)
	}

	return &Storage{db: db}, nil
}
