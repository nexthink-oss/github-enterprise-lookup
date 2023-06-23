# github-enterprise-lookup

Map GitHub Enterprise Cloud "Bring-Your-Own-Account" users to their corporate SSO SAML/SCIM identities.

## Authentication and Permissions

Both Personal Access Token (via `GITHUB_TOKEN`) and GitHub App (via `GITHUB_APP_ID`/`GITHUB_APP_INSTALLATION_ID`/`GITHUB_APP_PEM_FILE`) authentication are supported.

Minimal required permissions: `read:org`

## Usage

```console
$ github-enterprise-lookup org --help
Lookup users by Organization

Usage:
  github-enterprise-lookup org [flags] <organization>

Flags:
      --app-id string            GitHub App ID (GITHUB_APP_ID)
      --force-pat                force PAT authentication (GITHUB_FORCE_PAT)
  -h, --help                     help for org
      --installation-id string   GitHub App Installation ID (GITHUB_APP_INSTALLATION_ID)
      --pem-file string          GitHub App PEM file or contents (GITHUB_APP_PEM_FILE)

Global Flags:
      --debug           enable debug output
  -f, --format string   output format (default "yaml")
      --no-org-admin    skip organization admin lookup
  -o, --output string   output file (default "-")
      --token string    GitHub Personal Access Token (GITHUB_TOKEN)
```

## Example Output

```yaml
roger-rabbit:
  sso_name: Roger Rabbit
  sso_login: rrabbit
  sso_email: roger.rabbit@example.com
  github_name: Roger Rabbit, Esq.
  verified_emails: [roger.rabbit@example.com]
  sso_profile_url: https://github.com/orgs/example-org/people/roger-rabbit/sso
  org_admin: true
rabbitjess:
  sso_name: Jessica Rabbit
  sso_login: jrabbit
  sso_email: jessica.rabbit@example.com
  github_name: Jess Rabbit
  verified_emails: []
  sso_profile_url:  https://github.com/orgs/example-org/people/rabbitjess/sso
```
