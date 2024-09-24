package utils

import (
	"log"
	"os"
	"reflect"

	"gopkg.in/yaml.v2"
)

type RouteType string

const (
	TCP   = RouteType("tcp")
	UDP   = RouteType("udp")
	HTTP  = RouteType("http")
	HTTPS = RouteType("https")
)

type RouteConfig struct {
	Id      string    `yaml:"id,omitempty"`
	Enabled bool      `yaml:"enabled,omitempty"`
	Name    string    `yaml:"name"`
	Type    RouteType `yaml:"type"`
	Port    int       `yaml:"port,omitempty"`
	Machine Machine   `yaml:"machine"`
}

type Machine struct {
	Address string `yaml:"address"`
	Port    uint16 `yaml:"port"`
}

type DashboardConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type TailscaleConfig struct {
	APIKey   string `yaml:"api_key"`
	Hostname string `yaml:"hostnmae"`
}

type K8Config struct {
	Namespace    string `yaml:"namespace"`
	IngressName  string `yaml:"ingress_name"`
	ServiceName  string `yaml:"service_name"`
	IngressClass string `yaml:"ingress_class"`
}

type Config struct {
	Tailscale TailscaleConfig `yaml:"tailscale"`
	Dasboard  DashboardConfig `yaml:"dashboard"`
	K8Config  K8Config        `yaml:"kubernetes,omitempty"`
	Routes    []RouteConfig   `yaml:"routes"`
}

func LoadConfig() Config {
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) == 0 {
		configPath = "config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Unmarshal the YAML data into the Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return config
}

func Save(config Config) {
	b, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	os.WriteFile("config.yaml", b, os.ModeDir)
}

func SaveRoutes(routes []RouteConfig) {
	config := LoadConfig()
	config.Routes = routes
	Save(config)
}

func IsEmptyStruct(s interface{}) bool {
	return reflect.DeepEqual(s, reflect.Zero(reflect.TypeOf(s)).Interface())
}
