package postgresql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JustForWorld/banner-shift/internal/storage"
	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

type Banner struct {
	BannerID  int64       `json:"banner_id"`
	TagIDs    []int64     `json:"tag_ids"`
	FeatureID int64       `json:"feature_id"`
	Content   interface{} `json:"content"`
	IsActive  string      `json:"is_active"`
	CreatedAT string      `json:"created_at"`
	UpdatedAT string      `json:"updated_at"`
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

func (s *Storage) CreateBanner(featureID int64, tagIDs []int, content interface{}, isActive bool) (int64, error) {
	const op = "storage.postgresql.CreateBanner"

	// checking required fields
	if featureID == 0 || content == nil || len(tagIDs) == 0 {
		return 0, fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to start transaction: %w", op, err)
	}
	defer tx.Rollback()

	// insert new banner
	stmtNewBanner, err := s.db.Prepare(`
		INSERT INTO banner(content, is_active, feature_id) VALUES($1, $2, $3) RETURNING id
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmtNewBanner.Close()

	var bannerID int64
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	err = tx.Stmt(stmtNewBanner).QueryRow(contentBytes, isActive, featureID).Scan(&bannerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && (pqErr.Code.Name() == "invalid_text_representation" || pqErr.Code.Name() == "foreign_key_violation") {
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

		_, err = tx.Stmt(stmtNewBannerTag).Exec(bannerID, tagID, featureID)
		if err != nil {
			tx.Rollback()
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrBannerExists)
			}
			return 0, fmt.Errorf("%s: failed to add banner_tag row: %w", op, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return bannerID, nil
}

func (s *Storage) UpdateBanner(bannerID int64, featureID int64, tagIDs []int, content interface{}, isActive interface{}) error {
	const op = "storage.postgresql.UpdateBanner"

	// Check if feature_id has changed
	var existingFeatureID int64
	err := s.db.QueryRow(`SELECT feature_id FROM banner WHERE id = $1`, bannerID).Scan(&existingFeatureID)
	if err != nil {
		return fmt.Errorf("%s: failed to get existing feature_id: %w", op, err)
	}

	if existingFeatureID != featureID {
		// Delete existing banner_tag entries for the given banner_id and existingFeatureID
		stmtDeleteBannerTags, err := s.db.Prepare(`DELETE FROM banner_tag WHERE banner_id = $1 AND feature_id = $2`)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		_, err = stmtDeleteBannerTags.Exec(bannerID, existingFeatureID)
		if err != nil {
			return fmt.Errorf("%s: failed to delete existing banner_tag entries: %w", op, err)
		}
	}

	// optional params in query
	query := `UPDATE banner SET updated_at = CURRENT_TIMESTAMP `
	var args []interface{}
	cntArgs := 0
	queryRow := make([]string, 1, 4)
	queryRow[0] = ", "
	var queryParams string

	if content != nil {
		contentBytes, err := json.Marshal(content)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		cntArgs++
		queryRow = append(queryRow, fmt.Sprintf("content = ($%v)", cntArgs))
		args = append(args, contentBytes)
	}
	switch v := isActive.(type) {
	case bool:
		isActiveBool := v
		cntArgs++
		queryRow = append(queryRow, fmt.Sprintf("is_active = ($%v)", cntArgs))
		args = append(args, isActiveBool)
	case string:
		currentV := v
		if currentV != "" {
			return fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
		}
	default:
		return fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
	}
	if featureID != 0 {
		cntArgs++
		queryRow = append(queryRow, fmt.Sprintf("feature_id = ($%v)", cntArgs))
		args = append(args, featureID)
	}
	if len(queryRow) > 1 {
		queryParams = strings.Join(queryRow[1:], ", ")
		queryParams = ", " + queryParams
		query += queryParams
	}
	args = append(args, bannerID)
	cntArgs++
	query += fmt.Sprintf(` WHERE id = ($%v);`, cntArgs)

	// update banner
	stmtUpdateBanner, err := s.db.Prepare(query)
	if err != nil {
		fmt.Println(query)
		return fmt.Errorf(" %s: %w", op, err)
	}

	_, err = stmtUpdateBanner.Exec(args...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "invalid_text_representation" {
			return fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
		}
		fmt.Println(query)
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

	// checking required field
	if bannerID == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: failed to start transaction: %w", op, err)
	}
	defer tx.Rollback()

	// delete banner
	stmtDeleteBanner, err := s.db.Prepare(`
		DELETE FROM banner
		WHERE id = ($1);
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmtDeleteBanner.Close()

	result, err := tx.Stmt(stmtDeleteBanner).Exec(bannerID)
	if err != nil {
		tx.Rollback()
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return nil
}

func (s *Storage) GetBanner(tagID, featureID int64) (string, error) {
	const op = "storage.postgresql.GetBanner"

	// checking required fields
	if tagID == 0 || featureID == 0 {
		return "", fmt.Errorf("%s: %w", op, storage.ErrBannerInvalidData)
	}

	// get banner
	stmtGetBanner, err := s.db.Prepare(`
        SELECT b.content
        FROM banner b
        JOIN banner_tag bt ON b.id = bt.banner_id
        WHERE b.feature_id = $1 AND bt.tag_id = $2;
	`)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var content string
	err = stmtGetBanner.QueryRow(featureID, tagID).Scan(&content)
	if err != nil {
		// if not exist
		if strings.Contains(err.Error(), "no rows in result set") {
			return "", fmt.Errorf("%s: %w", op, storage.ErrBannerNotFound)
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
		query += fmt.Sprintf("WHERE b.feature_id = ($%v) AND ", cntArgs)
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
		// var contentJSON []byte
		// read the lines from the query result and add them to the list
		var stringContent string
		if err := rows.Scan(&banner.BannerID, &stringContent, &banner.IsActive, &banner.FeatureID, &banner.CreatedAT, &banner.UpdatedAT, &currentTagID); err != nil {
			return nil, fmt.Errorf("%s: failed to scan rows: %w", op, err)
		}
		banner.Content = stringContent

		// check if a banner with this ID already exists
		if existingBanner, found := bannerMap[banner.BannerID]; found {
			// if found, add a new tag_id
			existingBanner.TagIDs = append(existingBanner.TagIDs, currentTagID)
		} else {
			// if you haven't found it, create a new banner
			banner.TagIDs = append(banner.TagIDs, currentTagID)
			bannerMap[banner.BannerID] = &banner
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
