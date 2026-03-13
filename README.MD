# Go-Microservice-Template
A template for microservice development in Go. Most of the boilerplate code is already done. Setup and configuration is done with environment variables.  
Template assume your tech stack is:  
* Postgres
* Redis
* Kafka
* Gin (HTTP server)
* Datadog
* Sentry

Any of these components can be removed if not needed or replaced with others.
## Run project

```shell
cat .env.example > .env
docker-compose up -d
goose up
go run ./cmd/  server --dev
```
## STILL WORK IN PROGRESS

## App Structure
```text
.
├── cmd
│   ├── command
│   │   ├── root.go
│   │   └── server.go
│   └── main.go
├── internal
│   ├── entity
│   │   └── user.go
│   ├── http
│   │   ├── handler
│   │   │   └── probe_handler.go
│   │   ├── middleware
│   │   └── router.go
│   ├── kafka
│   │   ├── consumer
│   │   │   └── user_consumer.go
│   │   └── producer
│   │       └── user_producer.go
│   ├── repo
│   │   ├── postgres
│   │   │   └── user_repository.go
│   │   └── redis
│   └── service
│       └── user_service.go
├── migrations
│   └── 20260228094948_create_users_table.sql
├── pkg
│   ├── app
│   │   ├── app.go
│   │   ├── datadog.go
│   │   ├── gin.go
│   │   ├── kafka.go
│   │   ├── logger.go
│   │   ├── postgres.go
│   │   ├── redis.go
│   │   └── sentry.go
│   ├── config
│   │   ├── config.go
│   │   └── logger.go
│   ├── httpserver
│   │   └── server.go
│   ├── kafka
│   │   ├── consumer.go
│   │   └── producer.go
│   └── lifecycle
│       ├── component.go
│       ├── default.go
│       ├── restart.go
│       ├── root.go
│       └── supervisor.go
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── README.MD
└── treemd.sh
```

# Use as template

```shell
go install golang.org/x/tools/cmd/gonew@latest
gonew github.com/gingray/go-template example.com/myapp path-to-dir
```
`example.com/myapp` - package name of your application.  
`path-to-dir` - path to directory where you want to create application.  

One caveat when use gonew it will cache repo means you will need to manully delete cache folder to avoid confusion.
You need to run something like this
```shell
# check if you have cache
ls $(go env GOMODCACHE)/github.com/gingray

# delete cache for template package
rm -rf $(go env GOMODCACHE)/github.com/gingray

# or use go clean -modcache but it will clean all cache probably not what you want 
go clean -modcache
```
## Dependencies

```shell
go get github.com/joho/godotenv
go get github.com/caarlos0/env/v11
go get github.com/jackc/pgx/v5
go get github.com/redis/go-redis/v9
go get github.com/getsentry/sentry-go
go get gopkg.in/DataDog/dd-trace-go.v1/ddtrace
go get gopkg.in/DataDog/dd-trace-go.v1/profiler
go get github.com/gin-gonic/gin
```

```shell
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Migrations
Migrations are done with goose. Most popular migration tool for Go is go-migrate, but I found goose to be more intuitive.  
Goose keep track of all applied migrations (go-migrate doesn't only last one applied) in case if migration fails with go-migrate you need  to manually fix it 
which I think is not a good idea. Goose fix this problem. If migration fails, I need to fix it and goose try to apply it again.
example of migration:

```shell
goose create create_users_table sql
goose status
goose up
```
## Postgres
In app used pgxpool. Not database/sql, because pgxpool is more performant, but in some cases you need to use database/sql. In this case you can still use pgxpool 
but need additional setup like
```go
	// Create the pgxpool with configured settings
	pgxPool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	// Record pool statistics for observability
	err = otelpgx.RecordStats(pgxPool)
	if err != nil {
		return nil, fmt.Errorf("unable to record database stats: %w", err)
	}

	// Use stdlib connector to expose pgxpool as database/sql compatible interface
	// This maintains the pool semantics while providing sql.DB interface for GORM
	connector := stdlib.GetPoolConnector(pgxPool)
	sqlDB := sql.OpenDB(connector)

	// CRITICAL: Set MaxIdleConns to 0 as per GetPoolConnector documentation
	// The pgxpool manages all connections internally. Setting idle connections on sql.DB
	// would cause it to hold connections outside the pool, starving direct pgxpool users.
	sqlDB.SetMaxIdleConns(0)

	// Set MaxOpenConns to match pgxpool's MaxConns to prevent sql.DB from trying
	// to open more connections than the pool allows
	if poolOpts.MaxOpen > 0 {
		sqlDB.SetMaxOpenConns(poolOpts.MaxOpen)
	}

	// Connection lifetime is managed by pgxpool, but we set it on sql.DB as well
	// to ensure consistency in connection recycling behavior
	if poolOpts.MaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(poolOpts.MaxLifetime)
	}

	// Open GORM with the properly configured sql.DB backed by pgxpool
	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{
			Conn:                 sqlDB,
			PreferSimpleProtocol: poolOpts.PreferSimpleProtocol,
		}),
		&gorm.Config{
			Logger:                 datastoreLogger(ctx, poolOpts.TraceConfig),
			SkipDefaultTransaction: poolOpts.SkipDefaultTransaction,
			PrepareStmt:            true, // Enable prepared statement cache for better performance
		},
	)
```
I took this approach from here: https://github.com/pitabwire/frame/blob/v1.63.1/datastore/pool/connection.go#L46-L91
## Component system
Component system gives ability to create component based application. Each component can be dependend on other components.
It's solves problem, order of component initialization. Like you have Kafka, Postgres and you need Kafka and Postgres started
before your web application start. Component system can handle this problem.

Possible fail scenarios:
* What to do if one child component fails?
* Should app shutdown if child component fails (parent or child)?
* Shutdown all components if error occur in parent
* Should app retry if any child component fails?
* How retry logic should be implemented (shoutdown all child components, then parent and restart parent and then child components)?

All of these are possible scenarios hard to address properly. There is a lot of possible scenarios and to cover all of them is not possible. 
Solution can easy became way too complex then problem is trying to solve.  

Default approach is in case any error occurs in parent or child component, shutdown all components in proper order, from child to parent. 
Except default approach, there another strategy to handle component it's restart when component get signal it can restart full hierarchy. 
It's cover possible scenario when component like Kafka need to restart when configuration of topics is changed

## Other
Interesting projects to keep an eye on:  
https://github.com/stefanprodan/podinfo  
https://github.com/evrone/go-clean-template  