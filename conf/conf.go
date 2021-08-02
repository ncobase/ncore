package conf

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	c        *viper.Viper
	err      error
	confPath string
	// G Global config. e.g: conf.G.xxx
	G = &Config{}
)

// Config .
type Config struct {
	AppName    string
	RunMode    string
	Protocol   string
	Domain     string
	Host       string
	Port       int
	JWTSecret  string
	JWTExpTime int
	Logger     LoggerConfig
	DB         DBConfig
	Redis      RedisConfig
	Github     GithubConfig
	Facebook   FacebookConfig
	AWS        AWSConfig
	Mailgun    MailgunConfig
}

func init() {
	flag.StringVar(&confPath, "conf", "", "e.g: bin ./config.yml")
}

// Init load config
func Init() error {
	flag.Parse()

	*G, err = load(confPath)

	return err
}

func load(in string) (Config, error) {
	c = viper.New()

	// 增加执行文件所在目录
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	c.SetConfigFile(in)
	// 默认读取 config.{yaml,toml, yml,json} 等文件
	c.AddConfigPath(in)
	c.AddConfigPath("/etc/stone")
	c.AddConfigPath("$HOME/.stone")
	c.AddConfigPath("configs")
	c.AddConfigPath(filepath.Dir(ex))
	c.AddConfigPath(filepath.Dir(ex) + "/configs")

	err = c.ReadInConfig()

	return *getConfig(), err
}

func getConfig() *Config {
	return &Config{
		AppName:    c.GetString("app_name"),
		RunMode:    c.GetString("run_mode"),
		Protocol:   c.GetString("server.protocol"),
		Domain:     c.GetString("server.domain"),
		Host:       c.GetString("server.host"),
		Port:       c.GetInt("server.port"),
		JWTSecret:  c.GetString("jwt.secret"),
		JWTExpTime: c.GetInt("jwt.exp_time"),
		Logger:     *getLogConfig(),
		DB:         *getDBConfig(),
		Redis:      *getRedisConfig(),
		Github:     *getGithubConfig(),
		Facebook:   *getFacebookConfig(),
		AWS:        *getAWSConfig(),
		Mailgun:    *getMailgunConfig(),
	}
}
