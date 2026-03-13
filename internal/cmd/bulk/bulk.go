package bulk

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const helpText = `Bulk performs batch operations on multiple issues.

Reads a JSON array from stdin. Each element describes one operation.`

// NewCmdBulk is a bulk command.
func NewCmdBulk() *cobra.Command {
	cmd := cobra.Command{
		Use:         "bulk",
		Short:       "Batch operations on multiple issues",
		Long:        helpText,
		Aliases:     []string{"batch"},
		Annotations: map[string]string{"cmd:main": "true"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newCmdBulkCreate(),
		newCmdBulkEdit(),
		newCmdBulkTransition(),
	)

	return &cmd
}

func readJSONFromStdin(v interface{}) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("no input provided; pipe a JSON array to stdin")
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("invalid JSON input: %w", err)
	}
	return nil
}

// --- Bulk Create ---

type bulkCreateItem struct {
	Summary     string            `json:"summary"`
	Type        string            `json:"type"`
	Body        string            `json:"body,omitempty"`
	Priority    string            `json:"priority,omitempty"`
	Assignee    string            `json:"assignee,omitempty"`
	Labels      []string          `json:"labels,omitempty"`
	Components  []string          `json:"components,omitempty"`
	Parent      string            `json:"parent,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
	FixVersions []string          `json:"fix_versions,omitempty"`
}

func newCmdBulkCreate() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Batch create issues from JSON stdin",
		Long: `Reads a JSON array of issue objects from stdin and creates them.

Each object requires "summary" and "type" fields.

Example input:
  [
    {"summary": "Task one", "type": "Task", "priority": "High"},
    {"summary": "Bug report", "type": "Bug", "body": "Details here", "labels": ["urgent"]}
  ]`,
		Run: bulkCreate,
	}
}

func bulkCreate(cmd *cobra.Command, _ []string) {
	project := viper.GetString("project.key")
	projectType := viper.GetString("project.type")
	installation := viper.GetString("installation")

	debug, _ := cmd.Flags().GetBool("debug")
	client := api.DefaultClient(debug)

	var items []bulkCreateItem
	if err := readJSONFromStdin(&items); err != nil {
		cmdutil.Failed("Error: %s", err)
	}

	if len(items) == 0 {
		cmdutil.Failed("Error: empty array provided")
	}

	results := make([]map[string]string, 0, len(items))
	var errors []string

	for i, item := range items {
		if item.Summary == "" || item.Type == "" {
			errors = append(errors, fmt.Sprintf("item %d: summary and type are required", i))
			continue
		}

		cr := jira.CreateRequest{
			Project:        project,
			IssueType:      item.Type,
			ParentIssueKey: item.Parent,
			Summary:        item.Summary,
			Body:           item.Body,
			Priority:       item.Priority,
			Labels:         item.Labels,
			Components:     item.Components,
			FixVersions:    item.FixVersions,
			CustomFields:   item.Custom,
			EpicField:      viper.GetString("epic.link"),
		}
		cr.ForProjectType(projectType)
		cr.ForInstallationType(installation)

		resp, err := api.ProxyCreate(client, &cr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("item %d (%s): %s", i, item.Summary, err))
			continue
		}
		results = append(results, map[string]string{"key": resp.Key, "id": resp.ID, "summary": item.Summary})
	}

	output, _ := json.Marshal(map[string]interface{}{
		"created": results,
		"errors":  errors,
	})
	fmt.Println(string(output))

	if len(errors) > 0 {
		os.Exit(1)
	}
}

// --- Bulk Edit ---

type bulkEditItem struct {
	Key        string            `json:"key"`
	Summary    string            `json:"summary,omitempty"`
	Body       string            `json:"body,omitempty"`
	Priority   string            `json:"priority,omitempty"`
	Labels     []string          `json:"labels,omitempty"`
	Components []string          `json:"components,omitempty"`
	Assignee   string            `json:"assignee,omitempty"`
	Custom     map[string]string `json:"custom,omitempty"`
}

func newCmdBulkEdit() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Batch edit issues from JSON stdin",
		Long: `Reads a JSON array of edit objects from stdin and applies them.

Each object requires a "key" field.

Example input:
  [
    {"key": "ISSUE-1", "summary": "Updated title", "priority": "High"},
    {"key": "ISSUE-2", "labels": ["backend", "urgent"]}
  ]`,
		Run: bulkEdit,
	}
}

func bulkEdit(cmd *cobra.Command, _ []string) {
	debug, _ := cmd.Flags().GetBool("debug")
	client := api.DefaultClient(debug)

	var items []bulkEditItem
	if err := readJSONFromStdin(&items); err != nil {
		cmdutil.Failed("Error: %s", err)
	}

	if len(items) == 0 {
		cmdutil.Failed("Error: empty array provided")
	}

	successes := make([]string, 0, len(items))
	var errors []string

	for i, item := range items {
		if item.Key == "" {
			errors = append(errors, fmt.Sprintf("item %d: key is required", i))
			continue
		}

		edr := jira.EditRequest{
			Summary:      item.Summary,
			Priority:     item.Priority,
			Labels:       item.Labels,
			Components:   item.Components,
			CustomFields: item.Custom,
		}
		if item.Body != "" {
			edr.Body = item.Body
		}

		if err := client.Edit(item.Key, &edr); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", item.Key, err))
			continue
		}
		successes = append(successes, item.Key)
	}

	output, _ := json.Marshal(map[string]interface{}{
		"edited": successes,
		"errors": errors,
	})
	fmt.Println(string(output))

	if len(errors) > 0 {
		os.Exit(1)
	}
}

// --- Bulk Transition ---

type bulkTransitionItem struct {
	Key   string `json:"key"`
	State string `json:"state"`
}

func newCmdBulkTransition() *cobra.Command {
	return &cobra.Command{
		Use:     "transition",
		Short:   "Batch transition issues from JSON stdin",
		Aliases: []string{"move"},
		Long: `Reads a JSON array of transition objects from stdin and applies them.

Each object requires "key" and "state" fields.

Example input:
  [
    {"key": "ISSUE-1", "state": "In Progress"},
    {"key": "ISSUE-2", "state": "Done"}
  ]`,
		Run: bulkTransition,
	}
}

func bulkTransition(cmd *cobra.Command, _ []string) {
	debug, _ := cmd.Flags().GetBool("debug")
	client := api.DefaultClient(debug)

	var items []bulkTransitionItem
	if err := readJSONFromStdin(&items); err != nil {
		cmdutil.Failed("Error: %s", err)
	}

	if len(items) == 0 {
		cmdutil.Failed("Error: empty array provided")
	}

	successes := make([]string, 0, len(items))
	var errors []string

	for i, item := range items {
		if item.Key == "" || item.State == "" {
			errors = append(errors, fmt.Sprintf("item %d: key and state are required", i))
			continue
		}

		transitions, err := api.ProxyTransitions(client, item.Key)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to fetch transitions: %s", item.Key, err))
			continue
		}

		var targetTransition *jira.Transition
		for _, t := range transitions {
			if strings.EqualFold(t.Name, item.State) {
				targetTransition = t
				break
			}
		}

		if targetTransition == nil {
			available := make([]string, 0, len(transitions))
			for _, t := range transitions {
				available = append(available, t.Name)
			}
			errors = append(errors, fmt.Sprintf("%s: invalid state %q (available: %s)", item.Key, item.State, strings.Join(available, ", ")))
			continue
		}

		_, err = client.Transition(item.Key, &jira.TransitionRequest{
			Transition: &jira.TransitionRequestData{
				ID:   targetTransition.ID.String(),
				Name: targetTransition.Name,
			},
		})
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", item.Key, err))
			continue
		}
		successes = append(successes, item.Key)
	}

	output, _ := json.Marshal(map[string]interface{}{
		"transitioned": successes,
		"errors":       errors,
	})
	fmt.Println(string(output))

	if len(errors) > 0 {
		os.Exit(1)
	}
}
