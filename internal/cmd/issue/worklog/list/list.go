package list

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
	helpText = `List shows worklogs of an issue.`
	examples = `$ jira issue worklog list ISSUE-1

# Output as JSON
$ jira issue worklog list ISSUE-1 --raw`
)

// NewCmdWorklogList is a worklog list command.
func NewCmdWorklogList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list ISSUE-KEY",
		Short:   "List worklogs of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"ls"},
		Args:    cobra.MinimumNArgs(1),
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Run: listWorklogs,
	}

	cmd.Flags().Bool("raw", false, "Output raw JSON response")

	return &cmd
}

func listWorklogs(cmd *cobra.Command, args []string) {
	project := viper.GetString("project.key")
	key := cmdutil.GetJiraIssueKey(project, args[0])

	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	raw, err := cmd.Flags().GetBool("raw")
	cmdutil.ExitIfError(err)

	client := api.DefaultClient(debug)

	wl, err := client.GetIssueWorklog(key)
	cmdutil.ExitIfError(err)

	if raw {
		out, err := json.MarshalIndent(wl, "", "  ")
		cmdutil.ExitIfError(err)
		fmt.Println(string(out))
		return
	}

	if len(wl.Worklogs) == 0 {
		fmt.Println("No worklogs found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "ID\tAUTHOR\tTIME SPENT\tSTARTED\tCOMMENT\n")
	for _, entry := range wl.Worklogs {
		comment := entry.Comment
		if len(comment) > 50 {
			comment = comment[:50] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			entry.ID, entry.Author.DisplayName, entry.TimeSpent, entry.Started, comment)
	}
	w.Flush()
}
