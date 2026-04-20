package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	RustFS   RustFSConfig   `yaml:"rustfs"`
	ARK      ARKConfig      `yaml:"ark"`
	JWT      JWTConfig      `yaml:"jwt"`
	Auth     AuthConfig     `yaml:"auth"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type RustFSConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region          string `yaml:"region"`
	UsePathStyle    bool   `yaml:"use_path_style"`
}

// ARKConfig 火山引擎 ARK API 配置
type ARKConfig struct {
	APIKey string `yaml:"api_key"` // ARK API Key
	Model  string `yaml:"model"`   // 模型名称
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireTime int    `yaml:"expire_time"` // 过期时间，单位秒
}

// AuthConfig 认证与授权配置
type AuthConfig struct {
	AdminUsernames []string `yaml:"admin_usernames"` // 管理员账号白名单
}

var AppConfig *Config

// LoadConfig 加载配置文件
// 根据 ENV 环境变量决定加载哪个配置文件
// ENV=prod 或 ENV=production 加载 config.prod.yaml
// 默认加载 config.yaml（开发环境）
func LoadConfig() *Config {
	env := os.Getenv("ENV")
	var configFile string

	switch env {
	case "prod", "production":
		configFile = "config/config.prod.yaml"
		log.Println("Loading production config...")
	default:
		configFile = "config/config.yaml"
		log.Println("Loading development config...")
	}

	// 获取配置文件绝对路径
	absPath, err := filepath.Abs(configFile)
	if err != nil {
		log.Fatalf("Failed to get config file path: %v", err)
	}

	// 读取配置文件
	data, err := os.ReadFile(absPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// 解析 YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	AppConfig = &cfg
	log.Printf("Config loaded from: %s", absPath)
	return &cfg
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
	)
}

// IsAdminUsername 判断用户名是否在管理员白名单中
func (c *Config) IsAdminUsername(username string) bool {
	if c == nil {
		return false
	}

	username = strings.TrimSpace(username)
	for _, adminUsername := range c.Auth.AdminUsernames {
		if strings.TrimSpace(adminUsername) == username {
			return true
		}
	}
	return false
}

// IsAdminUsername 判断用户名是否在当前配置的管理员白名单中
func IsAdminUsername(username string) bool {
	return AppConfig.IsAdminUsername(username)
}
