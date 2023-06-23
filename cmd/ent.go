package cmd

import (
	"context"

	"github.com/nexthink-oss/github-enterprise-lookup/internal/auth"
	"github.com/nexthink-oss/github-enterprise-lookup/internal/ent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var entCmd = &cobra.Command{
	Use:        "ent [flags] <enterprise>",
	Short:      "lookup users in Enterprise with Managed Users",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"enterprise"},
	RunE:       runEntCmd,
}

func init() {
	entCmd.PersistentFlags().StringP("verified-email-org", "e", "", "GitHub Organization to use for Verified Email check (defaults to enterprise name)")
	viper.BindPFlag("verified_email_org", entCmd.PersistentFlags().Lookup("verified-email-org"))

	rootCmd.AddCommand(entCmd)
}

func runEntCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	var err error
	client, err = auth.NewTokenClient(ctx)
	if err != nil {
		return err
	}

	enterpriseSlug := args[0]
	verifiedEmailOrg := viper.GetString("verified_email_org")
	if verifiedEmailOrg == "" {
		verifiedEmailOrg = enterpriseSlug
	}

	enterprise := ent.NewEnterprise(enterpriseSlug, verifiedEmailOrg)
	err = enterprise.UpdateMembers(ctx, client)
	if err != nil {
		return err
	}

	members = enterprise.Members
	return nil
}
