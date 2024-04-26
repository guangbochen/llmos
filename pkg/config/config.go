package config

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/imdario/mergo"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const defaultVersion = "v1.0"

type LLMOSConfig struct {
	Version   string  `json:"version" yaml:"version"`
	ConfigDir string  `json:"config-dir,omitempty" yaml:"config-dir,omitempty"`
	Debug     bool    `json:"debug,omitempty" yaml:"debug,omitempty"`
	DevMode   bool    `json:"dev-mode,omitempty" yaml:"dev-mode,omitempty"`
	OS        LLMOS   `json:"os,omitempty" yaml:"os,omitempty"`
	Install   Install `json:"install,omitempty" yaml:"install,omitempty"`
}

type LLMOS struct {
	SSHAuthorizedKeys    []string          `json:"ssh-authorized-keys,omitempty" yaml:"ssh-authorized-keys,omitempty"`
	WriteFiles           []File            `json:"write-files,omitempty" yaml:"write-files,omitempty"`
	Hostname             string            `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Modules              []string          `json:"modules,omitempty" yaml:"modules,omitempty"`
	Sysctl               map[string]string `json:"sysctl,omitempty" yaml:"sysctl,omitempty"`
	Username             string            `json:"username,omitempty" yaml:"username,omitempty"`
	Password             string            `json:"password,omitempty" yaml:"password,omitempty"`
	NTPServers           []string          `json:"ntp-servers,omitempty" yaml:"ntp-servers,omitempty"`
	DNSNameservers       []string          `json:"dns-nameservers,omitempty" yaml:"dns-nameservers,omitempty"`
	Environment          map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	PersistentStatePaths []string          `json:"persistent-state-paths,omitempty" yaml:"persistent-state-paths,omitempty"`
	K3SConfig            `json:",inline,omitempty" yaml:",inline,omitempty"`
}

type K3SConfig struct {
	Token          string   `json:"token,omitempty" yaml:"token,omitempty"`
	NodeExternalIP string   `json:"node-external-ip,omitempty" yaml:"node-external-ip,omitempty"`
	NodeLabel      []string `json:"node-label,omitempty" yaml:"node-label,omitempty"`
}

type File struct {
	Encoding           string `json:"encoding" yaml:"encoding"`
	Content            string `json:"content" yaml:"content"`
	Owner              string `json:"owner" yaml:"owner"`
	Path               string `json:"path" yaml:"path"`
	RawFilePermissions string `json:"permissions" yaml:"permissions"`
}

type Install struct {
	Device     string   `json:"device" yaml:"device" binding:"required"`
	Silent     bool     `json:"silent,omitempty" yaml:"silent,omitempty"`
	ISOUrl     string   `json:"iso,omitempty" yaml:"iso,omitempty"`
	SystemURI  string   `json:"system-uri,omitempty" yaml:"system-uri,omitempty"`
	PowerOff   bool     `json:"poweroff,omitempty" yaml:"poweroff,omitempty"`
	Debug      bool     `json:"debug,omitempty" yaml:"debug,omitempty"`
	TTY        string   `json:"tty,omitempty" yaml:"tty,omitempty"`
	DataDevice string   `json:"data-device,omitempty" yaml:"data-device,omitempty"`
	Env        []string `json:"env,omitempty" yaml:"env,omitempty"`
	Reboot     bool     `json:"reboot,omitempty" yaml:"reboot,omitempty"`
	ConfigURL  string   `json:"config-url,omitempty" yaml:"config-url,omitempty"`
	ConfigDir  string   `json:"config-dir,omitempty" yaml:"config-dir,omitempty"`
}

func NewLLMOSConfig() *LLMOSConfig {
	configDir := viper.GetString("config-dir")
	debug := viper.GetBool("debug")
	devMode := viper.GetBool("dev")
	return &LLMOSConfig{
		Version:   defaultVersion,
		ConfigDir: configDir,
		Debug:     debug,
		DevMode:   devMode,
		Install: Install{
			Debug: debug,
		},
	}
}

func (c *LLMOSConfig) DeepCopy() (*LLMOSConfig, error) {
	newConf := NewLLMOSConfig()
	if err := mergo.Merge(newConf, c, mergo.WithAppendSlice); err != nil {
		return nil, fmt.Errorf("fail to create copy of %T at %p: %s", *c, c, err.Error())
	}
	return newConf, nil
}

func (i *Install) DeepCopy() (*Install, error) {
	install := &Install{}
	if err := mergo.Merge(install, i, mergo.WithAppendSlice); err != nil {
		return nil, fmt.Errorf("fail to create copy of %T at %p: %s", *i, i, err.Error())
	}
	return install, nil
}

func (l *LLMOS) DeepCopy() (*LLMOS, error) {
	llmos := &LLMOS{}
	if err := mergo.Merge(llmos, l, mergo.WithAppendSlice); err != nil {
		return nil, fmt.Errorf("fail to create copy of %T at %p: %s", *l, l, err.Error())
	}
	return llmos, nil
}

func (c *LLMOSConfig) ToCosInstallEnv() ([]string, error) {
	return ToEnv("LLMOS_", c.Install)
}

func (c *LLMOSConfig) HasDataPartition() bool {
	if c.Install.DataDevice == "" {
		return false
	}
	return true
}

func (c *LLMOSConfig) GetK3sNodeLabels() []string {
	if c.OS.NodeLabel == nil {
		return []string{}
	}
	return c.OS.NodeLabel
}

func (c *LLMOSConfig) GetK3sDisabledComponents() []string {
	return []string{
		"cloud-controller",
	}
}

func (c *LLMOSConfig) GetK3sNodeExternalIP() string {
	return c.OS.NodeExternalIP
}

func (c *LLMOSConfig) Merge(cfg *LLMOSConfig) error {
	if err := mergo.Merge(c, cfg, mergo.WithAppendSlice); err != nil {
		return err
	}
	return nil
}

func LoadLLMOSConfig(yamlBytes []byte) (*LLMOSConfig, error) {
	cfg := NewLLMOSConfig()
	if err := yaml.Unmarshal(yamlBytes, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}
	return cfg, nil
}

func ReadLLMOSConfigFile(file string) (*LLMOSConfig, error) {
	var err error
	var data []byte

	// check if is valid file
	_, err = os.Stat(file)
	if err == nil {
		data, err = GetLocalLLMOSConfig(file)
		if err != nil {
			return nil, err
		}
	}

	// check if source is a valid url
	if strings.Contains(file, "http") {
		url, err := url.Parse(file)
		if err != nil {
			slog.Debug("invalid source url", "file", file)
			return nil, err
		}
		data, err = GetURLLLMOSConfig(url.String())
		if err != nil {
			return nil, fmt.Errorf("error reading LLMOS config file from url: %s", err.Error())
		}

	}

	return LoadLLMOSConfig(data)
}

func GetLocalLLMOSConfig(path string) ([]byte, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading local LLMOS config file: %s", err.Error())
	}
	return bytes, nil
}

func GetURLLLMOSConfig(url string) ([]byte, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	resp, err := retryClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the HTTP response is a success (2xx) or success-like code (3xx)
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, fmt.Errorf("url response status code is invalid: %d", resp.StatusCode)
}
