package manager

import (
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

type DuckDBManager struct {
	db *sql.DB
}

func NewDuckDBManager() (*DuckDBManager, error) {
	db, err := sql.Open("duckdb", "?access_mode=READ_WRITE")
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}
	return &DuckDBManager{db: db}, nil
}

func (d *DuckDBManager) QueryParquet(filePath string, query string) error {
	_, err := d.db.Exec(fmt.Sprintf("CREATE TABLE temp AS SELECT * FROM parquet_scan('%s')", filePath))
	if err != nil {
		return fmt.Errorf("failed to create table from Parquet: %w", err)
	}

	rows, err := d.db.Query(query)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		// Process rows
	}

	return nil
}

func (d *DuckDBManager) Close() error {
	return d.db.Close()
}
