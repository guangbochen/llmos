package cmd

import (
	"github.com/spf13/viper"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/llmos-ai/llmos/pkg/cli/upgrade"
)

func newUpgradeCmd(root *cobra.Command, checkRoot bool) *cobra.Command {
	//cfg := types.NewConfig()
	u := upgrade.NewUpgrade()
	c := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the LLMOS system",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if checkRoot {
				return CheckRoot(viper.GetBool("dev"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Running upgrade", "u:", u)
			return u.Run()
		},
	}
	c.Flags().StringVarP(&u.Source, "source", "s", "dir:/", "Set the source of the upgrade")
	c.Flags().BoolVarP(&u.UpgradeRecovery, "upgrade-recovery", "r", false, "Enable upgrade recovery")
	c.Flags().BoolVarP(&u.Force, "force", "f", false, "Force upgrade")
	c.Flags().StringVarP(&u.HostDir, "host-dir", "d", "host", "Set the host directory")
	return c
}
