package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Encoder interface {
	Encode(any) error
}

var client *githubv4.Client
var encoder Encoder
var members any
var output io.Writer

var rootCmd = &cobra.Command{
	Use:                "github-enterprise-lookup",
	Short:              "Map GitHub users to their corporate identities",
	PersistentPreRunE:  prepareOutputEncoder,
	PersistentPostRunE: encodeMembers,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initViper)

	rootCmd.PersistentFlags().Bool("debug", false, "enable debug output")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	rootCmd.PersistentFlags().String("token", "", "GitHub Personal Access Token (GITHUB_TOKEN)")
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))

	rootCmd.PersistentFlags().Bool("no-org-admin", false, "skip organization admin lookup")
	viper.BindPFlag("no-org-admin", rootCmd.PersistentFlags().Lookup("no-org-admin"))

	rootCmd.PersistentFlags().StringP("output", "o", "-", "output file")
	viper.BindPFlag("output-file", rootCmd.PersistentFlags().Lookup("output"))

	rootCmd.PersistentFlags().StringP("format", "f", "yaml", "output format")
	viper.BindPFlag("output-format", rootCmd.PersistentFlags().Lookup("format"))
}

// initViper initializes Viper to load config from the environment
func initViper() {
	viper.SetEnvPrefix("GITHUB") // match variables expected by terraform-provider-github
	viper.AutomaticEnv()         // read in environment variables that match bound variables
}

func prepareOutputEncoder(cmd *cobra.Command, args []string) (err error) {
	outputFile := viper.GetString("output-file")
	switch outputFile {
	case "-":
		output = os.Stdout
	default:
		output, err = os.OpenFile(outputFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
	}

	outputFormat := viper.GetString("output-format")
	switch outputFormat {
	case "json":
		encoder = json.NewEncoder(output)
	case "yaml":
		yamlEncoder := yaml.NewEncoder(output)
		yamlEncoder.SetIndent(2)
		encoder = yamlEncoder
	default:
		return fmt.Errorf("invalid output format")
	}
	return nil
}

func encodeMembers(cmd *cobra.Command, args []string) error {
	return encoder.Encode(members)
}
