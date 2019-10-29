package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	wof_utils "github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index"
	"github.com/whosonfirst/warning"
	"io"
	"io/ioutil"
)

func NewDefaultSQLiteFeaturesIndexer(db sqlite.Database, to_index []sqlite.Table) (*sql_index.SQLiteIndexer, error) {

	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) (interface{}, error) {

		select {

		case <-ctx.Done():
			return nil, nil
		default:

			path, err := wof_index.PathForContext(ctx)

			if err != nil {
				return nil, err
			}

			// TO DO ... something something something allow alt files?
			// depends on this: https://github.com/whosonfirst/go-whosonfirst-sqlite-features/tree/alt
			// https://github.com/whosonfirst/go-whosonfirst-sqlite-features/blob/master/tables/geojson.go
			// https://github.com/whosonfirst/go-whosonfirst-sqlite-features-index/blob/master/cmd/wof-sqlite-index-features/main.go#L97
			// https://github.com/whosonfirst/go-whosonfirst-sqlite-features/blob/master/tables/geojson.go#L48

			ok, err := wof_utils.IsPrincipalWOFRecord(fh, ctx)

			if err != nil {
				return nil, err
			}

			if !ok {
				return nil, nil
			}

			closer := ioutil.NopCloser(fh)

			// this will probably trigger an error for alt files?
			
			i, err := feature.LoadWOFFeatureFromReader(closer)

			// because this:
			// https://github.com/whosonfirst/go-whosonfirst-dist/issues/14
			// i, err := feature.LoadGeoJSONFeatureFromReader(closer)

			if err != nil && !warning.IsWarning(err) {
				msg := fmt.Sprintf("Unable to load %s, because %s", path, err)
				return nil, errors.New(msg)
			}

			return i, nil
		}
	}

	return sql_index.NewSQLiteIndexer(db, to_index, cb)
}
