package auth

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

func NewTokenClient(ctx context.Context) (*githubv4.Client, error) {
	if token := viper.GetString("token"); token != "" {
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient := oauth2.NewClient(ctx, src)

		return githubv4.NewClient(httpClient), nil
	}
	return nil, fmt.Errorf("no Personal Access Token available")
}
