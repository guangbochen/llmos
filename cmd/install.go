package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/llmos-ai/llmos/pkg/cli/install"
)

type InstallOptions struct {
	Source string `json:"source"`
	Silent bool   `json:"silent"`
	Reboot bool   `json:"reboot"`
	Force  bool   `json:"force"`
	Url    string `json:"url"`
}

func newInstallCmd(root *cobra.Command, checkRoot bool) *cobra.Command {
	opts := &InstallOptions{}
	c := &cobra.Command{
		Use:   "install",
		Short: "Run the LLMOS installation",
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
			newInstall := install.NewInstaller(opts.Source, opts.Reboot, opts.Force, logger)
			if opts.Silent {
				if opts.Url == "" {
					return fmt.Errorf("url is required in silent mode")
				}
				return newInstall.RunInstall()
			}

			return newInstall.AskInstall()
		},
	}
	c.Flags().StringVarP(&opts.Source, "source", "s", "", "Source of the LLMOS installation")
	c.Flags().BoolVarP(&opts.Silent, "silent", "q", false, "Run installation in silent mode(without CLI interaction)")
	c.Flags().BoolVarP(&opts.Reboot, "reboot", "r", false, "Reboot the system after installation")
	c.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force installation even if the target device is not empty")
	c.Flags().StringVarP(&opts.Url, "url", "u", "", "URL of the LLMOS installation config")
	return c
}
