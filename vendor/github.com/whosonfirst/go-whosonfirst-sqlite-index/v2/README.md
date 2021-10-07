# go-whosonfirst-sqlite-index

Go package for creating SQLite databases.

## Important

Documentation for this package is incomplete and will be updated shortly.

## Tools

```
$> make cli
go build -mod vendor -o bin/example cmd/example/main.go
```

### example

```
> ./bin/example -h
Usage of ./bin/example:
  -driver string
    	 (default "sqlite3")
  -dsn string
    	 (default ":memory:")
  -emitter-uri string
    	A valid whosonfirst/go-whosonfirst-iterate/emitter URI. Valid schemes are: directory://,featurecollection://,file://,filelist://,geojsonl://,repo://. (default "repo://")
  -live-hard-die-fast
    	Enable various performance-related pragmas at the expense of possible (unlikely) database corruption (default true)
  -post-index
    	Enable post indexing callback function
  -timings
    	Display timings during and after indexing
```

For example:

```
$> ./bin/example -dsn test.db /usr/local/data/sfomuseum-data-architecture/
2021/02/18 11:34:58 time to index paths (1) 403.514656ms

$> sqlite3  test.db 
SQLite version 3.28.0 2019-04-15 14:49:49
Enter ".help" for usage hints.
sqlite> .tables
example

sqlite> SELECT COUNT(id) FROM example;
12751
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-sqlite
* https://github.com/whosonfirst/go-whosonfirst-iterate
