package field

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const helpText = `Field lets you discover Jira fields, including custom fields.

This is useful for agents and scripts that need to know which fields
are available before creating or editing issues.`

// NewCmdField is a field command.
func NewCmdField() *cobra.Command {
	cmd := cobra.Command{
		Use:         "field",
		Short:       "Discover Jira fields",
		Long:        helpText,
		Aliases:     []string{"fields"},
		Annotations: map[string]string{"cmd:main": "true"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newCmdFieldList(),
		newCmdFieldScreen(),
	)

	return &cmd
}

func newCmdFieldList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all available fields",
		Aliases: []string{"ls"},
		Long: `Lists all fields configured for the Jira instance.

By default shows all fields. Use --custom to show only custom fields,
or --search to filter by name.

Output is JSON for easy parsing by agents.`,
		Run: func(cmd *cobra.Command, _ []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			client := api.DefaultClient(debug)

			fields, err := client.GetField()
			if err != nil {
				cmdutil.Failed("Error fetching fields: %s", err)
			}

			customOnly, _ := cmd.Flags().GetBool("custom")
			search, _ := cmd.Flags().GetString("search")

			filtered := make([]*jira.Field, 0, len(fields))
			for _, f := range fields {
				if customOnly && !f.Custom {
					continue
				}
				if search != "" && !strings.Contains(strings.ToLower(f.Name), strings.ToLower(search)) &&
					!strings.Contains(strings.ToLower(f.ID), strings.ToLower(search)) {
					continue
				}
				filtered = append(filtered, f)
			}

			out, _ := json.MarshalIndent(filtered, "", "  ")
			fmt.Println(string(out))
		},
	}

	cmd.Flags().Bool("custom", false, "Show only custom fields")
	cmd.Flags().String("search", "", "Filter fields by name or ID (case-insensitive)")

	return cmd
}

func newCmdFieldScreen() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screen",
		Short: "List fields available on a project's create screen",
		Long: `Lists fields available when creating issues of a given type in a project.

This uses the Jira createmeta endpoint to discover which fields are
actually available on the create screen — including required vs optional.

Output is JSON for easy parsing by agents.

Example:
  jira field screen --project MYPROJ --type Bug`,
		Run: func(cmd *cobra.Command, _ []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			client := api.DefaultClient(debug)

			project, _ := cmd.Flags().GetString("project")
			issueType, _ := cmd.Flags().GetString("type")

			if project == "" {
				cmdutil.Failed("Error: --project is required")
			}

			meta, err := client.GetCreateMeta(&jira.CreateMetaRequest{
				Projects:       project,
				IssueTypeNames: issueType,
				Expand:         "projects.issuetypes.fields",
			})
			if err != nil {
				cmdutil.Failed("Error fetching create metadata: %s", err)
			}

			type fieldInfo struct {
				Key      string `json:"key"`
				Name     string `json:"name"`
				Required bool   `json:"required,omitempty"`
				Type     string `json:"type"`
				Items    string `json:"items,omitempty"`
			}

			type issueTypeFields struct {
				IssueType string      `json:"issue_type"`
				Fields    []fieldInfo `json:"fields"`
			}

			var result []issueTypeFields
			for _, p := range meta.Projects {
				for _, it := range p.IssueTypes {
					itf := issueTypeFields{IssueType: it.Name}
					for key, f := range it.Fields {
						itf.Fields = append(itf.Fields, fieldInfo{
							Key:   key,
							Name:  f.Name,
							Type:  f.Schema.DataType,
							Items: f.Schema.Items,
						})
					}
					result = append(result, itf)
				}
			}

			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(out))
		},
	}

	cmd.Flags().StringP("project", "p", "", "Project key (required)")
	cmd.Flags().StringP("type", "t", "", "Issue type name (optional, shows all types if omitted)")

	return cmd
}
