package postgresql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/JustForWorld/banner-shift/internal/storage"
	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

type Banner struct {
	id        int64
	content   string
	isActive  bool
	featureID int64
	tagIDs    []int64
	createdAT string
	updatedAT string
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
			id SERIAL PRIMARY KEY,
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
		ALTER TABLE banner_tag ADD FOREIGN KEY (banner_id) REFERENCES banner (id) ON DELETE CASCADE;
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
		ALTER TABLE banner_tag ADD CONSTRAINT unique_tag_feature UNIQUE (tag_id, feature_id);
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

func (s *Storage) CreateBanner(featureID int, tagIDs []int, content string, isActive bool) (int64, error) {
	const op = "storage.postgresql.CreateBanner"

	// checking required fields
	if featureID == 0 || content == "" || len(tagIDs) == 0 {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
	}

	// insert new banner
	stmtNewBanner, err := s.db.Prepare(`
		INSERT INTO banner(content, is_active, feature_id) VALUES($1, $2, $3) RETURNING id
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var bannerID int64
	err = stmtNewBanner.QueryRow(content, isActive, featureID).Scan(&bannerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "invalid_text_representation" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
		}
		return 0, fmt.Errorf("%s: failed to get last insert banner id: %w", op, err)
	}

	// insert banner with tagIDs
	for _, tagID := range tagIDs {
		stmtNewBannerTag, err := s.db.Prepare(`
			INSERT INTO banner_tag(banner_id, tag_id, feature_id)  VALUES($1, $2, $3)
		`)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}

		err = stmtNewBannerTag.QueryRow(bannerID, tagID, featureID).Err()
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrBannerExists)
			}
			return 0, fmt.Errorf("%s: failed to add banner_tag row: %w", op, err)
		}
	}

	return bannerID, nil
}

func (s *Storage) UpdateBanner(bannerID int64, featureID int64, tagIDs []int, content string, isActive bool) error {
	const op = "storage.postgresql.UpdateBanner"

	// update banner
	stmtUpdateBanner, err := s.db.Prepare(`
		UPDATE banner
		SET content = ($1), is_active = ($2), feature_id = ($3), updated_at = CURRENT_TIMESTAMP
		WHERE id = ($4);
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmtUpdateBanner.Exec(content, isActive, featureID, bannerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "invalid_text_representation" {
			return fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
		}
		return fmt.Errorf("%s: failed to get last insert banner id: %w", op, err)
	}

	// insert banner with tagIDs
	for _, tagID := range tagIDs {
		// TODO: change to UPDATE (there's no point now?)
		stmtNewBannerTag, err := s.db.Prepare(`
			INSERT INTO banner_tag(banner_id, tag_id, feature_id)  VALUES($1, $2, $3)
			ON CONFLICT (tag_id, feature_id) DO UPDATE SET banner_id = ($1), tag_id = ($2), feature_id = ($3);
		`)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = stmtNewBannerTag.QueryRow(bannerID, tagID, featureID).Err()
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "foreign_key_violation" {
				return fmt.Errorf("%s: %w", op, storage.ErrBannerNotExists)
			}
			return fmt.Errorf("%s: failed to update banner_tag row: %w", op, err)
		}
	}

	return nil
}

func (s *Storage) DeleteBanner(bannerID int64) error {
	const op = "storage.postgresql.DeleteBanner"

	// delete banner
	stmtDeleteBanner, err := s.db.Prepare(`
		DELETE FROM banner
		WHERE id = ($1);
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	result, err := stmtDeleteBanner.Exec(bannerID)
	if err != nil {
		return fmt.Errorf("%s: failed delete row: %w", op, err)
	}

	// cnt affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: error getting rows affected: %w", op, err)
	}

	// check if exist banner
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrBannerNotExists)
	}

	return nil
}

func (s *Storage) GetBanner(bannerID, featureID int64) (string, error) {
	const op = "storage.postgresql.GetBanner"

	// checking required fields
	if bannerID == 0 || featureID == 0 {
		return "", fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
	}

	// get banner
	stmtGetBanner, err := s.db.Prepare(`
		SELECT content FROM banner WHERE id = ($1) AND feature_id = ($2);
	`)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var content string
	err = stmtGetBanner.QueryRow(bannerID, featureID).Scan(&content)
	if err != nil {
		// if not exist
		if strings.Contains(err.Error(), "no rows in result set") {
			return "", fmt.Errorf("%s: %w", op, storage.ErrBannerNotExists)
		}
		return "", fmt.Errorf("%s: failed get content row: %w", op, err)
	}

	return content, nil
}

func (s *Storage) GetBannerList(featureID, tagID, limit, offset int64) ([]*Banner, error) {
	const op = "storage.postgresql.GetBannerList"

	// optional params in query
	query := `
		SELECT
		b.*,
		bt.tag_id
		FROM
		banner b
		LEFT JOIN banner_tag bt ON b.id = bt.banner_id	
	`
	var args []interface{}
	cntArgs := 0

	if featureID != 0 && tagID != 0 {
		cntArgs++
		query += fmt.Sprintf("WHERE b.feature_id = ($%v) AND", cntArgs)
		args = append(args, featureID)
		cntArgs++
		query += fmt.Sprintf("bt.tag_id = ($%v)", cntArgs)
		args = append(args, tagID)
	} else if featureID != 0 {
		cntArgs++
		query += fmt.Sprintf("WHERE b.feature_id = ($%v)", cntArgs)
		args = append(args, featureID)
	} else if tagID != 0 {
		cntArgs++
		query += fmt.Sprintf("WHERE bt.tag_id = ($%v)", cntArgs)
		args = append(args, tagID)
	}
	if limit != 0 {
		cntArgs++
		query += fmt.Sprintf("LIMIT ($%v);", cntArgs)
		args = append(args, limit)
	}

	// get banner list
	stmtGetBannerList, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmtGetBannerList.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	defer rows.Close()

	bannerMap := make(map[int64]*Banner)

	for rows.Next() {
		var banner Banner
		var currentTagID int64
		// read the lines from the query result and add them to the list
		if err := rows.Scan(&banner.id, &banner.content, &banner.isActive, &banner.featureID, &banner.createdAT, &banner.updatedAT, &currentTagID); err != nil {
			return nil, fmt.Errorf("%s: failed to scan rows: %w", op, err)
		}

		// check if a banner with this ID already exists
		if existingBanner, found := bannerMap[banner.id]; found {
			// if found, add a new tag_id
			existingBanner.tagIDs = append(existingBanner.tagIDs, currentTagID)
		} else {
			// if you haven't found it, create a new banner
			banner.tagIDs = append(banner.tagIDs, currentTagID)
			bannerMap[banner.id] = &banner
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: error during iteration: %w", op, err)
	}

	resultList := getBannerListFromMap(bannerMap)
	return resultList, nil
}

func getBannerListFromMap(bannerMap map[int64]*Banner) []*Banner {
	bannerList := make([]*Banner, 0, len(bannerMap))

	for _, banner := range bannerMap {
		bannerList = append(bannerList, banner)
	}

	return bannerList
}
