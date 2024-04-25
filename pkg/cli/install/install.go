package install

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jaypipes/ghw"
	elcnst "github.com/rancher/elemental-toolkit/pkg/constants"

	"github.com/llmos-ai/llmos/pkg/elemental"
	"github.com/llmos-ai/llmos/pkg/utils/cmd"
	"github.com/llmos-ai/llmos/pkg/utils/log"

	"github.com/llmos-ai/llmos/pkg/config"
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

	//defaultLogFilePath     = "/var/log/llmos-install.log"
	defaultLoginUser       = "llmos"
	invalidDeviceNameError = "invalid device name"
	oemTargetPath          = elcnst.OEMDir
)

type Installer struct {
	LLMOSConfig  *config.LLMOSConfig
	runner       cmd.Runner
	logger       log.Logger
	elementalCli elemental.Elemental
}

func NewInstaller(cfg *config.LLMOSConfig, logger log.Logger) *Installer {
	return &Installer{
		LLMOSConfig:  cfg,
		logger:       logger,
		elementalCli: elemental.NewElemental(),
		runner:       cmd.NewRunner(),
	}
}

func (i *Installer) RunInstall(files []string) error {
	cfg := i.LLMOSConfig
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
	for _, f := range files {
		if err := utils.CopyFile(f, fmt.Sprintf("%s/%s", oemTargetPath, f)); err != nil {
			return err
		}
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

	i.logger.Info("Installation complete")
	return nil
}

func (i *Installer) GenerateInstallConfigs(rootDisk *ghw.Disk) error {
	var files []string
	cosConfig, err := config.ConvertToCos(i.LLMOSConfig)
	if err != nil {
		return err
	}

	cosConfigFile, err := utils.SaveTemp(cosConfig, "cos", i.logger)
	if err != nil {
		return err
	}
	defer os.Remove(cosConfigFile)
	files = append(files, cosConfigFile)

	llmOSConfigFile, err := utils.SaveTemp(i.LLMOSConfig, "llmos", i.logger)
	if err != nil {
		return err
	}
	defer os.Remove(llmOSConfigFile)
	files = append(files, llmOSConfigFile)

	i.LLMOSConfig.Install.ConfigURL = cosConfigFile

	// create a tmp config file for installation
	elementalConfig, err := elemental.GenerateElementalConfig(i.LLMOSConfig, rootDisk)
	if err != nil {
		return err
	}

	elementalConfigDir, elementalConfigFile, err := utils.SaveElementalConfig(elementalConfig, i.logger)
	if err != nil {
		return err
	}
	i.LLMOSConfig.Install.ConfigDir = elementalConfigDir
	defer os.Remove(elementalConfigFile)

	if err = i.RunInstall(files); err != nil {
		return err
	}

	return nil
}

// DeactivateDevices helps to tear down LVM and MD devices on the system, if the installing device is occupied, the partitioning operation could fail later.
func (i *Installer) DeactivateDevices() error {
	slog.Info("Deactivating LVM and MD devices")
	_, err := i.runner.Run("blkdeactivate", "--lvmoptions", "wholevg,retry", "--dmoptions", "force,retry", "--errors")
	if err != nil {
		return fmt.Errorf("deactivating LVM and MD devices failed: %s", err)
	}
	return nil
}
