package index

import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features-index/v2"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2/tables"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index/v4"
	"log"
	"runtime"
)

func Run(ctx context.Context, logger *log.Logger) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs, logger)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet, logger *log.Logger) error {

	flagset.Parse(fs)

	runtime.GOMAXPROCS(procs)

	if spatial_tables {
		rtree = true
		geojson = true
		properties = true
		spr = true
	}

	db, err := sqlite.NewDatabase(ctx, db_uri)

	if err != nil {
		return fmt.Errorf("Unable to create database (%s) because %v", db_uri, err)
	}

	// optimize query performance
	// https://www.sqlite.org/pragma.html#pragma_optimize
	if optimize {

		defer func() {

			defer db.Close(ctx)

			conn, err := db.Conn(ctx)

			if err != nil {
				logger.Printf("Unable to optimize, because %v", err)
				return
			}

			_, err = conn.Exec("PRAGMA optimize")

			if err != nil {
				logger.Printf("Unable to optimize, because %v", err)
				return
			}
		}()

	} else {

		defer db.Close(ctx)
	}

	if live_hard {

		err = sqlite.LiveHardDieFast(ctx, db)

		if err != nil {
			return fmt.Errorf("Unable to live hard and die fast so just dying fast instead, because %v", err)
		}
	}

	to_index := make([]sqlite.Table, 0)

	if geojson || all {

		geojson_opts, err := tables.DefaultGeoJSONTableOptions()

		if err != nil {
			return fmt.Errorf("failed to create 'geojson' table options because %s", err)
		}

		geojson_opts.IndexAltFiles = alt_files

		gt, err := tables.NewGeoJSONTableWithDatabaseAndOptions(ctx, db, geojson_opts)

		if err != nil {
			return fmt.Errorf("failed to create 'geojson' table because %s", err)
		}

		to_index = append(to_index, gt)
	}

	if supersedes || all {

		t, err := tables.NewSupersedesTableWithDatabase(ctx, db)

		if err != nil {
			return fmt.Errorf("failed to create 'supersedes' table because %s", err)
		}

		to_index = append(to_index, t)
	}

	if rtree || all {

		rtree_opts, err := tables.DefaultRTreeTableOptions()

		if err != nil {
			return fmt.Errorf("failed to create 'rtree' table options because %s", err)
		}

		rtree_opts.IndexAltFiles = alt_files

		gt, err := tables.NewRTreeTableWithDatabaseAndOptions(ctx, db, rtree_opts)

		if err != nil {
			return fmt.Errorf("failed to create 'rtree' table because %s", err)
		}

		to_index = append(to_index, gt)
	}

	if properties || all {

		properties_opts, err := tables.DefaultPropertiesTableOptions()

		if err != nil {
			return fmt.Errorf("failed to create 'properties' table options because %s", err)
		}

		properties_opts.IndexAltFiles = alt_files

		gt, err := tables.NewPropertiesTableWithDatabaseAndOptions(ctx, db, properties_opts)

		if err != nil {
			return fmt.Errorf("failed to create 'properties' table because %s", err)
		}

		to_index = append(to_index, gt)
	}

	if spr || all {

		spr_opts, err := tables.DefaultSPRTableOptions()

		if err != nil {
			return fmt.Errorf("Failed to create 'spr' table options because %v", err)
		}

		spr_opts.IndexAltFiles = alt_files

		st, err := tables.NewSPRTableWithDatabaseAndOptions(ctx, db, spr_opts)

		if err != nil {
			return fmt.Errorf("failed to create 'spr' table because %s", err)
		}

		to_index = append(to_index, st)
	}

	if names || all {

		nm, err := tables.NewNamesTableWithDatabase(ctx, db)

		if err != nil {
			return fmt.Errorf("failed to create 'names' table because %s", err)
		}

		to_index = append(to_index, nm)
	}

	if ancestors || all {

		an, err := tables.NewAncestorsTableWithDatabase(ctx, db)

		if err != nil {
			return fmt.Errorf("failed to create 'ancestors' table because %s", err)
		}

		to_index = append(to_index, an)
	}

	if concordances || all {

		cn, err := tables.NewConcordancesTableWithDatabase(ctx, db)

		if err != nil {
			return fmt.Errorf("failed to create 'concordances' table because %s", err)
		}

		to_index = append(to_index, cn)
	}

	// see the way we don't check all here - that's so people who don't have
	// spatialite installed can still use all (20180122/thisisaaronland)

	if geometries {

		geometries_opts, err := tables.DefaultGeometriesTableOptions()

		if err != nil {
			return fmt.Errorf("failed to create 'geometries' table options because %v", err)
		}

		geometries_opts.IndexAltFiles = alt_files

		gm, err := tables.NewGeometriesTableWithDatabaseAndOptions(ctx, db, geometries_opts)

		if err != nil {
			return fmt.Errorf("failed to create 'geometries' table because %v", err)
		}

		to_index = append(to_index, gm)
	}

	// see the way we don't check all here either - that's because this table can be
	// brutally slow to index and should probably really just be a separate database
	// anyway... (20180214/thisisaaronland)

	if search {

		st, err := tables.NewSearchTableWithDatabase(ctx, db)

		if err != nil {
			return fmt.Errorf("failed to create 'search' table because %v", err)
		}

		to_index = append(to_index, st)
	}

	if len(to_index) == 0 {
		return fmt.Errorf("You forgot to specify which (any) tables to index")
	}

	record_opts := &index.SQLiteFeaturesLoadRecordFuncOptions{
		StrictAltFiles: strict_alt_files,
	}

	record_func := index.SQLiteFeaturesLoadRecordFunc(record_opts)

	idx_opts := &sql_index.SQLiteIndexerOptions{
		DB:             db,
		Tables:         to_index,
		LoadRecordFunc: record_func,
	}

	if index_relations {

		r, err := reader.NewReader(ctx, relations_uri)

		if err != nil {
			return fmt.Errorf("Failed to load reader (%s), %v", relations_uri, err)
		}

		belongsto_func := index.SQLiteFeaturesIndexRelationsFunc(r)
		idx_opts.PostIndexFunc = belongsto_func
	}

	idx, err := sql_index.NewSQLiteIndexer(idx_opts)

	if err != nil {
		return fmt.Errorf("failed to create sqlite indexer because %v", err)
	}

	idx.Timings = timings
	idx.Logger = logger

	uris := fs.Args()

	err = idx.IndexURIs(ctx, iterator_uri, uris...)

	if err != nil {
		return fmt.Errorf("Failed to index paths in %s mode because: %s", iterator_uri, err)
	}

	return nil
}
