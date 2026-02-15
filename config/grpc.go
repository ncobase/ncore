package config

import (
	"time"

	"github.com/spf13/viper"
)

type GRPC struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Host    string `yaml:"host" json:"host"`
	Port    int    `yaml:"port" json:"port"`

	// TLS Configuration
	TLSEnabled bool   `yaml:"tls_enabled" json:"tls_enabled"`
	CertFile   string `yaml:"cert_file" json:"cert_file"`
	KeyFile    string `yaml:"key_file" json:"key_file"`
	CAFile     string `yaml:"ca_file" json:"ca_file"` // For mTLS

	// Connection Configuration
	MaxConnIdle      time.Duration `yaml:"max_conn_idle" json:"max_conn_idle"`
	MaxConnAge       time.Duration `yaml:"max_conn_age" json:"max_conn_age"`
	KeepaliveTime    time.Duration `yaml:"keepalive_time" json:"keepalive_time"`
	KeepaliveTimeout time.Duration `yaml:"keepalive_timeout" json:"keepalive_timeout"`

	// Performance Configuration
	MaxConcurrentStreams uint32 `yaml:"max_concurrent_streams" json:"max_concurrent_streams"`
	MaxRecvMsgSize       int    `yaml:"max_recv_msg_size" json:"max_recv_msg_size"` // bytes
	MaxSendMsgSize       int    `yaml:"max_send_msg_size" json:"max_send_msg_size"` // bytes
}

func getGRPCConfig(v *viper.Viper) *GRPC {
	return &GRPC{
		Enabled: v.GetBool("grpc.enabled"),
		Host:    v.GetString("grpc.host"),
		Port:    v.GetInt("grpc.port"),

		// TLS Configuration
		TLSEnabled: v.GetBool("grpc.tls_enabled"),
		CertFile:   v.GetString("grpc.cert_file"),
		KeyFile:    v.GetString("grpc.key_file"),
		CAFile:     v.GetString("grpc.ca_file"),

		// Connection Configuration with defaults
		MaxConnIdle:      v.GetDuration("grpc.max_conn_idle"),
		MaxConnAge:       v.GetDuration("grpc.max_conn_age"),
		KeepaliveTime:    getDurationOrDefault(v, "grpc.keepalive_time", 2*time.Hour),
		KeepaliveTimeout: getDurationOrDefault(v, "grpc.keepalive_timeout", 20*time.Second),

		// Performance Configuration with defaults
		MaxConcurrentStreams: getUint32OrDefault(v, "grpc.max_concurrent_streams", 100),
		MaxRecvMsgSize:       getIntOrDefault(v, "grpc.max_recv_msg_size", 4*1024*1024),  // 4MB
		MaxSendMsgSize:       getIntOrDefault(v, "grpc.max_send_msg_size", 4*1024*1024),  // 4MB
	}
}
