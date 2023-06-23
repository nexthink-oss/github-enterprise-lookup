package org

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
)

type Member struct {
	SSOName        string            `json:"sso_name" yaml:"sso_name"`
	SSOLogin       string            `json:"sso_login" yaml:"sso_login"`
	SSOEmail       string            `json:"sso_email" yaml:"sso_email"`
	GitHubName     string            `json:"github_name" yaml:"github_name"`
	VerifiedEmails []githubv4.String `json:"verified_emails" yaml:"verified_emails,flow"`
	SSOProfileUrl  string            `json:"sso_profile_url" yaml:"sso_profile_url"`
	OrgAdmin       bool              `json:"org_admin,omitempty" yaml:"org_admin,omitempty"`
}

type Organization struct {
	Name    string
	Members map[string]Member // key is GitHub login
}

func NewOrganization(name string) *Organization {
	return &Organization{
		Name: name,
	}
}

func (org *Organization) UpdateMembers(ctx context.Context, client *githubv4.Client) error {
	/*
		query($org: String!) {
			organization(login: $org) {
				samlIdentityProvider {
					externalIdentities(first: 100) {
						nodes {
							user {
								login
								name // GitHub display name
								organizationVerifiedDomainEmails(login: $org)
							}
							scimIdentity {
								givenName
								familyName
								username
								emails {
									value
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
		{"org": "organization-name", "cursor": null}
	*/
	type memberNode struct {
		User struct {
			Login                            githubv4.String
			Name                             githubv4.String   // GitHub display name
			OrganizationVerifiedDomainEmails []githubv4.String `graphql:"organizationVerifiedDomainEmails(login: $org)"`
		}
		ScimIdentity struct {
			Username   githubv4.String // SAML/SCIM NameID; requires 'admin:org' scope
			GivenName  githubv4.String
			FamilyName githubv4.String
			Emails     []struct {
				Value githubv4.String
			}
		}
	}
	var q struct {
		Organization struct {
			SamlIdentityProvider struct {
				ExternalIdentities struct {
					Nodes    []memberNode
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"externalIdentities(first: 100, after: $cursor)"`
			}
		} `graphql:"organization(login:$org)"`
	}
	variables := map[string]interface{}{
		"org":    (githubv4.String)(org.Name),
		"cursor": (*githubv4.String)(nil),
	}
	var allMemberNodes []memberNode
	for {
		err := client.Query(ctx, &q, variables)
		if err != nil {
			return err
		}
		allMemberNodes = append(allMemberNodes, q.Organization.SamlIdentityProvider.ExternalIdentities.Nodes...)
		if !q.Organization.SamlIdentityProvider.ExternalIdentities.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(q.Organization.SamlIdentityProvider.ExternalIdentities.PageInfo.EndCursor)
	}

	org.Members = make(map[string]Member)
	for _, m := range allMemberNodes {
		login := string(m.User.Login)
		if login != "" {
			member := Member{
				SSOName:        fmt.Sprintf("%s %s", string(m.ScimIdentity.GivenName), string(m.ScimIdentity.FamilyName)),
				SSOLogin:       string(m.ScimIdentity.Username),
				GitHubName:     string(m.User.Name),
				SSOEmail:       string(m.ScimIdentity.Emails[0].Value), // all *members* provisioned by SCIM, so all have at least one email
				VerifiedEmails: m.User.OrganizationVerifiedDomainEmails,
				SSOProfileUrl:  fmt.Sprintf("https://github.com/orgs/%s/people/%s/sso", org.Name, m.User.Login),
			}
			org.Members[login] = member
		}
	}

	if !viper.GetBool("no-org-admin") {
		type memberRoleNode struct {
			Role githubv4.String
			Node struct {
				Login githubv4.String
			}
		}

		var roleQuery struct {
			Organization struct {
				MembersWithRole struct {
					Edges    []memberRoleNode
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"membersWithRole(first: 100, after: $cursor)"`
			} `graphql:"organization(login:$org)"`
		}
		roleVariables := map[string]interface{}{
			"org":    (githubv4.String)(org.Name),
			"cursor": (*githubv4.String)(nil),
		}

		var allMemberRoleNodes []memberRoleNode
		for {
			err := client.Query(ctx, &roleQuery, roleVariables)
			if err != nil {
				return err
			}
			allMemberRoleNodes = append(allMemberRoleNodes, roleQuery.Organization.MembersWithRole.Edges...)
			if !roleQuery.Organization.MembersWithRole.PageInfo.HasNextPage {
				break
			}
			roleVariables["cursor"] = githubv4.NewString(roleQuery.Organization.MembersWithRole.PageInfo.EndCursor)
		}

		memberRoles := make(map[string]string)
		for _, m := range allMemberRoleNodes {
			login := string(m.Node.Login)
			if login != "" {
				memberRoles[login] = string(m.Role)
			}
		}

		for login, m := range org.Members {
			m.OrgAdmin = (memberRoles[login] == "ADMIN")
			org.Members[login] = m
		}
	}

	return nil
}
