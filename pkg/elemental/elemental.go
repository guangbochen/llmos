package elemental

import (
	"log/slog"
	"os"
	"os/exec"

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
	var installOpts []string

	if conf.Debug {
		installOpts = append(installOpts, "--debug")
	}

	if conf.ConfigDir != "" {
		installOpts = append(installOpts, "--config-dir", conf.ConfigDir)
	}

	installOpts = append(installOpts, "install")

	cmd := exec.Command("elemental")
	cmd.Args = installOpts
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	slog.Info("Running elemental install", "command", cmd.String())
	return cmd.Run()
}
