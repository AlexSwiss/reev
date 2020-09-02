package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	//mysql driver
	"github.com/AlexSwiss/reev/pkg/protocol/grpc"
	v1 "github.com/AlexSwiss/reev/pkg/service/v1"
	_ "github.com/go-sql-driver/mysql"
)

// Config is configuration for Server
type Config struct {
	// gRPC server start parameter section
	// gRPC is TCP port to listen by gRPC server
	GRPCPort string

	// DB Datastore parameter section
	// DatastoreDBHost is host of database
	DatastoreDBHost string

	// DatatoreDBUser is username to connect to database
	DatatoreDBUser string

	// DatastoreDBPassword to connect to database
	DatastoreDBPassword string

	// DatastoreSchema is schema of database
	DatastoreDBSchema string
}

// Run server run gRPC server and HTTP gateway
func RunServer() error {
	ctx := context.Background()

	//get configuration
	var cfg Config
	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&cfg.DatastoreDBHost, "db-host", "", "Database host")
	flag.StringVar(&cfg.DatatoreDBUser, "db-user", "", "Database user")
	flag.StringVar(&cfg.DatastoreDBPassword, "db-password", "", "Database password")
	flag.StringVar(&cfg.DatastoreDBSchema, "db-schema", "", "Database schema")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", cfg.GRPCPort)
	}

	// add mySQL driver specific parameter to parse date/time
	// Drop it for another database
	param := "parseTime=true"

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		cfg.DatatoreDBUser,
		cfg.DatastoreDBPassword,
		cfg.DatastoreDBHost,
		cfg.DatastoreDBSchema,
		param)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	v1API := v1.NewPostServiceServer(db)

	return grpc.RunServer(ctx, v1API, cfg.GRPCPort)
}
