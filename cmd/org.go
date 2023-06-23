package cmd

import (
	"context"

	"github.com/nexthink-oss/github-enterprise-lookup/internal/auth"
	"github.com/nexthink-oss/github-enterprise-lookup/internal/org"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var orgCmd = &cobra.Command{
	Use:        "org [flags] <organization>",
	Short:      "Lookup users by Organization",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"organization"},
	RunE:       runOrgCmd,
}

func init() {
	orgCmd.PersistentFlags().Bool("force-pat", false, "force PAT authentication (GITHUB_FORCE_PAT)")
	viper.BindPFlag("force_pat", orgCmd.PersistentFlags().Lookup("force-pat"))

	orgCmd.PersistentFlags().String("app-id", "", "GitHub App ID (GITHUB_APP_ID)")
	viper.BindPFlag("app_id", orgCmd.PersistentFlags().Lookup("app-id"))
	orgCmd.PersistentFlags().String("installation-id", "", "GitHub App Installation ID (GITHUB_APP_INSTALLATION_ID)")
	viper.BindPFlag("app_installation_id", orgCmd.PersistentFlags().Lookup("installation-id"))
	orgCmd.PersistentFlags().String("pem-file", "", "GitHub App PEM file or contents (GITHUB_APP_PEM_FILE)")
	viper.BindPFlag("app_pem_file", orgCmd.PersistentFlags().Lookup("pem-file"))

	rootCmd.AddCommand(orgCmd)
}

func runOrgCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	var err error
	if !viper.GetBool("force_pat") && viper.GetString("app_pem_file") != "" {
		client, err = auth.NewGHAClient(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to create client with GitHub App credentials")
		}
	} else {
		client, err = auth.NewTokenClient(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to create client with Personal Access Token")
		}
	}

	orgName := args[0]
	o := org.NewOrganization(orgName)
	err = o.UpdateMembers(ctx, client)
	if err != nil {
		return err
	}

	members = o.Members
	return nil
}
