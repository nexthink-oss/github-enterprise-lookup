package emu

import (
	"context"

	"github.com/shurcooL/githubv4"
)

type Member struct {
	Name  string `json:"name" yaml:"name"`
	Email string `json:"email" yaml:"email"`
}

type Enterprise struct {
	Name    string
	Members map[string]Member // key is GitHub login
}

func NewEnterprise(name string) *Enterprise {
	return &Enterprise{
		Name: name,
	}
}

func (ent *Enterprise) UpdateMembers(ctx context.Context, client *githubv4.Client) error {
	/*
		query($ent: String!, $cursor: String!) {
			enterprise(slug: $ent) {
				members(first: 100, after: $cursor) {
						nodes {
							... on EnterpriseUserAccount {
								id
								login
								name
								user {
									email
								}
							}
						}
						pageInfo {
							endCursor
							hasNextPage
						}
					}
				}
			}
		}
		{"ent": "nexthink", "cursor": null}
	*/
	type memberNode struct {
		EnterpriseUserAccount struct {
			Id    githubv4.String
			Login githubv4.String
			Name  githubv4.String
			User  struct {
				Email githubv4.String
			}
		} `graphql:"... on EnterpriseUserAccount"`
	}
	var q struct {
		Enterprise struct {
			Members struct {
				Nodes    []memberNode
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"members(first: 100, after: $cursor)"`
		} `graphql:"enterprise(slug: $ent)"`
	}
	variables := map[string]interface{}{
		"ent":    (githubv4.String)(ent.Name),
		"cursor": (*githubv4.String)(nil),
	}
	var allMemberNodes []memberNode
	for {
		err := client.Query(ctx, &q, variables)
		if err != nil {
			return err
		}
		allMemberNodes = append(allMemberNodes, q.Enterprise.Members.Nodes...)
		if !q.Enterprise.Members.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.Enterprise.Members.PageInfo.EndCursor)
	}

	ent.Members = make(map[string]Member)
	for _, m := range allMemberNodes {
		member := Member{
			Name:  string(m.EnterpriseUserAccount.Name),
			Email: string(m.EnterpriseUserAccount.User.Email),
		}
		ent.Members[string(m.EnterpriseUserAccount.Login)] = member
	}

	return nil
}
