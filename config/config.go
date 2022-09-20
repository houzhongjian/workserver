package config

type Config struct {
	AccessLog string `yaml:"access_log"`
	ErrorLog  string `yaml:"error_log"`
	Port      int    `yaml:"port"`
	Email     string `yaml:"email"`
	CertsDir  string `yaml:"certs_dir"`

	FileServer   []FileServerConfig   `yaml:"file_server"`
	ReverseProxy []ReverseProxyConfig `yaml:"reverse_proxy"`
}

//FileServerConfig 静态文件配置.
type FileServerConfig struct {
	Type       ConfigType `yaml:"type"`
	ServerName string     `yaml:"server_name"`
	Index      string     `yaml:"index"`
	Root       string     `yaml:"root"`
}

//ReverseProxyConfig 反向代理配置.
type ReverseProxyConfig struct {
	Type       ConfigType `yaml:"type"`
	ServerName string     `yaml:"server_name"`
	ProxyPass  string     `yaml:"proxy_pass"`
}

type ConfigType string

const (
	ConfigType_FileServer   ConfigType = "file_server"   //静态文件服务.
	ConfigType_ReverseProxy ConfigType = "reverse_proxy" //反向代理服务.
)

var ConfigMgr *Config
