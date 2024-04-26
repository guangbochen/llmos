package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/llmos-ai/llmos/pkg/system"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llmos",
		Short: "LLMOS CLI Management Tool",
	}
	cmd.PersistentFlags().String("config-dir", system.LocalConfigs, "Set config directory")
	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	cmd.PersistentFlags().Bool("dev", false, "Enable dev mode")
	_ = viper.BindPFlag("config-dir", cmd.PersistentFlags().Lookup("config-dir"))
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("dev", cmd.PersistentFlags().Lookup("dev"))

	cmd.AddCommand(
		newInstallCmd(cmd, true),
		newVersionCmd(cmd),
		newUpgradeCmd(cmd, true),
	)
	cmd.SilenceUsage = true
	cmd.InitDefaultHelpCmd()
	return cmd
}
