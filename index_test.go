package index

import (
	"context"
	"testing"
	"github.com/aaronland/go-sqlite"
	"fmt"
	"github.com/aaronland/go-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index/v3"
	"path/filepath"
	"github.com/whosonfirst/go-reader"	
)

func TestIndexFeatures(t *testing.T) {

	ctx := context.Background()

	path_fixtures, err := filepath.Abs("fixtures")

	if err != nil {
		t.Fatalf("Failed to determine path for fixtures, %v", err)
	}

	path_relations := filepath.Join(path_fixtures, "relations")
	path_data := filepath.Join(path_fixtures, "data")	
	
	driver := "sqlite3"
	dsn := ":memory:"

	db, err := database.NewDBWithDriver(ctx, driver, dsn)

	if err != nil {
		t.Fatalf("Unable to create database (%s) because %v", dsn, err)
	}

	to_index := make([]sqlite.Table, 0)

	geojson_opts, err := tables.DefaultGeoJSONTableOptions()

	if err != nil {
		t.Fatalf("failed to create 'geojson' table options because %v", err)
	}

	geojson_opts.IndexAltFiles = true

	gt, err := tables.NewGeoJSONTableWithDatabaseAndOptions(ctx, db, geojson_opts)

	if err != nil {
		t.Fatalf("failed to create 'geojson' table because %v", err)
	}

	to_index = append(to_index, gt)

	record_opts := &SQLiteFeaturesLoadRecordFuncOptions{}

	record_func := SQLiteFeaturesLoadRecordFunc(record_opts)

	idx_opts := &sql_index.SQLiteIndexerOptions{
		DB:             db,
		Tables:         to_index,
		LoadRecordFunc: record_func,
	}

	reader_uri := fmt.Sprintf("fs://%s?allow_bz2=1", path_relations)
	
	r, err := reader.NewReader(ctx, reader_uri)
	
	if err != nil {
		t.Fatalf("Failed to load reader (%s), %v", reader_uri, err)
	}
	
	belongsto_func := SQLiteFeaturesIndexRelationsFunc(r)
	idx_opts.PostIndexFunc = belongsto_func
	
	idx, err := sql_index.NewSQLiteIndexer(idx_opts)

	if err != nil {
		t.Fatalf("Failed to create sqlite indexer because %v", err)
	}

	// Blocked on changes to go-whosonfirst-sqlite-features
	// See 'props' branch for details
	
	err = idx.IndexURIs(ctx, "directory://", path_data)

	if err != nil {
		t.Fatalf("Failed to index paths, %v", err)
	}
	
}