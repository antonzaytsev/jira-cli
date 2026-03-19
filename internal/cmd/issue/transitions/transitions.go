package transitions

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
)

const (
	helpText = `Transitions lists available transitions for an issue.`
	examples = `$ jira issue transitions ISSUE-1

# Output as JSON
$ jira issue transitions ISSUE-1 --raw`
)

// NewCmdTransitions is a transitions command.
func NewCmdTransitions() *cobra.Command {
	cmd := cobra.Command{
		Use:     "transitions ISSUE-KEY",
		Short:   "List available transitions for an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"trans"},
		Args:    cobra.MinimumNArgs(1),
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Run: transitions,
	}

	cmd.Flags().Bool("raw", false, "Output raw JSON response")

	return &cmd
}

func transitions(cmd *cobra.Command, args []string) {
	project := viper.GetString("project.key")
	key := cmdutil.GetJiraIssueKey(project, args[0])

	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	raw, err := cmd.Flags().GetBool("raw")
	cmdutil.ExitIfError(err)

	client := api.DefaultClient(debug)

	t, err := api.ProxyTransitions(client, key)
	cmdutil.ExitIfError(err)

	if raw {
		out, err := json.MarshalIndent(t, "", "  ")
		cmdutil.ExitIfError(err)
		fmt.Println(string(out))
		return
	}

	if len(t) == 0 {
		fmt.Println("No transitions available")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "ID\tNAME\tAVAILABLE\n")
	for _, tr := range t {
		fmt.Fprintf(w, "%s\t%s\t%v\n", tr.ID, tr.Name, tr.IsAvailable)
	}
	w.Flush()
}
