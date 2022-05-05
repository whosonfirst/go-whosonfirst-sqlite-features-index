# go-whosonfirst-sqlite-features-index

Go package for indexing Who's On First features in SQLite databases using the `whosonfirst/go-whosonfirst-sqlite-index` and `whosonfirst/go-whosonfirst-sqlite-features` packages.

The `go-whosonfirst-sqlite-index` package indexes SQLite databases using table constucts defined in the `aaronland/go-sqlite` package and records defined by the `whosonfirst/go-whosonfirst-iterate/v2` package

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-sqlite-features-index.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-sqlite-features-index)

## Tools

To build binary versions of these tools run the `cli` Makefile target. For example:

```
$> make cli
go build -mod vendor -o bin/wof-sqlite-index-features cmd/wof-sqlite-index-features/main.go
```

### wof-sqlite-index-features

```
$> ./bin/wof-sqlite-index-features -h
Usage of ./bin/wof-sqlite-index-features:
  -all
    	Index all tables (except the 'search' and 'geometries' tables which you need to specify explicitly)
  -ancestors
    	Index the 'ancestors' tables
  -concordances
    	Index the 'concordances' tables
  -driver string
    	 (default "sqlite3")
  -dsn string
    	 (default ":memory:")
  -geojson
    	Index the 'geojson' table
  -geometries
    	Index the 'geometries' table (requires that libspatialite already be installed)
  -geometry
    	Index the 'geometry' table
  -index-alt-files
    	Index alt geometries
  -index-relations
    	Index the records related to a feature, specifically wof:belongsto, wof:depicts and wof:involves. Alt files for relations are not indexed at this time.
  -index-relations-reader-uri string
    	A valid go-reader.Reader URI from which to read data for a relations candidate.
  -iterator-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI. Supported emitter URI schemes are: directory://,featurecollection://,file://,filelist://,geojsonl://,git://,repo:// (default "repo://")
  -live-hard-die-fast
    	Enable various performance-related pragmas at the expense of possible (unlikely) database corruption (default true)
  -mode string
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI. Supported emitter URI schemes are: directory://,featurecollection://,file://,filelist://,geojsonl://,git://,repo://. THIS FLAG IS DEPRECATED, please use -iterator-uri instead. (default "repo://")
  -names
    	Index the 'names' table
  -optimize
    	Attempt to optimize the database before closing connection (default true)
  -processes int
    	The number of concurrent processes to index data with (default 16)
  -properties
    	Index the 'properties' table
  -rtree
    	Index the 'rtree' table
  -search
    	Index the 'search' table (using SQLite FTS4 full-text indexer)
  -spr
    	Index the 'spr' table
  -strict-alt-files
    	Be strict when indexing alt geometries (default true)
  -supersedes
    	Index the 'supersedes' table
  -timings
    	Display timings during and after indexing
```

For example:

```
$> ./bin/wof-sqlite-index-features \
	-dsn microhoods.db \
	-all \
	-iterator-uri 'repo://?include=properties.wof:placetype=microhood' \
	/usr/local/data/whosonfirst-data-admin-us
```

Or creating databases for all the Who's On First repos:

```
#!/bin/sh

for REPO in $@
do

    if [ ! -d ${REPO}/data ]
    then
	echo "${REPO} has no data directory"
	continue
    fi
    
    FNAME=`basename ${REPO}`
    echo "make db for ${FNAME}"

    if [ -f "/usr/local/data/whosonfirst-sqlite/${FNAME}.db" ]
    then
	rm /usr/local/data/whosonfirst-sqlite/${FNAME}.db
    fi

    ./bin/wof-sqlite-index-features -timings -all -dsn /usr/local/data/whosonfirst-sqlite/${FNAME}-latest.db -mode repo:// ${REPO} 

done
```    

#### Inline queries

You can also specify inline queries by appending one or more `include` or `exclude` parameters to a `emitter.Emitter` URI, where the value is a string in the format of:

```
{PATH}={REGULAR EXPRESSION}
```

Paths follow the dot notation syntax used by the [tidwall/gjson](https://github.com/tidwall/gjson) package and regular expressions are any valid [Go language regular expression](https://golang.org/pkg/regexp/). Successful path lookups will be treated as a list of candidates and each candidate's string value will be tested against the regular expression's [MatchString](https://golang.org/pkg/regexp/#Regexp.MatchString) method.

For example:

```
$> ./bin/wof-sqlite-index-features \
	-all \
	-dsn ca-region.db \
	-iterator-uri 'repo://?include=properties.wof:placetype=region' \	
	/usr/local/data/whosonfirst-data-admin-ca

$> sqlite3 ca-region.db

SQLite version 3.28.0 2019-04-15 14:49:49
Enter ".help" for usage hints.
sqlite> SELECT id,name,placetype FROM spr;
85682057|Ontario|region
85682117|British Columbia|region
85682065|New Brunswick|region
85682123|Newfoundland and Labrador|region
85682067|Northwest Territories|region
85682075|Nova Scotia|region
85682081|Prince Edward Island|region
85682085|Manitoba|region
85682091|Alberta|region
85682095|Yukon|region
85682113|Saskatchewan|region
136251273|Quebec|region
85682105|Nunavut|region
```

You can pass multiple query parameters. For example:

```
$> ./bin/wof-sqlite-index-features \
	-all \
	-dsn ca-region.db \
	-iterator-uri 'repo://?include=properties.wof:placetype=region&include=properties.wof:name=(?i)new.*' \	
	/usr/local/data/whosonfirst-data-admin-ca

$> sqlite3 ca-region-new.db

SQLite version 3.28.0 2019-04-15 14:49:49
Enter ".help" for usage hints.
sqlite> SELECT id,name,placetype FROM spr;
85682065|New Brunswick|region
85682123|Newfoundland and Labrador|region
```

The default query mode is to ensure that all queries match but you can also specify that only one or more queries need to match by appending a `include_mode` or `exclude_mode` parameter where the value is either "ANY" or "ALL".

#### SQLite performace-related PRAGMA

Note that the `-live-hard-die-fast` flag is enabled by default. That is to enable a number of performace-related PRAGMA commands (described [here](https://blog.devart.com/increasing-sqlite-performance.html) and [here](https://www.gaia-gis.it/gaia-sins/spatialite-cookbook/html/system.html)) without which database index can be prohibitive and time-consuming. These is a small but unlikely chance of database corruptions when this flag is enabled.

Also note that the `-live-hard-die-fast` flag will cause the `PAGE_SIZE` and `CACHE_SIZE` PRAGMAs to be set to `4096` and `1000000` respectively so the eventual cache size will require 4GB of memory. This is probably fine on most systems where you'll be indexing data but I am open to the idea that we may need to revisit those numbers or at least make them configurable.

## Spatial indexes

### RTree

RTree indexes are available if SQLite has been compiled with the [R*Tree module](https://www.sqlite.org/rtree.html) and you have indexed the [rtree](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#rtree), [spr](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#spr) and [properties](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#properties) tables. For example:

```
$> ./bin/wof-sqlite-index-features \
	-index-alt-files \
	-rtree \
	-spr \
	-properties \
	-timings \
	-dsn /usr/local/ca-alt.db \
	/usr/local/data/whosonfirst-data-admin-ca/
```

### Spatialite

Spatial indexes are also available if you have the [Spatialite extension](https://www.gaia-gis.it/fossil/libspatialite/index) installed and have indexed the `geometries` table. For example:

```
$> ./bin/wof-sqlite-index-features \
	-driver spatialite \
	-timings \
	-spr \
	-geometries \
	-dsn test.db /usr/local/data/whosonfirst-data-constituency-ca/

10:09:46.534281 [wof-sqlite-index-features] STATUS time to index geometries (87) : 21.251828704s
10:09:46.534379 [wof-sqlite-index-features] STATUS time to index spr (87) : 3.206930799s
10:09:46.534385 [wof-sqlite-index-features] STATUS time to index all (87) : 24.48004637s

$> sqlite3 test.db
SQLite version 3.21.0 2017-10-24 18:55:49
Enter ".help" for usage hints.

sqlite> SELECT load_extension('mod_spatialite.dylib');
sqlite> SELECT s.id, s.name FROM spr s, geometries g WHERE ST_Intersects(g.geom, GeomFromText('POINT(-122.229137 49.450129)', 4326)) AND g.id = s.id;
1108962831|Maple Ridge-Pitt Meadows
```

Or:

```
> spatialite whosonfirst-data-latest.db
SpatiaLite version ..: 4.1.1	Supported Extensions:
...spatialite chatter goes here...
SQLite version 3.8.2 2013-12-06 14:53:30
Enter ".help" for instructions
Enter SQL statements terminated with a ";

spatialite> SELECT s.id, s.name FROM spr AS s, geometries AS g1, geometries AS g2 WHERE g1.id =  85834637 AND s.placetype = 'neighbourhood' AND g2.id = s.id AND ST_Touches(g1.geom, g2.geom) AND g2.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'geometries' AND search_frame=g2.geom);
102112179|La Lengua
1108831803|Showplace Square

spatialite> SELECT s.id, s.name FROM spr AS s, geometries AS g1, geometries AS g2 WHERE g1.id != g2.id AND g1.id =  85865959 AND s.placetype = 'neighbourhood' AND s.is_current=1 AND g2.id = s.id AND (ST_Touches(g1.geom, g2.geom) OR ST_Intersects(g1.geom, g2.geom)) AND g2.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'geometries' AND search_frame=g2.geom);
1108831807|Fairmount
85814471|Diamond Heights
85869221|Eureka Valley

SELECT s.id, s.name, s.is_current FROM spr AS s, geometries AS g1, geometries AS g2 WHERE g1.id != g2.id AND g1.id =  102061079 AND s.placetype = 'neighbourhood' AND g2.id = s.id AND (ST_Touches(g1.geom, g2.geom) OR ST_Intersects(g1.geom, g2.geom)) AND g2.ROWID IN (SELECT ROWID FROM SpatialIndex WHERE f_table_name = 'geometries' AND search_frame=g2.geom);
85892915|BoCoCa|0
85869125|Boerum Hill|1
420782915|Carroll Gardens|1
85865587|Gowanus|1
```

_Remember: When indexing geometries you will need to explcitly pass both the `-geometries` and `-driver spatialite` flags, even if you are already passing in the `-all` flag. This is so `-all` will continue to work as expected for people who don't have Spatialite installed on their computer._

## Indexing 

Indexing time will vary depending on the specifics of your hardware (available RAM, CPU, disk I/O) but as a rule building indexes with the `geometries` table will take longer, and create a larger database, than doing so without. For example indexing the [whosonfirst-data](https://github.com/whosonfirst-data/whosonfirst-data) repository with spatial indexes:

```
$> ./bin/wof-sqlite-index-features \
	-driver spatialite \
	-all \
	-geometries \
	-dsn /usr/local/data/dist/sqlite/whosonfirst-data-latest.db \
	-timings \
	/usr/local/data/whosonfirst-data

...time passes...
06:12:51.274132 [wof-sqlite-index-features] STATUS time to index geojson (951541) : 13m41.994217581s
06:12:51.274158 [wof-sqlite-index-features] STATUS time to index spr (951541) : 13m0.21007633s
06:12:51.274173 [wof-sqlite-index-features] STATUS time to index names (951541) : 17m50.759093941s
06:12:51.274178 [wof-sqlite-index-features] STATUS time to index ancestors (951541) : 3m37.431723948s
06:12:51.274182 [wof-sqlite-index-features] STATUS time to index concordances (951541) : 2m36.737857568s
06:12:51.274187 [wof-sqlite-index-features] STATUS time to index geometries (951541) : 43m48.39054903s
06:12:51.274192 [wof-sqlite-index-features] STATUS time to index all (951541) : 4h41m45.492361401s

> du -h /usr/local/data/dist/sqlite/whosonfirst-data-latest.db
15G     /usr/local/data/dist/sqlite/whosonfirst-data-latest.db
```

And without:

```
$> ./bin/wof-sqlite-index-features \
	-all \
	-dsn /usr/local/data/dist/sqlite/whosonfirst-data-latest-nospatial.db \
	-timings \
	/usr/local/data/whosonfirst-data
...time passes...
10:06:13.226187 [wof-sqlite-index-features] STATUS time to index names (951541) : 12m32.359733539s
10:06:13.226206 [wof-sqlite-index-features] STATUS time to index ancestors (951541) : 3m27.294843778s
10:06:13.226212 [wof-sqlite-index-features] STATUS time to index concordances (951541) : 2m5.947968206s
10:06:13.226220 [wof-sqlite-index-features] STATUS time to index geojson (951541) : 10m11.355455209s
10:06:13.226226 [wof-sqlite-index-features] STATUS time to index spr (951541) : 11m32.687081163s
10:06:13.226233 [wof-sqlite-index-features] STATUS time to index all (951541) : 3h43m20.687783762s

> du -h /usr/local/data/dist/sqlite/whosonfirst-data-latest-nospatial.db 
12G     /usr/local/data/dist/sqlite/whosonfirst-data-latest-nospatial.db
```

As of this writing individual tables are indexed atomically. There may be some improvements to be made indexing tables in separate Go routines but my hunch is this will make SQLite sad and cause a lot of table lock errors. I don't need to be right about that, though...

## See also

* https://github.com/aaronland/go-sqlite
* https://github.com/whosonfirst/go-whosonfirst-sqlite-features
* https://github.com/whosonfirst/go-whosonfirst-sqlite-index
* https://github.com/whosonfirst/go-whosonfirst-iterate/v2