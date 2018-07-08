package main

import (
	"fmt"
	"net"

	"github.com/hatlonely/grpc-go-template/api/addapi"
	"github.com/hatlonely/grpc-go-template/api/echoapi"
	"github.com/hatlonely/grpc-go-template/internal/add"
	"github.com/hatlonely/grpc-go-template/internal/echo"
	"github.com/hatlonely/grpc-go-template/internal/health"
	"github.com/hatlonely/grpc-go-template/pkg/grpchelper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	_ "github.com/spf13/viper/remote"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	conf := pflag.StringP("config", "c", "grpc-go-template/configs/server/server.json", "config filename")
	host := pflag.StringP("host", "h", "127.0.0.1:8500", "consul host address")
	port := pflag.IntP("register.port", "p", 3000, "service port")
	pflag.Parse()

	logrus.SetFormatter(&logrus.JSONFormatter{})

	// read configs from consul or local
	config, err := grpchelper.NewConfig(*host, *conf)
	if err != nil {
		panic(err)
	}
	config.BindPFlags(pflag.CommandLine)

	// register service to consul
	register := grpchelper.NewConsulRegister()
	config.Sub("register").Unmarshal(register)
	register.Port = config.GetInt("register.port")
	if err := register.Register(); err != nil {
		panic(err)
	}

	// create server
	server := grpc.NewServer()
	addapi.RegisterServiceServer(server, &add.Service{})
	echoapi.RegisterServiceServer(server, &echo.Service{})
	grpc_health_v1.RegisterHealthServer(server, &health.Service{})

	address, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", *port))
	if err != nil {
		panic(err)
	}

	if err := server.Serve(address); err != nil {
		panic(err)
	}
}
