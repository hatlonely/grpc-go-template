package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/hatlonely/grpc-go-template/api/addapi"
	"github.com/hatlonely/grpc-go-template/pkg/grpchelper"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(
				grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Duration(1)*time.Millisecond)),
				grpc_retry.WithMax(3),
				grpc_retry.WithPerRetryTimeout(time.Duration(5)*time.Millisecond),
				grpc_retry.WithCodes(codes.ResourceExhausted, codes.Unavailable, codes.DeadlineExceeded),
			),
		),
		grpc.WithBalancer(grpc.RoundRobin(grpchelper.NewConsulResolver(
			"127.0.0.1:8500", "grpc.health.v1.grpc-go-template",
		))),
	)

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
