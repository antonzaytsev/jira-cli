package webhook

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const helpText = `Webhook manages Jira webhooks (Jira Server/Data Center only).

Jira Cloud webhooks require OAuth 2.0 / Connect app authentication
and cannot be managed via API token. Use the Jira Cloud admin UI instead.`

// NewCmdWebhook is a webhook command.
func NewCmdWebhook() *cobra.Command {
	cmd := cobra.Command{
		Use:         "webhook",
		Short:       "Manage Jira webhooks (Server/DC)",
		Long:        helpText,
		Aliases:     []string{"webhooks"},
		Annotations: map[string]string{"cmd:main": "true"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newCmdWebhookList(),
		newCmdWebhookCreate(),
		newCmdWebhookDelete(),
	)

	return &cmd
}

func newCmdWebhookList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all registered webhooks",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, _ []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			client := api.DefaultClient(debug)

			hooks, err := client.ListWebhooks()
			if err != nil {
				cmdutil.Failed("Error listing webhooks: %s", err)
			}

			out, _ := json.MarshalIndent(hooks, "", "  ")
			fmt.Println(string(out))
		},
	}
}

func newCmdWebhookCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a new webhook",
		Long: `Register a new webhook on Jira Server/DC.

Example:
  jira webhook create --name "My Hook" --url "https://example.com/hook" \
    --event jira:issue_created --event jira:issue_updated --jql "project = MYPROJ"`,
		Run: func(cmd *cobra.Command, _ []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			client := api.DefaultClient(debug)

			name, _ := cmd.Flags().GetString("name")
			url, _ := cmd.Flags().GetString("url")
			events, _ := cmd.Flags().GetStringArray("event")
			jqlFilter, _ := cmd.Flags().GetString("jql")
			excludeBody, _ := cmd.Flags().GetBool("exclude-body")

			if name == "" || url == "" {
				cmdutil.Failed("Error: --name and --url are required")
			}
			if len(events) == 0 {
				cmdutil.Failed("Error: at least one --event is required")
			}

			hook, err := client.CreateWebhook(&jira.WebhookCreateRequest{
				Name:        name,
				URL:         url,
				Events:      events,
				JqlFilter:   jqlFilter,
				ExcludeBody: excludeBody,
				Enabled:     true,
			})
			if err != nil {
				cmdutil.Failed("Error creating webhook: %s", err)
			}

			out, _ := json.MarshalIndent(hook, "", "  ")
			fmt.Println(string(out))
		},
	}

	cmd.Flags().String("name", "", "Webhook name (required)")
	cmd.Flags().String("url", "", "Webhook callback URL (required)")
	cmd.Flags().StringArray("event", nil, "Jira event to subscribe to (repeatable)")
	cmd.Flags().String("jql", "", "JQL filter to restrict which issues trigger the webhook")
	cmd.Flags().Bool("exclude-body", false, "Exclude issue body from webhook payload")

	return cmd
}

func newCmdWebhookDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ID",
		Short: "Delete a webhook by ID",
		Long: `Delete a registered webhook by its numeric ID.

Get the ID from 'jira webhook list' — it is the last segment of the "self" URL.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			client := api.DefaultClient(debug)

			id := strings.TrimSpace(args[0])
			if err := client.DeleteWebhook(id); err != nil {
				cmdutil.Failed("Error deleting webhook %s: %s", id, err)
			}
			cmdutil.Success("Webhook %s deleted", id)
		},
	}
}
