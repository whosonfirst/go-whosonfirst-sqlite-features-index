package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	sql_index "github.com/whosonfirst/go-whosonfirst-sqlite-index"
	"github.com/whosonfirst/warning"
	"io"
	"io/ioutil"
	_ "log"
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

			body, err := ioutil.ReadAll(fh)

			if err != nil {
				return nil, err
			}

			i, err := feature.NewWOFFeature(body)

			if err != nil && !warning.IsWarning(err) {

				alt, alt_err := feature.NewWOFAltFeature(body)

				if alt_err != nil && !warning.IsWarning(alt_err) {
					msg := fmt.Sprintf("Unable to load %s, because %s (%s)", path, alt_err, err)
					return nil, errors.New(msg)
				}

				i = alt
			}

			return i, nil
		}
	}

	return sql_index.NewSQLiteIndexer(db, to_index, cb)
}
