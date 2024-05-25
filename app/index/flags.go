package index

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
)

var iterator_uri string

var db_uri string

var all bool
var ancestors bool
var concordances bool
var geojson bool
var spelunker bool
var geometries bool
var names bool
var rtree bool
var properties bool
var search bool
var spr bool
var supersedes bool

var spatial_tables bool
var spelunker_tables bool

var live_hard bool
var timings bool
var optimize bool

var alt_files bool
var strict_alt_files bool

var index_alt multi.MultiString

var index_relations bool
var relations_uri string

var procs int

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("index")

	valid_schemes := strings.Join(emitter.Schemes(), ",")
	iterator_desc := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-iterate/v2 URI. Supported emitter URI schemes are: %s", valid_schemes)

	fs.StringVar(&iterator_uri, "iterator-uri", "repo://", iterator_desc)

	fs.StringVar(&db_uri, "database-uri", "modernc://mem", "")

	fs.BoolVar(&all, "all", false, "Index all tables (except the 'search' and 'geometries' tables which you need to specify explicitly)")
	fs.BoolVar(&ancestors, "ancestors", false, "Index the 'ancestors' tables")
	fs.BoolVar(&concordances, "concordances", false, "Index the 'concordances' tables")
	fs.BoolVar(&geojson, "geojson", false, "Index the 'geojson' table")
	fs.BoolVar(&spelunker, "spelunker", false, "Index the 'spelunker' table")
	fs.BoolVar(&geometries, "geometries", false, "Index the 'geometries' table (requires that libspatialite already be installed)")
	fs.BoolVar(&names, "names", false, "Index the 'names' table")
	fs.BoolVar(&rtree, "rtree", false, "Index the 'rtree' table")
	fs.BoolVar(&properties, "properties", false, "Index the 'properties' table")
	fs.BoolVar(&search, "search", false, "Index the 'search' table (using SQLite FTS4 full-text indexer)")
	fs.BoolVar(&spr, "spr", false, "Index the 'spr' table")
	fs.BoolVar(&supersedes, "supersedes", false, "Index the 'supersedes' table")

	fs.BoolVar(&spatial_tables, "spatial-tables", false, "If true then index the necessary tables for use with the whosonfirst/go-whosonfirst-spatial-sqlite package.")
	fs.BoolVar(&spelunker_tables, "spelunker-tables", false, "If true then index the necessary tables for use with the whosonfirst/go-whosonfirst-spelunker packages")

	fs.BoolVar(&live_hard, "live-hard-die-fast", true, "Enable various performance-related pragmas at the expense of possible (unlikely) database corruption")
	fs.BoolVar(&timings, "timings", false, "Display timings during and after indexing")
	fs.BoolVar(&optimize, "optimize", true, "Attempt to optimize the database before closing connection")

	fs.BoolVar(&alt_files, "index-alt-files", false, "Index alt geometries. This flag is deprecated, please use -index-alt=TABLE,TABLE,etc. instead. To index alt geometries in all the applicable tables use -index-alt=*")
	fs.Var(&index_alt, "index-alt", "Zero or more table names where alt geometry files should be indexed.")

	fs.BoolVar(&strict_alt_files, "strict-alt-files", true, "Be strict when indexing alt geometries")

	fs.BoolVar(&index_relations, "index-relations", false, "Index the records related to a feature, specifically wof:belongsto, wof:depicts and wof:involves. Alt files for relations are not indexed at this time.")
	fs.StringVar(&relations_uri, "index-relations-reader-uri", "", "A valid go-reader.Reader URI from which to read data for a relations candidate.")

	fs.IntVar(&procs, "processes", (runtime.NumCPU() * 2), "The number of concurrent processes to index data with")

	return fs
}
