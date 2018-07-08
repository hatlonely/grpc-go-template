package grpchelper

import (
	"context"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

type limiterOption struct {
	Interval time.Duration
	Buckets  int
}

type hystrixOption struct {
	Command                string
	Timeout                time.Duration
	MaxConcurrentRequests  int
	RequestVolumeThreshold int
	ErrorPercentThreshold  float64
	SleepWindow            time.Duration
}

// NewRPCHelper create a new RPC Helper
func NewRPCHelper(config *viper.Viper) *RPCHelper {
	limiterOpt := &limiterOption{}
	var limiter *rate.Limiter
	if sub := config.Sub("limiter"); sub != nil {
		sub.Unmarshal(limiterOpt)
		limiter = rate.NewLimiter(rate.Every(limiterOpt.Interval), limiterOpt.Buckets)
	}

	hystrixOpt := &hystrixOption{}
	if sub := config.Sub("hystrix"); sub != nil {
		sub.Unmarshal(hystrixOpt)
		hystrix.ConfigureCommand(
			hystrixOpt.Command,
			hystrix.CommandConfig{
				Timeout:                int(hystrixOpt.Timeout / time.Millisecond),
				MaxConcurrentRequests:  hystrixOpt.MaxConcurrentRequests,
				RequestVolumeThreshold: hystrixOpt.RequestVolumeThreshold,
				ErrorPercentThreshold:  int(hystrixOpt.ErrorPercentThreshold * 100.0),
				SleepWindow:            int(hystrixOpt.SleepWindow / time.Millisecond),
			},
		)
	}
	return &RPCHelper{
		limiter: limiter,
		command: hystrixOpt.Command,
	}
}

// RPCHelper helper
type RPCHelper struct {
	limiter  *rate.Limiter
	command  string
	callback RPCFunc
	fallback RPCFunc
}

// SetCallback set callback and fallback
func (h *RPCHelper) SetCallback(callback, fallback RPCFunc) {
	h.callback = callback
	h.fallback = fallback
}

// Do do request
func (h *RPCHelper) Do(request interface{}) (interface{}, error, error) {
	var response interface{}
	var err error
	var callErr error

	if h.limiter != nil {
		if err := h.limiter.Wait(context.Background()); err != nil {
			return nil, nil, err
		}
	}

	if h.fallback == nil {
		err = hystrix.Do(h.command, func() error {
			response, err = h.callback(request)
			callErr = err
			return err
		}, nil)
	} else {
		err = hystrix.Do(h.command, func() error {
			response, err = h.callback(request)
			callErr = err
			return err
		}, func(inErr error) error {
			callErr = inErr
			response, err = h.fallback(request)
			return err
		})
	}

	return response, callErr, err
}

// RPCFunc helper function
type RPCFunc func(request interface{}) (interface{}, error)
