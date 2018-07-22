package grpchelper

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

// NewConsulRegister create a new consul register
func NewConsulRegister() *ConsulRegister {
	return &ConsulRegister{
		Address:                        "127.0.0.1:8500",
		Service:                        "unknown",
		Tag:                            []string{},
		Port:                           3000,
		BalanceFactor:                  100,
		DeregisterCriticalServiceAfter: time.Duration(1) * time.Minute,
		Interval:                       time.Duration(10) * time.Second,
	}
}

// ConsulRegister consul service register
type ConsulRegister struct {
	Address                        string
	Service                        string
	Tag                            []string
	Port                           int
	BalanceFactor                  int
	DeregisterCriticalServiceAfter time.Duration
	Interval                       time.Duration
}

// Register register service
func (r *ConsulRegister) Register() error {
	config := api.DefaultConfig()
	config.Address = r.Address
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}
	agent := client.Agent()

	IP := localIP()
	reg := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%v-%v-%v", r.Service, IP, r.Port), // 服务节点的名称
		Name:    fmt.Sprintf("%v", r.Service),                   // 服务名称
		Tags:    r.Tag,                                          // tag，可以为空
		Port:    r.Port,                                         // 服务端口
		Address: IP,                                             // 服务 IP
		Meta: map[string]string{
			"balanceFactor": strconv.Itoa(r.BalanceFactor),
		},
		Check: &api.AgentServiceCheck{ // 健康检查
			Interval: r.Interval.String(),                            // 健康检查间隔
			GRPC:     fmt.Sprintf("%v:%v/%v", IP, r.Port, r.Service), // grpc 支持，执行健康检查的地址，service 会传到 Health.Check 函数中
			DeregisterCriticalServiceAfter: r.DeregisterCriticalServiceAfter.String(), // 注销时间，相当于过期时间
		},
	}

	if err := agent.ServiceRegister(reg); err != nil {
		return err
	}

	return nil
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknow"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknow"
}
