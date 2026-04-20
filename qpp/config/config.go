package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Browser    BrowserConfig    `yaml:"browser"`
	Login      LoginConfig      `yaml:"login"`
	Storage    StorageConfig    `yaml:"storage"`
	Automation AutomationConfig `yaml:"automation"`
}

type BrowserConfig struct {
	Headless bool           `yaml:"headless"`
	Timeout  int            `yaml:"timeout"`
	Viewport ViewportConfig `yaml:"viewport"`
}

type ViewportConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type LoginConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	WaitTime int    `yaml:"wait_time"`
}

type StorageConfig struct {
	StatePath      string `yaml:"state_path"`
	ScreenshotPath string `yaml:"screenshot_path"`
}

type AutomationConfig struct {
	RetryCount        int `yaml:"retry_count"`
	RetryDelay        int `yaml:"retry_delay"`
	NavigationTimeout int `yaml:"navigation_timeout"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		Browser: BrowserConfig{
			Headless: false,
			Timeout:  30,
			Viewport: ViewportConfig{
				Width:  1920,
				Height: 1080,
			},
		},
		Login: LoginConfig{
			URL:      "https://www.dianxiaomi.com/home.htm",
			Username: "",
			Password: "",
			WaitTime: 30,
		},
		Storage: StorageConfig{
			StatePath:      "./storage/state.json",
			ScreenshotPath: "./screenshots",
		},
		Automation: AutomationConfig{
			RetryCount:        3,
			RetryDelay:        5,
			NavigationTimeout: 30,
		},
	}
}
