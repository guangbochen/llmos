package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/llmos-ai/llmos/pkg/cli/install"
	"github.com/llmos-ai/llmos/pkg/config"
)

type InstallOptions struct {
	Source    string `json:"source"`
	Reboot    bool   `json:"reboot"`
	ConfigURL string `json:"configURL"`
	Silent    bool   `json:"silent"`
}

func newInstallCmd(_ *cobra.Command, checkRoot bool) *cobra.Command {
	opts := &InstallOptions{}
	c := &cobra.Command{
		Use:   "install",
		Short: "Install LLMOS to the target system",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := CheckSource(opts.Source); err != nil {
				return err
			}
			if checkRoot {
				return CheckRoot(viper.GetBool("dev"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := setupLogger(cmd.Context())
			cfg := setupConfig(opts)
			newInstall := install.NewInstaller(cfg, logger)
			if opts.Silent {
				if opts.ConfigURL == "" {
					return fmt.Errorf("config-url is required in silent mode")
				}
				return newInstall.RunInstall()
			}

			return newInstall.AskInstall()
		},
	}
	c.Flags().StringVarP(&opts.Source, "source", "s", "", "Source of the LLMOS installation")
	c.Flags().BoolVarP(&opts.Reboot, "reboot", "r", true, "Reboot the system after installation")
	c.Flags().StringVarP(&opts.ConfigURL, "config-url", "c", "", "URL or path of the LLMOS configuration file")
	c.Flags().BoolVarP(&opts.Silent, "silent", "S", false, "Run the installer in silent mode(without prompts)")
	return c
}

func setupConfig(opts *InstallOptions) *config.LLMOSConfig {
	cfg := config.NewLLMOSConfig()
	cfg.Install.Reboot = opts.Reboot
	cfg.Install.ConfigURL = opts.ConfigURL
	cfg.Install.Silent = opts.Silent
	cfg.Install.SystemURI = opts.Source
	return cfg
}
