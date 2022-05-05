package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/aaronland/go-sqlite"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	wof_tables "github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/warning"
	"io"
	"log"
	"sync"
)

// SQLiteFeaturesLoadRecordFuncOptions is a struct to define options when loading Who's On First feature records.
type SQLiteFeaturesLoadRecordFuncOptions struct {
	// StrictAltFiles is a boolean flag indicating whether the failure to load or parse an alternate geometry file should trigger a critical error.
	StrictAltFiles bool
}

// SQLiteFeaturesIndexRelationsFuncOptions
type SQLiteFeaturesIndexRelationsFuncOptions struct {
	// Reader is a valid `whosonfirst/go-reader` instance used to load Who's On First feature data
	Reader reader.Reader
	// Strict is a boolean flag indicating whether the failure to load or parse feature record should trigger a critical error.
	Strict bool
}

// SQLiteFeaturesLoadRecordFunc
func SQLiteFeaturesLoadRecordFunc(opts *SQLiteFeaturesLoadRecordFuncOptions) sql_index.SQLiteIndexerLoadRecordFunc {

	cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) (interface{}, error) {

		select {

		case <-ctx.Done():
			return nil, nil
		default:

			i, err := feature.LoadWOFFeatureFromReader(fh)

			if err != nil && !warning.IsWarning(err) {

				_, err := fh.Seek(0, 0)

				if err != nil {
					return nil, err
				}

				alt, alt_err := feature.LoadWOFAltFeatureFromReader(fh)

				if alt_err != nil && !warning.IsWarning(alt_err) {

					msg := fmt.Sprintf("Unable to load %s, because %s (%s)", path, alt_err, err)

					if !opts.StrictAltFiles {
						log.Printf("%s - SKIPPING\n", msg)
						return nil, nil
					}

					return nil, errors.New(msg)
				}

				i = alt
			}

			return i, nil
		}
	}

	return cb
}

func SQLiteFeaturesIndexRelationsFunc(r reader.Reader) sql_index.SQLiteIndexerPostIndexFunc {

	opts := &SQLiteFeaturesIndexRelationsFuncOptions{}
	opts.Reader = r

	return SQLiteFeaturesIndexRelationsFuncWithOptions(opts)
}

func SQLiteFeaturesIndexRelationsFuncWithOptions(opts *SQLiteFeaturesIndexRelationsFuncOptions) sql_index.SQLiteIndexerPostIndexFunc {

	seen := new(sync.Map)

	cb := func(ctx context.Context, db sqlite.Database, tables []sqlite.Table, record interface{}) error {

		geojson_t, err := wof_tables.NewGeoJSONTable(ctx)

		if err != nil {
			return err
		}

		conn, err := db.Conn()

		if err != nil {
			return err
		}

		f := record.(geojson.Feature)
		body := f.Bytes()

		relations := make(map[int64]bool)
		to_index := make([]geojson.Feature, 0)

		candidates := []string{
			"properties.wof:belongsto",
			"properties.wof:involves",
			"properties.wof:depicts",
		}

		for _, path := range candidates {

			// log.Println("RELATIONS", path)

			rsp := gjson.GetBytes(body, path)

			if !rsp.Exists() {
				// log.Println("MISSING", path)
				continue
			}

			for _, r := range rsp.Array() {

				id := r.Int()

				// skip -1, -4, etc.
				// (20201224/thisisaaronland)

				if id <= 0 {
					continue
				}

				relations[id] = true
			}
		}

		for id, _ := range relations {

			_, ok := seen.Load(id)

			if ok {
				continue
			}

			seen.Store(id, true)

			sql := fmt.Sprintf("SELECT COUNT(id) FROM %s WHERE id=?", geojson_t.Name())
			row := conn.QueryRow(sql, id)

			var count int
			err = row.Scan(&count)

			if err != nil {
				return err
			}

			if count != 0 {
				continue
			}

			rel_path, err := uri.Id2RelPath(id)

			if err != nil {
				return err
			}

			fh, err := opts.Reader.Read(ctx, rel_path)

			if err != nil {

				if opts.Strict {
					return err
				}

				log.Printf("Failed to read '%s' because '%v'. Strict mode is disabled so skipping\n", rel_path, err)
				continue
			}

			defer fh.Close()

			ancestor, err := feature.LoadFeatureFromReader(fh)

			// check for warnings in case this record has a non-standard
			// placetype (20201224/thisisaaronland)

			if err != nil && !warning.IsWarning(err) {

				if opts.Strict {
					return err
				}

				log.Printf("Failed to load feature for '%s' because '%v'. Strict mode is disabled so skipping\n", rel_path, err)
				continue
			}

			to_index = append(to_index, ancestor)

			// TO DO: CHECK WHETHER TO INDEX ALT FILES FOR ANCESTOR(S)
			// https://github.com/whosonfirst/go-whosonfirst-sqlite-features-index/issues/3
		}

		for _, record := range to_index {

			for _, t := range tables {

				err = t.IndexRecord(ctx, db, record)

				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return cb
}
