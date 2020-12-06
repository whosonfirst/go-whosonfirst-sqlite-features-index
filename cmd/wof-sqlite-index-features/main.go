package main

import (
	_ "github.com/whosonfirst/go-reader-http"
	_ "github.com/whosonfirst/go-whosonfirst-index-csv"
	_ "github.com/whosonfirst/go-whosonfirst-index-sqlite"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"
)

import (
	"context"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-reader"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	log "github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features-index"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"io"
	"os"
	"runtime"
	"strings"
)

func main() {

	valid_modes := strings.Join(wof_index.Modes(), ",")
	desc_modes := fmt.Sprintf("The mode to use importing data. Valid modes are: %s.", valid_modes)

	dsn := flag.String("dsn", ":memory:", "")
	driver := flag.String("driver", "sqlite3", "")

	mode := flag.String("mode", "files", desc_modes)

	all := flag.Bool("all", false, "Index all tables (except the 'search' and 'geometries' tables which you need to specify explicitly)")
	ancestors := flag.Bool("ancestors", false, "Index the 'ancestors' tables")
	concordances := flag.Bool("concordances", false, "Index the 'concordances' tables")
	geojson := flag.Bool("geojson", false, "Index the 'geojson' table")
	geometries := flag.Bool("geometries", false, "Index the 'geometries' table (requires that libspatialite already be installed)")
	names := flag.Bool("names", false, "Index the 'names' table")
	rtree := flag.Bool("rtree", false, "Index the 'rtree' table")	
	search := flag.Bool("search", false, "Index the 'search' table (using SQLite FTS4 full-text indexer)")
	spr := flag.Bool("spr", false, "Index the 'spr' table")
	live_hard := flag.Bool("live-hard-die-fast", true, "Enable various performance-related pragmas at the expense of possible (unlikely) database corruption")
	timings := flag.Bool("timings", false, "Display timings during and after indexing")
	optimize := flag.Bool("optimize", true, "Attempt to optimize the database before closing connection")

	alt_files := flag.Bool("index-alt-files", false, "Index alt geometries")
	strict_alt_files := flag.Bool("strict-alt-files", true, "Be strict when indexing alt geometries")

	index_relations := flag.Bool("index-relations", false, "Index the records related to a feature, specifically wof:belongsto, wof:depicts and wof:involves. Alt files for relations are not indexed at this time.")
	relations_uri := flag.String("index-relations-reader-uri", "", "A valid go-reader.Reader URI from which to read data for a relations candidate.")

	var procs = flag.Int("processes", (runtime.NumCPU() * 2), "The number of concurrent processes to index data with")

	flag.Parse()

	runtime.GOMAXPROCS(*procs)

	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

	if *geometries && *driver != "spatialite" {
		logger.Fatal("you asked to index geometries but specified the '%s' driver instead of spatialite", *driver)
	}

	db, err := database.NewDBWithDriver(*driver, *dsn)

	if err != nil {
		logger.Fatal("unable to create database (%s) because %s", *dsn, err)
	}

	// optimize query performance
	// https://www.sqlite.org/pragma.html#pragma_optimize
	if *optimize {

		defer func() {
			conn, err := db.Conn()
			if err != nil {
				logger.Fatal("Unable to optimize, because %s", err)
			}

			logger.Info("Optimizing database...")

			_, err = conn.Exec("PRAGMA optimize")
			if err != nil {
				logger.Fatal("Unable to optimize, because %s", err)
			}
		}()
	}

	defer db.Close()

	if *live_hard {

		err = db.LiveHardDieFast()

		if err != nil {
			logger.Fatal("Unable to live hard and die fast so just dying fast instead, because %s", err)
		}
	}

	to_index := make([]sqlite.Table, 0)

	if *geojson || *all {

		geojson_opts, err := tables.DefaultGeoJSONTableOptions()

		if err != nil {
			logger.Fatal("failed to create 'geojson' table options because %s", err)
		}

		geojson_opts.IndexAltFiles = *alt_files

		gt, err := tables.NewGeoJSONTableWithDatabaseAndOptions(db, geojson_opts)

		if err != nil {
			logger.Fatal("failed to create 'geojson' table because %s", err)
		}

		to_index = append(to_index, gt)
	}

	if *rtree || *all {

		rtree_opts, err := tables.DefaultRTreeTableOptions()

		if err != nil {
			logger.Fatal("failed to create 'rtree' table options because %s", err)
		}

		rtree_opts.IndexAltFiles = *alt_files

		gt, err := tables.NewRTreeTableWithDatabaseAndOptions(db, rtree_opts)

		if err != nil {
			logger.Fatal("failed to create 'rtree' table because %s", err)
		}

		to_index = append(to_index, gt)
	}
	
	if *spr || *all {

		st, err := tables.NewSPRTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'spr' table because %s", err)
		}

		to_index = append(to_index, st)
	}

	if *names || *all {

		nm, err := tables.NewNamesTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'names' table because %s", err)
		}

		to_index = append(to_index, nm)
	}

	if *ancestors || *all {

		an, err := tables.NewAncestorsTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'ancestors' table because %s", err)
		}

		to_index = append(to_index, an)
	}

	if *concordances || *all {

		cn, err := tables.NewConcordancesTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'concordances' table because %s", err)
		}

		to_index = append(to_index, cn)
	}

	// see the way we don't check *all here - that's so people who don't have
	// spatialite installed can still use *all (20180122/thisisaaronland)

	if *geometries {

		geometries_opts, err := tables.DefaultGeometriesTableOptions()

		if err != nil {
			logger.Fatal("failed to create 'geometries' table options because %s", err)
		}

		geometries_opts.IndexAltFiles = *alt_files

		gm, err := tables.NewGeometriesTableWithDatabaseAndOptions(db, geometries_opts)

		if err != nil {
			logger.Fatal("failed to create 'geometries' table because %s", err)
		}

		to_index = append(to_index, gm)
	}

	// see the way we don't check *all here either - that's because this table can be
	// brutally slow to index and should probably really just be a separate database
	// anyway... (20180214/thisisaaronland)

	if *search {

		st, err := tables.NewSearchTableWithDatabase(db)

		if err != nil {
			logger.Fatal("failed to create 'search' table because %s", err)
		}

		to_index = append(to_index, st)
	}

	if len(to_index) == 0 {
		logger.Fatal("You forgot to specify which (any) tables to index")
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

		ctx := context.Background()
		r, err := reader.NewReader(ctx, *relations_uri)

		if err != nil {
			logger.Fatal("Failed to load reader (%s), %v", *relations_uri, err)
		}

		belongsto_func := index.SQLiteFeaturesIndexRelationsFunc(r)
		idx_opts.PostIndexFunc = belongsto_func
	}

	idx, err := sql_index.NewSQLiteIndexer(idx_opts)

	if err != nil {
		logger.Fatal("failed to create sqlite indexer because %s", err)
	}

	idx.Timings = *timings
	idx.Logger = logger

	err = idx.IndexPaths(*mode, flag.Args())

	if err != nil {
		logger.Fatal("Failed to index paths in %s mode because: %s", *mode, err)
	}

	os.Exit(0)
}
