package grpchelper

import (
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type retryOption struct {
	WaitBetween time.Duration
	Retries     uint
	Timeout     time.Duration
	Codes       []codes.Code
}

type balancerOption struct {
	Consul  string
	Service string
}

// NewConn create a new grpc connection
func NewConn(config *viper.Viper) (*grpc.ClientConn, error) {
	retryOpt := &retryOption{}
	config.Sub("retry").Unmarshal(retryOpt)
	balancerOpt := &balancerOption{}
	config.Sub("balancer").Unmarshal(balancerOpt)

	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(
				grpc_retry.WithBackoff(grpc_retry.BackoffLinear(retryOpt.WaitBetween)),
				grpc_retry.WithMax(retryOpt.Retries),
				grpc_retry.WithPerRetryTimeout(retryOpt.Timeout),
				grpc_retry.WithCodes(retryOpt.Codes...),
			),
		),
		grpc.WithBalancer(grpc.RoundRobin(NewConsulResolver(
			balancerOpt.Consul, balancerOpt.Service,
		))),
	)

	return conn, err
}
