package upgrade

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	ISOSource = "iso"
	OCISource = "oci"

	ReleaseFile    = "/etc/os-release"
	ConfigFile     = "run/data/cloud-config"
	LockFile       = "/run/cos/upgrade.lock"
	CosDir         = "/run/cos"
	TimeoutSeconds = 600
)

type Upgrade struct {
	//config          *types.Config
	Source          string `json:"source"`
	UpgradeRecovery bool   `json:"upgrade_recovery"`
	Force           bool   `json:"force"`
	HostDir         string `json:"host_dir"`
}

func NewUpgrade() *Upgrade {
	return &Upgrade{
		//config: cfg,
	}
}

func (u *Upgrade) Run() error {
	src, err := validateSource(u.Source)
	if err != nil {
		return err
	}

	if src == "" {
		fmt.Errorf("unsupported source type: %s", u.Source)
	}

	return nil
}

func validateSource(source string) (string, error) {
	if source == "" {
		return "", fmt.Errorf("source is empty")
	}

	if source == "dir:/" {
		return OCISource, nil
	}

	// check if source is a iso image url via http
	if strings.Contains(source, "http") {
		_, err := url.Parse(source)
		if err != nil {
			return "", fmt.Errorf("invalid source url: %s", source)
		}
		return ISOSource, nil
	}

	return "", nil
}

func checkIsSystemStopping() bool {
	//isContainer := utils.IsRunningInContainer()
	//if isContainer {
	//	return false
	//}

	return false
}
