package repository

import (
	"database/sql"
	model "remarks/internal/models"
	"time"
)

type StoreInterface interface {
	GetRemarksByProjectID(projectID int) ([]*model.Remark, error)
	AddRemark(remark *model.Remark) error
}

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) StoreInterface {
	return &SQLStore{
		db: db,
	}
}

func (s *SQLStore) GetRemarksByProjectID(projectID int) ([]*model.Remark, error) {
	remarks := []*model.Remark{}

	query := `
        SELECT id, project_id, direction, section, subsection, content, urgency, created_at
        FROM remarks
        WHERE project_id = $1
    `
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var remark model.Remark
		err := rows.Scan(
			&remark.ID,
			&remark.ProjectID,
			&remark.Direction,
			&remark.Section,
			&remark.Subsection,
			&remark.Content,
			&remark.Urgency,
			&remark.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		remarks = append(remarks, &remark)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return remarks, nil
}

func (s *SQLStore) AddRemark(remark *model.Remark) error {
	query := `
        INSERT INTO remarks (
            project_id,
            direction,
            section,
            subsection,
            content,
            urgency,
            created_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7
        )
    `
	remark.CreatedAt = time.Now()

	_, err := s.db.Exec(
		query,
		remark.ProjectID,
		remark.Direction,
		remark.Section,
		remark.Subsection,
		remark.Content,
		remark.Urgency,
		remark.CreatedAt,
	)

	return err
}
