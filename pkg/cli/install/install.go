package install

import (
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/llmos-ai/llmos/pkg/elemental"
	"github.com/llmos-ai/llmos/pkg/utils/cmd"

	elcnst "github.com/rancher/elemental-toolkit/pkg/constants"

	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/log"
	"github.com/llmos-ai/llmos/pkg/utils"
)

const (
	_ = 1 << (10 * iota)
	KiB
	MiB
	GiB
	TiB
)

const (
	emptyPlaceHolder = "Unset"
	yesOrNo          = "[Y]es/[n]o"

	defaultLoginUser       = "llmos"
	defaultLogFilePath     = "/var/log/llmos-install.log"
	invalidDeviceNameError = "invalid device name"
	oemTargetPath          = elcnst.OEMDir
)

type configFiles struct {
	elementalConfigDir  string
	elementalConfigFile string
	cosConfigFile       string
	llmOSConfigFile     string
}

type Installer struct {
	Source string `json:"source"`
	Reboot bool   `json:"reboot"`
	Force  bool   `json:"force"`

	LLMOSConfig  *config.LLMOSConfig
	cfs          configFiles
	runner       cmd.Runner
	logger       log.Logger
	elementalCli elemental.Elemental
}

func NewInstaller(source string, reboot, force bool, logger log.Logger) *Installer {
	return &Installer{
		Source:       source,
		Reboot:       reboot,
		Force:        force,
		LLMOSConfig:  config.NewLLMOSConfig(),
		runner:       cmd.NewRunner(),
		logger:       logger,
		elementalCli: elemental.NewElemental(),
	}
}

func (i *Installer) RunInstall() error {
	cfg := i.LLMOSConfig
	//cfs := i.cfs
	utils.SetEnv(cfg.OS.Env)
	utils.SetEnv(cfg.Install.Env)

	if cfg.Install.Device == "" || cfg.Install.Device == "auto" {
		cfg.Install.Device = detectInstallationDevice()
	}

	if cfg.Install.Device == "" {
		return fmt.Errorf("no device found to install LLMOS")
	}

	if err := i.runInstall(); err != nil {
		return err
	}

	// copy template file to the /oem config directory
	// Note: don't use yaml file extension for the config files as it will be applied again
	if err := utils.CopyFile(i.cfs.llmOSConfigFile, oemTargetPath+"/llmos.config"); err != nil {
		return err
	}

	if err := utils.CopyFile(i.cfs.elementalConfigFile, oemTargetPath+"/elemental.config"); err != nil {
		return err
	}

	return nil
}

func (i *Installer) runInstall() error {
	i.logger.Info("Running install")
	if err := Sanitize(i.LLMOSConfig.Install); err != nil {
		return err
	}

	if err := i.DeactivateDevices(); err != nil {
		return err
	}

	if err := i.elementalCli.Install(i.LLMOSConfig.Install); err != nil {
		return err
	}

	// run the elemental install
	//args := []string{
	//	"install", "--config-dir", cfs.elementalConfigDir,
	//	"--debug",
	//}
	//cmd := exec.Command("elemental", args...)
	//var stdBuffer bytes.Buffer
	//mw := io.MultiWriter(os.Stdout, &stdBuffer)
	//
	//cmd.Stdout = mw
	//cmd.Stderr = mw
	//
	//// Execute the command
	//if err := cmd.Run(); err != nil {
	//	return fmt.Errorf("elemental install failed: %s", err)
	//}
	//slog.Info(stdBuffer.String())

	i.logger.Info("Installation complete")
	return nil
}

// DeactivateDevices helps to tear down LVM and MD devices on the system, if the installing device is occupied, the partitioning operation could fail later.
func (i *Installer) DeactivateDevices() error {
	slog.Info("Deactivating LVM and MD devices")
	cmd := exec.Command("blkdeactivate", "--lvmoptions", "wholevg,retry",
		"--dmoptions", "force,retry", "--errors")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deactivating LVM and MD devices failed: %s", err)
	}
	return nil
}
