package grpchelper

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sohlich/elogrus"
	"github.com/spf13/viper"
	elastic "gopkg.in/olivere/elastic.v5"
)

type elasticSearchHookInfo struct {
	Index  string
	Client string
}

// NewLoggerGroup new logger group
func NewLoggerGroup(config *viper.Viper) (*LoggerGroup, error) {
	hooks, err := initHooks(config.Sub("hooks"))
	if err != nil {
		return nil, err
	}

	IP, err := publicIP()
	if err != nil {
		return nil, err
	}

	loggers := map[string]*logrus.Logger{}
	loggerConfig := config.Sub("logger")
	for key := range loggerConfig.AllSettings() {
		c := loggerConfig.Sub(key)
		log := logrus.New()
		if c.Sub("file") != nil {
			out, err := os.OpenFile(c.GetString("file.filename"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				return nil, err
			}
			log.Formatter = &logrus.JSONFormatter{}
			log.Out = out
		}
		if c.Sub("elasticsearch") != nil {
			info := &elasticSearchHookInfo{}
			c.Sub("elasticsearch").Unmarshal(info)

			hook, err := elogrus.NewElasticHook(hooks[info.Client].(*elastic.Client), IP.String(), logrus.InfoLevel, info.Index)
			if err != nil {
				return nil, err
			}
			log.AddHook(hook)
		}
		loggers[key] = log
	}

	return &LoggerGroup{
		loggers: loggers,
	}, nil
}

func initHooks(config *viper.Viper) (map[string]interface{}, error) {
	hooks := map[string]interface{}{}
	for key := range config.AllSettings() {
		c := config.Sub(key)
		t := c.GetString("type")
		if t == "elasticsearch" {
			client, err := elastic.NewClient(elastic.SetURL(strings.Split(c.GetString("address"), ",")...))
			if err != nil {
				return nil, err
			}
			hooks["hooks."+key] = client
		} else {
			return nil, fmt.Errorf("no such hook type [%v]", t)
		}
	}

	return hooks, nil
}

// LoggerGroup logger group
type LoggerGroup struct {
	loggers map[string]*logrus.Logger
}

// GetLogger by name
func (g *LoggerGroup) GetLogger(name string) (*logrus.Logger, error) {
	log, ok := g.loggers[name]
	if !ok {
		return nil, fmt.Errorf("no such logger name [%v]", name)
	}
	return log, nil
}

// GetLoggerWithoutError by name
func (g *LoggerGroup) GetLoggerWithoutError(name string) *logrus.Logger {
	return g.loggers[name]
}

func publicIP() (net.IP, error) {
	conn, err := net.DialTimeout("tcp", "ns1.dnspod.net:6666", 3*time.Second)
	if err != nil {
		return net.IPv4zero, err
	}
	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		return net.IPv4zero, err
	}
	return net.ParseIP(string(buf)), nil
}
