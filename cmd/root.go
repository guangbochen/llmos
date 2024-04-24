package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llmos",
		Short: "LLMOS CLI Tool",
	}
	cmd.PersistentFlags().String("config-dir", "", "Set config directory")
	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	cmd.PersistentFlags().Bool("quiet", false, "Disable output")
	cmd.PersistentFlags().String("logfile", "", "Config logfile")
	cmd.PersistentFlags().Bool("dev", false, "Enable dev mode")
	_ = viper.BindPFlag("config-dir", cmd.PersistentFlags().Lookup("config-dir"))
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("quiet", cmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("logfile", cmd.PersistentFlags().Lookup("logfile"))
	_ = viper.BindPFlag("dev", cmd.PersistentFlags().Lookup("dev"))

	//logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	//slog.SetDefault(logger)
	cmd.AddCommand(
		newInstallCmd(cmd, true),
		newServeCmd(cmd),
		newVersionCmd(cmd),
		newUpgradeCmd(cmd, true),
	)
	//logger := slog.New(slog.NewTextHandler(os.Stderr))
	//if viper.GetBool("debug") {
	//	//fmt.Print("Debug mode enabled")
	//	//slog.SetLogLoggerLevel(slog.LevelDebug)
	//}
	cmd.SilenceUsage = true
	cmd.InitDefaultHelpCmd()
	return cmd
}
