package install

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jaypipes/ghw"

	"github.com/llmos-ai/llmos/pkg/config"
	"github.com/llmos-ai/llmos/pkg/elemental"
	"github.com/llmos-ai/llmos/pkg/utils"
	"github.com/llmos-ai/llmos/pkg/utils/cmd"
	"github.com/llmos-ai/llmos/pkg/utils/log"
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
	invalidDeviceNameError = "invalid device name"
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

func (i *Installer) RunInstall() error {
	cfg := i.LLMOSConfig
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
	var configUrls []string

	// add llmos config file
	llmOSConfigFile, err := utils.SaveTemp(i.LLMOSConfig, "llmos", i.logger, true)
	if err != nil {
		return err
	}
	defer os.Remove(llmOSConfigFile)

	// add after install chroot files
	afterInstallStage, err := config.AddStageAfterInstallChroot(llmOSConfigFile, i.LLMOSConfig)
	if err != nil {
		return err
	}

	// add cos config file
	cosConfig, err := config.ConvertToCosStages(i.LLMOSConfig, *afterInstallStage)
	if err != nil {
		return err
	}

	cosConfigFile, err := utils.SaveTemp(cosConfig, "cos", i.logger, false)
	if err != nil {
		return err
	}
	defer os.Remove(cosConfigFile)
	configUrls = append(configUrls, cosConfigFile)

	// add the cosConfig file to the cloud-init config files of install
	i.LLMOSConfig.Install.ConfigURL = strings.Join(configUrls[:], ",")

	// add elemental config dir and file
	elementalConfig, err := elemental.GenerateElementalConfig(i.LLMOSConfig, rootDisk)
	if err != nil {
		return err
	}

	elementalConfigDir, elementalConfigFile, err := utils.SaveElementalConfig(elementalConfig, i.logger)
	if err != nil {
		return err
	}
	defer os.Remove(elementalConfigFile)

	// specify the elemental install config-dir
	i.LLMOSConfig.Install.ConfigDir = elementalConfigDir

	if err = i.RunInstall(); err != nil {
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
