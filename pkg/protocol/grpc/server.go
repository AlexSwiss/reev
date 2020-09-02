package grpc

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	v1 "github.com/AlexSwiss/reev/pkg/api/v1"
	"google.golang.org/grpc"
)

//RunServer run grpc service to publish post service
func RunServer(ctx context.Context, v1API v1.PostServiceServer, port string) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	//register service
	server := grpc.NewServer()
	v1.RegisterPostServiceServer(server, v1API)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a C, handle it
			log.Println("shuttinh down gRPC server...")
			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	//start gRPC server
	log.Println("starting gRPC server...")
	return server.Serve(listen)
}
