package elemental

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/llmos-ai/llmos/pkg/config"
)

type Elemental interface {
	Install(config.Install) error
	//Upgrade(config.Upgrade) error
}

type elemental struct{}

func NewElemental() Elemental {
	return &elemental{}
}

func (r *elemental) Install(conf config.Install) error {
	installOpts := []string{"elemental"}

	if conf.Debug {
		installOpts = append(installOpts, "--debug")
	}

	if conf.ConfigDir != "" {
		installOpts = append(installOpts, "--config-dir", conf.ConfigDir)
	}

	installOpts = append(installOpts, "install")

	cmd := exec.Command("elemental")
	//environmentVariables := mapToInstallEnv(conf)
	//cmd.Env = append(os.Environ(), environmentVariables...)
	cmd.Args = installOpts
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	slog.Debug("running elemental install", "command", cmd.String(), "args", strings.Join(installOpts, " "))
	return cmd.Run()
}

//func mapToInstallEnv(conf config.Install) []string {
//	var variables []string
//	// See GetInstallKeyEnvMap() in https://github.com/rancher/elemental-toolkit/blob/main/pkg/constants/constants.go
//	variables = append(variables, formatEV("ELEMENTAL_INSTALL_CLOUD_INIT", strings.Join(conf.ConfigURLs[:], ",")))
//	variables = append(variables, formatEV("ELEMENTAL_INSTALL_TARGET", conf.Device))
//	variables = append(variables, formatEV("ELEMENTAL_INSTALL_SYSTEM", conf.SystemURI))
//	variables = append(variables, formatEV("ELEMENTAL_INSTALL_FIRMWARE", conf.Firmware))
//	variables = append(variables, formatEV("ELEMENTAL_INSTALL_ISO", conf.ISOURL))
//	variables = append(variables, formatEV("ELEMENTAL_INSTALL_TTY", conf.TTY))
//	variables = append(variables, formatEV("ELEMENTAL_INSTALL_DISABLE_BOOT_ENTRY", strconv.FormatBool(conf.DisableBootEntry)))
//	variables = append(variables, formatEV("ELEMENTAL_INSTALL_NO_FORMAT", strconv.FormatBool(conf.NoFormat)))
//	// See GetRunKeyEnvMap() in https://github.com/rancher/elemental-toolkit/blob/main/pkg/constants/constants.go
//	variables = append(variables, formatEV("ELEMENTAL_POWEROFF", strconv.FormatBool(conf.PowerOff)))
//	variables = append(variables, formatEV("ELEMENTAL_REBOOT", strconv.FormatBool(conf.Reboot)))
//	variables = append(variables, formatEV("ELEMENTAL_EJECT_CD", strconv.FormatBool(conf.EjectCD)))
//	variables = append(variables, formatEV("ELEMENTAL_SNAPSHOTTER_TYPE", conf.Snapshotter.Type))
//	return variables
//}

//func mapToResetEnv(conf elementalv1.Reset) []string {
//	var variables []string
//	// See GetResetKeyEnvMap() in https://github.com/rancher/elemental-toolkit/blob/main/pkg/constants/constants.go
//	variables = append(variables, formatEV("ELEMENTAL_RESET_CLOUD_INIT", strings.Join(conf.ConfigURLs[:], ",")))
//	variables = append(variables, formatEV("ELEMENTAL_RESET_SYSTEM", conf.SystemURI))
//	variables = append(variables, formatEV("ELEMENTAL_RESET_PERSISTENT", strconv.FormatBool(conf.ResetPersistent)))
//	variables = append(variables, formatEV("ELEMENTAL_RESET_OEM", strconv.FormatBool(conf.ResetOEM)))
//	variables = append(variables, formatEV("ELEMENTAL_RESET_DISABLE_BOOT_ENTRY", strconv.FormatBool(conf.DisableBootEntry)))
//	// See GetRunKeyEnvMap() in https://github.com/rancher/elemental-toolkit/blob/main/pkg/constants/constants.go
//	variables = append(variables, formatEV("ELEMENTAL_POWEROFF", strconv.FormatBool(conf.PowerOff)))
//	variables = append(variables, formatEV("ELEMENTAL_REBOOT", strconv.FormatBool(conf.Reboot)))
//	return variables
//}

func formatEV(key string, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
