package main

// To do: Move all of this in to app/query

import (
	_ "github.com/whosonfirst/go-reader-http"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
)

import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features-index"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2/tables"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index/v4"
	"log"
	"os"
	"runtime"
	"strings"
)

func main() {

	valid_schemes := strings.Join(emitter.Schemes(), ",")
	iterator_desc := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-iterate/v2 URI. Supported emitter URI schemes are: %s", valid_schemes)

	iterator_uri := flag.String("iterator-uri", "repo://", iterator_desc)

	mode_desc := fmt.Sprintf("%s. THIS FLAG IS DEPRECATED, please use -iterator-uri instead.", iterator_desc)
	mode := flag.String("mode", "repo://", mode_desc)

	db_uri := flag.String("database-uri", "modernc://mem", "")

	all := flag.Bool("all", false, "Index all tables (except the 'search' and 'geometries' tables which you need to specify explicitly)")
	ancestors := flag.Bool("ancestors", false, "Index the 'ancestors' tables")
	concordances := flag.Bool("concordances", false, "Index the 'concordances' tables")
	geojson := flag.Bool("geojson", false, "Index the 'geojson' table")
	geometries := flag.Bool("geometries", false, "Index the 'geometries' table (requires that libspatialite already be installed)")
	names := flag.Bool("names", false, "Index the 'names' table")
	rtree := flag.Bool("rtree", false, "Index the 'rtree' table")
	properties := flag.Bool("properties", false, "Index the 'properties' table")
	search := flag.Bool("search", false, "Index the 'search' table (using SQLite FTS4 full-text indexer)")
	spr := flag.Bool("spr", false, "Index the 'spr' table")
	supersedes := flag.Bool("supersedes", false, "Index the 'supersedes' table")

	spatial_tables := flag.Bool("spatial-tables", false, "If true then index the necessary tables for use with the whosonfirst/go-whosonfirst-spatial-sqlite package.")

	live_hard := flag.Bool("live-hard-die-fast", true, "Enable various performance-related pragmas at the expense of possible (unlikely) database corruption")
	timings := flag.Bool("timings", false, "Display timings during and after indexing")
	optimize := flag.Bool("optimize", true, "Attempt to optimize the database before closing connection")

	alt_files := flag.Bool("index-alt-files", false, "Index alt geometries")
	strict_alt_files := flag.Bool("strict-alt-files", true, "Be strict when indexing alt geometries")

	index_relations := flag.Bool("index-relations", false, "Index the records related to a feature, specifically wof:belongsto, wof:depicts and wof:involves. Alt files for relations are not indexed at this time.")
	relations_uri := flag.String("index-relations-reader-uri", "", "A valid go-reader.Reader URI from which to read data for a relations candidate.")

	var procs = flag.Int("processes", (runtime.NumCPU() * 2), "The number of concurrent processes to index data with")

	flag.Parse()

	ctx := context.Background()

	if *iterator_uri == "" {
		*iterator_uri = *mode
	}

	runtime.GOMAXPROCS(*procs)

	logger := log.Default()

	if *spatial_tables {
		*rtree = true
		*geojson = true
		*properties = true
		*spr = true
	}

	db, err := sqlite.NewDatabase(ctx, *db_uri)

	if err != nil {
		logger.Fatalf("Unable to create database (%s) because %v", *db_uri, err)
	}

	// optimize query performance
	// https://www.sqlite.org/pragma.html#pragma_optimize
	if *optimize {

		defer func() {

			conn, err := db.Conn(ctx)

			if err != nil {
				logger.Fatalf("Unable to optimize, because %v", err)
			}

			_, err = conn.Exec("PRAGMA optimize")

			if err != nil {
				logger.Fatalf("Unable to optimize, because %v", err)
			}
		}()
	}

	defer db.Close(ctx)

	if *live_hard {

		err = sqlite.LiveHardDieFast(ctx, db)

		if err != nil {
			logger.Fatalf("Unable to live hard and die fast so just dying fast instead, because %v", err)
		}
	}

	to_index := make([]sqlite.Table, 0)

	if *geojson || *all {

		geojson_opts, err := tables.DefaultGeoJSONTableOptions()

		if err != nil {
			logger.Fatalf("failed to create 'geojson' table options because %s", err)
		}

		geojson_opts.IndexAltFiles = *alt_files

		gt, err := tables.NewGeoJSONTableWithDatabaseAndOptions(ctx, db, geojson_opts)

		if err != nil {
			logger.Fatalf("failed to create 'geojson' table because %s", err)
		}

		to_index = append(to_index, gt)
	}

	if *supersedes || *all {

		t, err := tables.NewSupersedesTableWithDatabase(ctx, db)

		if err != nil {
			logger.Fatalf("failed to create 'supersedes' table because %s", err)
		}

		to_index = append(to_index, t)
	}

	if *rtree || *all {

		rtree_opts, err := tables.DefaultRTreeTableOptions()

		if err != nil {
			logger.Fatalf("failed to create 'rtree' table options because %s", err)
		}

		rtree_opts.IndexAltFiles = *alt_files

		gt, err := tables.NewRTreeTableWithDatabaseAndOptions(ctx, db, rtree_opts)

		if err != nil {
			logger.Fatalf("failed to create 'rtree' table because %s", err)
		}

		to_index = append(to_index, gt)
	}

	if *properties || *all {

		properties_opts, err := tables.DefaultPropertiesTableOptions()

		if err != nil {
			logger.Fatalf("failed to create 'properties' table options because %s", err)
		}

		properties_opts.IndexAltFiles = *alt_files

		gt, err := tables.NewPropertiesTableWithDatabaseAndOptions(ctx, db, properties_opts)

		if err != nil {
			logger.Fatalf("failed to create 'properties' table because %s", err)
		}

		to_index = append(to_index, gt)
	}

	if *spr || *all {

		spr_opts, err := tables.DefaultSPRTableOptions()

		if err != nil {
			logger.Fatalf("Failed to create 'spr' table options because %v", err)
		}

		spr_opts.IndexAltFiles = *alt_files

		st, err := tables.NewSPRTableWithDatabaseAndOptions(ctx, db, spr_opts)

		if err != nil {
			logger.Fatalf("failed to create 'spr' table because %s", err)
		}

		to_index = append(to_index, st)
	}

	if *names || *all {

		nm, err := tables.NewNamesTableWithDatabase(ctx, db)

		if err != nil {
			logger.Fatalf("failed to create 'names' table because %s", err)
		}

		to_index = append(to_index, nm)
	}

	if *ancestors || *all {

		an, err := tables.NewAncestorsTableWithDatabase(ctx, db)

		if err != nil {
			logger.Fatalf("failed to create 'ancestors' table because %s", err)
		}

		to_index = append(to_index, an)
	}

	if *concordances || *all {

		cn, err := tables.NewConcordancesTableWithDatabase(ctx, db)

		if err != nil {
			logger.Fatalf("failed to create 'concordances' table because %s", err)
		}

		to_index = append(to_index, cn)
	}

	// see the way we don't check *all here - that's so people who don't have
	// spatialite installed can still use *all (20180122/thisisaaronland)

	if *geometries {

		geometries_opts, err := tables.DefaultGeometriesTableOptions()

		if err != nil {
			logger.Fatalf("failed to create 'geometries' table options because %v", err)
		}

		geometries_opts.IndexAltFiles = *alt_files

		gm, err := tables.NewGeometriesTableWithDatabaseAndOptions(ctx, db, geometries_opts)

		if err != nil {
			logger.Fatalf("failed to create 'geometries' table because %v", err)
		}

		to_index = append(to_index, gm)
	}

	// see the way we don't check *all here either - that's because this table can be
	// brutally slow to index and should probably really just be a separate database
	// anyway... (20180214/thisisaaronland)

	if *search {

		st, err := tables.NewSearchTableWithDatabase(ctx, db)

		if err != nil {
			logger.Fatalf("failed to create 'search' table because %v", err)
		}

		to_index = append(to_index, st)
	}

	if len(to_index) == 0 {
		logger.Fatalf("You forgot to specify which (any) tables to index")
	}

	record_opts := &index.SQLiteFeaturesLoadRecordFuncOptions{
		StrictAltFiles: *strict_alt_files,
	}

	record_func := index.SQLiteFeaturesLoadRecordFunc(record_opts)

	idx_opts := &sql_index.SQLiteIndexerOptions{
		DB:             db,
		Tables:         to_index,
		LoadRecordFunc: record_func,
	}

	if *index_relations {

		r, err := reader.NewReader(ctx, *relations_uri)

		if err != nil {
			logger.Fatalf("Failed to load reader (%s), %v", *relations_uri, err)
		}

		belongsto_func := index.SQLiteFeaturesIndexRelationsFunc(r)
		idx_opts.PostIndexFunc = belongsto_func
	}

	idx, err := sql_index.NewSQLiteIndexer(idx_opts)

	if err != nil {
		logger.Fatalf("failed to create sqlite indexer because %v", err)
	}

	idx.Timings = *timings
	idx.Logger = logger

	uris := flag.Args()

	err = idx.IndexURIs(ctx, *iterator_uri, uris...)

	if err != nil {
		logger.Fatalf("Failed to index paths in %s mode because: %s", *iterator_uri, err)
	}

	os.Exit(0)
}
