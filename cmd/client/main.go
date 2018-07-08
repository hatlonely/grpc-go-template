package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/hatlonely/grpc-go-template/api/addapi"
	"github.com/hatlonely/grpc-go-template/pkg/grpchelper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"golang.org/x/time/rate"
)

func main() {
	conf := pflag.StringP("config", "c", "grpc-go-template/configs/client/client.json", "config filename")
	host := pflag.StringP("host", "h", "127.0.0.1:8500", "consul host address")
	pflag.Parse()

	config := viper.New()
	if *host != "" {
		config.AddRemoteProvider("consul", *host, *conf)
		config.SetConfigType("json")
		if err := config.ReadRemoteConfig(); err != nil {
			panic(err)
		}
	} else {
		fp, err := os.Open(*conf)
		if err != nil {
			panic(err)
		}
		if err := config.ReadConfig(fp); err != nil {
			panic(err)
		}
	}
	config.BindPFlags(pflag.CommandLine)

	logrus.SetFormatter(&logrus.JSONFormatter{})

	conn, err := grpchelper.NewConn(config.Sub("conn"))
	if err != nil {
		fmt.Printf("dial failed. err: [%v]\n", err)
		return
	}
	defer conn.Close()

	client := addapi.NewServiceClient(conn)
	limiter := rate.NewLimiter(rate.Every(time.Duration(800)*time.Millisecond), 1)
	hystrix.ConfigureCommand(
		"grpc-go-template",
		hystrix.CommandConfig{
			Timeout:                100,
			MaxConcurrentRequests:  2,
			RequestVolumeThreshold: 4,
			ErrorPercentThreshold:  25,
			SleepWindow:            1000,
		},
	)

	for i := 0; i < 10; i++ {
		if err := limiter.Wait(context.Background()); err != nil {
			panic(err)
		}
		request := &addapi.Request{
			A: int64(rand.Intn(1000)),
			B: int64(rand.Intn(1000)),
		}
		var response *addapi.Response
		err := hystrix.Do("addservice", func() error {
			var err error
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(50*time.Millisecond))
			defer cancel()
			response, err = client.Do(ctx, request)
			return err
		}, func(err error) error {
			fmt.Println(err)
			response = &addapi.Response{V: request.A + request.B}
			return nil
		})
		logrus.WithField("request", request).WithField("response", response).Info()
		if err != nil {
			panic(err)
		}
	}
}
