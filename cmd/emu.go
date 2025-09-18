package cmd

import (
	"context"

	"github.com/nexthink-oss/github-enterprise-lookup/internal/auth"
	"github.com/nexthink-oss/github-enterprise-lookup/internal/emu"
	"github.com/spf13/cobra"
)

var emuCmd = &cobra.Command{
	Use:        "emu [flags] <enterprise>",
	Short:      "lookup users in Enterprise with Managed Users",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"enterprise"},
	RunE:       runEmuCmd,
}

func init() {
	rootCmd.AddCommand(emuCmd)
}

func runEmuCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	var err error
	client, err = auth.NewTokenClient(ctx)
	if err != nil {
		return err
	}

	enterprise := emu.NewEnterprise(args[0])
	err = enterprise.UpdateMembers(ctx, client)
	if err != nil {
		return err
	}

	members = enterprise.Members
	return nil
}
