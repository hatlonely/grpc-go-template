package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/hatlonely/grpc-go-template/api/addapi"
	"github.com/hatlonely/grpc-go-template/pkg/grpchelper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	_ "github.com/spf13/viper/remote"
)

func main() {
	conf := pflag.StringP("config", "c", "grpc-go-template/configs/client/client.json", "config filename")
	host := pflag.StringP("host", "h", "127.0.0.1:8500", "consul host address")
	pflag.Parse()

	// read configs from consul or local
	config, err := grpchelper.NewConfig(*host, *conf)
	if err != nil {
		panic(err)
	}
	config.BindPFlags(pflag.CommandLine)

	// logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	// create connection
	conn, err := grpchelper.NewConn(config.Sub("conn"))
	if err != nil {
		fmt.Printf("dial failed. err: [%v]\n", err)
		return
	}
	defer conn.Close()

	client := addapi.NewServiceClient(conn)
	time.Sleep(time.Millisecond * 200)

	// init helper
	addHelper := grpchelper.NewRPCHelper(config.Sub("helper"))

	addHelper.SetCallback(
		func(requestI interface{}) (interface{}, error) {
			return client.Do(context.Background(), requestI.(*addapi.Request))
		},
		func(requestI interface{}) (interface{}, error) {
			request := requestI.(*addapi.Request)
			return &addapi.Response{
				V: request.A + request.B,
			}, nil
		},
	)

	for i := 0; i < 100; i++ {
		request := &addapi.Request{
			A: int64(rand.Intn(1000)),
			B: int64(rand.Intn(1000)),
		}

		response, info, err := addHelper.Do(request)
		logrus.WithFields(logrus.Fields{
			"request":  request,
			"response": response,
			"err":      err,
			"costUs":   int(info.Cost / time.Microsecond),
			"callErr":  info.CallErr,
		}).Info()
	}
}
