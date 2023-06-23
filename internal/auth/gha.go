package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/beatlabs/github-auth/app/inst"
	"github.com/beatlabs/github-auth/key"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
)

func NewGHAClient(ctx context.Context) (*githubv4.Client, error) {
	if appId := viper.GetString("app_id"); appId != "" {
		instId := viper.GetString("app_installation_id")
		appPemFile := viper.GetString("app_pem_file")
		appPem := []byte(appPemFile)

		// support both embedded PEM and path-to-pem-file
		if _, err := os.Stat(appPemFile); err == nil {
			appPem, err = os.ReadFile(appPemFile)
			if err != nil {
				return nil, err
			}
		}

		appKey, err := key.Parse(appPem)
		if err != nil {
			return nil, err
		}

		conf, err := inst.NewConfig(appId, instId, appKey)
		if err != nil {
			return nil, err
		}

		httpClient := conf.Client(ctx)

		return githubv4.NewClient(httpClient), nil
	}
	return nil, fmt.Errorf("no GitHub App credentials available")
}
