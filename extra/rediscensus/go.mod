module github.com/wgqi1126/go-redis/extra/rediscensus/v9

go 1.15

replace github.com/wgqi1126/go-redis/v9 => ../..

replace github.com/wgqi1126/go-redis/extra/rediscmd/v9 => ../rediscmd

require (
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/wgqi1126/go-redis/extra/rediscmd/v9 v9.0.2
	github.com/wgqi1126/go-redis/v9 v9.0.2
	go.opencensus.io v0.24.0
)
