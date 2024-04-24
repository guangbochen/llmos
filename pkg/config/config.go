package config

import (
	"fmt"

	"github.com/imdario/mergo"
)

type LLMOSConfig struct {
	OS      LLMOS   `json:"os,omitempty"`
	Install Install `json:"install,omitempty"`
}

type LLMOS struct {
	SSHAuthorizedKeys []string `json:"sshAuthorizedKeys,omitempty"`
	WriteFiles        []File   `json:"writeFiles,omitempty"`
	Hostname          string   `json:"hostname,omitempty"`
	Runcmd            []string `json:"runCmd,omitempty"`
	Bootcmd           []string `json:"bootCmd,omitempty"`
	Initcmd           []string `json:"initCmd,omitempty"`
	Env               []string `yaml:"env,omitempty"`

	Username             string            `json:"username,omitempty"`
	Password             string            `json:"password,omitempty"`
	ExternalIP           string            `json:"externalIP,omitempty"`
	Modules              []string          `json:"modules,omitempty"`
	Sysctls              map[string]string `json:"sysctls,omitempty"`
	NTPServers           []string          `json:"ntpServers,omitempty"`
	DNSNameservers       []string          `json:"dnsNameservers,omitempty"`
	Environment          map[string]string `json:"environment,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
	PersistentStatePaths []string          `json:"persistentStatePaths,omitempty"`
}

type File struct {
	Encoding           string `json:"encoding"`
	Content            string `json:"content"`
	Owner              string `json:"owner"`
	Path               string `json:"path"`
	RawFilePermissions string `json:"permissions"`
}

type Install struct {
	Device     string   `json:"device,omitempty"`
	ConfigURL  string   `json:"configUrl,omitempty"`
	Silent     bool     `json:"silent,omitempty"`
	ISOURL     string   `json:"isoUrl,omitempty"`
	PowerOff   bool     `json:"powerOff,omitempty"`
	NoFormat   bool     `json:"noFormat,omitempty"`
	Debug      bool     `json:"debug,omitempty"`
	TTY        string   `json:"tty,omitempty"`
	DataDevice string   `json:"dataDevice,omitempty"`
	Env        []string `json:"env,omitempty"`
	Reboot     bool     `json:"reboot,omitempty"`
	ConfigDir  string   `json:"configDir,omitempty"`
}

func NewLLMOSConfig() *LLMOSConfig {
	return &LLMOSConfig{}
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

func (c *LLMOSConfig) ToCosInstallEnv() ([]string, error) {
	return ToEnv("LLMOS_", c.Install)
}

func (c *LLMOSConfig) HasDataPartition() bool {
	if c.Install.DataDevice == "" {
		return false
	}
	return true
}

func (c *LLMOSConfig) GetNodeLabels() map[string]string {
	if c.OS.Labels == nil {
		return map[string]string{}
	}
	return c.OS.Labels
}

func (c *LLMOSConfig) GetDisabledComponents() []string {
	return []string{
		"cloud-controller",
	}
}

func (c *LLMOSConfig) GetNodeExternalIP() string {
	return c.OS.ExternalIP
}
