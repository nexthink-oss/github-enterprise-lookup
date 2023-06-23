package ent

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

type MemberOrgDetails struct {
	SSOName       string `json:"sso_name" yaml:"sso_name"`
	SSOLogin      string `json:"sso_login" yaml:"sso_login"`
	SSOEmail      string `json:"sso_email" yaml:"sso_email"`
	SSOProfileUrl string `json:"sso_profile_url" yaml:"sso_profile_url"`
}

type Member struct {
	GitHubName     string                      `json:"github_name" yaml:"github_name"`
	VerifiedEmails []githubv4.String           `json:"verified_emails" yaml:"verified_emails,flow"`
	Orgs           map[string]MemberOrgDetails `json:"orgs" yaml:"orgs"`
}

type Enterprise struct {
	Name              string
	VerifiedDomainOrg string
	Members           map[string]Member // map GitHub username to Member object
}

func NewEnterprise(name string, verifiedDomainOrg string) *Enterprise {
	return &Enterprise{
		Name:              name,
		VerifiedDomainOrg: verifiedDomainOrg,
	}
}

func (ent *Enterprise) UpdateMembers(ctx context.Context, client *githubv4.Client) error {
	/*
		query {
			enterprise(slug: $ent) {
				organizations(first: 1, after: $orgCursor) {
					nodes {
						login
						samlIdentityProvider {
							externalIdentities(first: 100, after: $userCursor) {
								nodes {
									scimIdentity {
										username
									}
									user {
										login
										organizationVerifiedDomainEmails(login: $verifiedDomainOrg)
									}
								}
								pageInfo {
									endCursor
									hasNextPage
								}
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
		{"ent": "enterprise-slug", "verifiedDomainOrg": "org-name" "orgCursor": null, "userCursor": null}
	*/
	type memberNode struct {
		ScimIdentity struct {
			Username   githubv4.String
			GivenName  githubv4.String
			FamilyName githubv4.String
			Emails     []struct {
				Value githubv4.String
			}
		}
		User struct {
			Login                            githubv4.String   // GitHub username
			Name                             githubv4.String   // GitHub display name
			OrganizationVerifiedDomainEmails []githubv4.String `graphql:"organizationVerifiedDomainEmails(login: $verifiedDomainOrg)"`
		}
	}
	var q struct {
		Enterprise struct {
			Organizations struct {
				Nodes []struct {
					Login                githubv4.String
					SamlIdentityProvider struct {
						ExternalIdentities struct {
							Nodes    []memberNode
							PageInfo struct {
								EndCursor   githubv4.String
								HasNextPage bool
							}
						} `graphql:"externalIdentities(first: 100, after: $userCursor)"`
					}
				}
				PageInfo struct {
					EndCursor   githubv4.String
					HasNextPage bool
				}
			} `graphql:"organizations(first: 1, after: $orgCursor)"`
		} `graphql:"enterprise(slug: $enterprise)"`
	}
	variables := map[string]interface{}{
		"enterprise":        githubv4.String(ent.Name),
		"verifiedDomainOrg": githubv4.String(ent.VerifiedDomainOrg),
		"orgCursor":         (*githubv4.String)(nil),
		"userCursor":        (*githubv4.String)(nil),
	}

	allOrgMembers := make(map[string][]memberNode)

Query:
	for {
		err := client.Query(ctx, &q, variables)
		if err != nil {
			return err
		}

		orgNode := q.Enterprise.Organizations.Nodes[0]
		orgName := string(orgNode.Login)

		if _, exists := allOrgMembers[orgName]; !exists {
			allOrgMembers[orgName] = orgNode.SamlIdentityProvider.ExternalIdentities.Nodes
		} else {
			allOrgMembers[orgName] = append(allOrgMembers[orgName], orgNode.SamlIdentityProvider.ExternalIdentities.Nodes...)
		}

		switch {
		case orgNode.SamlIdentityProvider.ExternalIdentities.PageInfo.HasNextPage:
			variables["userCursor"] = githubv4.NewString(orgNode.SamlIdentityProvider.ExternalIdentities.PageInfo.EndCursor)
		case q.Enterprise.Organizations.PageInfo.HasNextPage:
			variables["userCursor"] = (*githubv4.String)(nil)
			variables["orgCursor"] = githubv4.NewString(q.Enterprise.Organizations.PageInfo.EndCursor)
		default:
			break Query
		}
	}

	ent.Members = make(map[string]Member)
	for orgName, orgMemberNodes := range allOrgMembers {
		for _, m := range orgMemberNodes {
			login := string(m.User.Login)
			if login != "" {
				member := Member{
					GitHubName:     string(m.User.Name),
					VerifiedEmails: m.User.OrganizationVerifiedDomainEmails,
				}
				if _, exists := ent.Members[login]; !exists {
					member.Orgs = make(map[string]MemberOrgDetails)
				} else {
					member.Orgs = ent.Members[login].Orgs
				}
				member.Orgs[orgName] = MemberOrgDetails{
					SSOName:       fmt.Sprintf("%s %s", string(m.ScimIdentity.GivenName), string(m.ScimIdentity.FamilyName)),
					SSOLogin:      string(m.ScimIdentity.Username),
					SSOEmail:      string(m.ScimIdentity.Emails[0].Value), // all *members* provisioned by SCIM, so all have at least one email
					SSOProfileUrl: fmt.Sprintf("https://github.com/orgs/%s/people/%s/sso", orgName, m.User.Login),
				}
				ent.Members[login] = member
			}
		}
	}

	return nil
}
