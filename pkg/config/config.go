package config

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/spf13/viper"
)

type LLMOSConfig struct {
	OS      LLMOS   `json:"os,omitempty" yaml:"os,omitempty"`
	Install Install `json:"install,omitempty" yaml:"install,omitempty"`

	ConfigDir string `json:"configDir,omitempty" yaml:"configDir,omitempty"`
	Debug     bool   `json:"debug,omitempty" yaml:"debug,omitempty"`
	DevMode   bool   `json:"devMode,omitempty" yaml:"devMode,omitempty"`
}

type LLMOS struct {
	SSHAuthorizedKeys []string `json:"sshAuthorizedKeys,omitempty" yaml:"sshAuthorized,omitempty"`
	WriteFiles        []File   `json:"writeFiles,omitempty" yaml:"writeFiles,omitempty"`
	Hostname          string   `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Runcmd            []string `json:"runCmd,omitempty" yaml:"runCmd,omitempty"`
	Bootcmd           []string `json:"bootCmd,omitempty" yaml:"bootCmd,omitempty"`
	Initcmd           []string `json:"initCmd,omitempty" yaml:"initCmd,omitempty"`
	Env               []string `yaml:"env,omitempty" yaml:"env,omitempty"`

	Username             string            `json:"username,omitempty" yaml:"username,omitempty"`
	Password             string            `json:"password,omitempty" yaml:"password,omitempty"`
	ExternalIP           string            `json:"externalIP,omitempty" yaml:"externalIP,omitempty"`
	Modules              []string          `json:"modules,omitempty" yaml:"modules,omitempty"`
	Sysctls              map[string]string `json:"sysctls,omitempty" yaml:"sysctls,omitempty"`
	NTPServers           []string          `json:"ntpServers,omitempty" yaml:"ntpServers,omitempty"`
	DNSNameservers       []string          `json:"dnsNameservers,omitempty" yaml:"dnsNameservers,omitempty"`
	Environment          map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	PersistentStatePaths []string          `json:"persistentStatePaths,omitempty" yaml:"persistentStatePaths,omitempty"`
}

type File struct {
	Encoding           string `json:"encoding" yaml:"encoding"`
	Content            string `json:"content" yaml:"content"`
	Owner              string `json:"owner" yaml:"owner"`
	Path               string `json:"path" yaml:"path"`
	RawFilePermissions string `json:"permissions" yaml:"permissions"`
}

type Install struct {
	Device string `json:"device" yaml:"device" binding:"required"`
	// +optional,
	ConfigURL string `json:"configUrl,omitempty" yaml:"configUrl,omitempty"`
	// +optional
	Silent bool `json:"silent,omitempty" yaml:"silent,omitempty"`
	// +optional
	ISOUrl string `json:"isoUrl,omitempty" yaml:"isoUrl,omitempty"`
	// +optional
	SystemURI string `json:"systemUri,omitempty" yaml:"systemUri,omitempty"`
	// +optional
	PowerOff bool `json:"powerOff,omitempty" yaml:"powerOff,omitempty"`
	// +optional
	Debug bool `json:"debug,omitempty" yaml:"debug,omitempty"`
	// +optional
	TTY string `json:"tty,omitempty" yaml:"tty,omitempty"`
	// +optional
	DataDevice string `json:"dataDevice,omitempty" yaml:"dataDevice,omitempty"`
	// +optional
	Env []string `json:"env,omitempty" yaml:"env,omitempty"`
	// +optional
	Reboot bool `json:"reboot,omitempty" yaml:"reboot,omitempty"`
	// +optional
	EjectCD bool `json:"eject-cd,omitempty" yaml:"eject-cd,omitempty"`
	// +optional
	ConfigDir string `json:"configDir,omitempty" yaml:"configDir,omitempty"`
}

func NewLLMOSConfig() *LLMOSConfig {
	configDir := viper.GetString("config-dir")
	debug := viper.GetBool("debug")
	devMode := viper.GetBool("dev")
	return &LLMOSConfig{
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
