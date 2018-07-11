package add

import (
	"context"
	"math/rand"
	"time"

	"github.com/hatlonely/grpc-go-template/api/addapi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// NewService create a new service
func NewService(config *viper.Viper, log *logrus.Logger) *Service {
	return &Service{
		log: log,
	}
}

// Service implement
type Service struct {
	log *logrus.Logger
}

// Do implement
func (s *Service) Do(ctx context.Context, request *addapi.Request) (*addapi.Response, error) {
	// 50% 概率 sleep，模拟超时场景
	if rand.Int()%2 == 0 {
		time.Sleep(time.Duration(200) * time.Millisecond)
	}
	response := &addapi.Response{
		V: request.A + request.B,
	}
	s.log.WithField("request", request).WithField("response", response).Info()
	return response, nil
}
