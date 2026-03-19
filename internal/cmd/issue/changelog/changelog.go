package changelog

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
	helpText = `Changelog shows the field-level change history of an issue.`
	examples = `$ jira issue changelog ISSUE-1

# Output as JSON
$ jira issue changelog ISSUE-1 --raw`
)

// NewCmdChangelog is a changelog command.
func NewCmdChangelog() *cobra.Command {
	cmd := cobra.Command{
		Use:     "changelog ISSUE-KEY",
		Short:   "Show issue changelog",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"changes", "history"},
		Args:    cobra.MinimumNArgs(1),
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Run: changelogRun,
	}

	cmd.Flags().Bool("raw", false, "Output raw JSON response")

	return &cmd
}

func changelogRun(cmd *cobra.Command, args []string) {
	project := viper.GetString("project.key")
	key := cmdutil.GetJiraIssueKey(project, args[0])

	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	raw, err := cmd.Flags().GetBool("raw")
	cmdutil.ExitIfError(err)

	client := api.DefaultClient(debug)

	cl, err := client.GetIssueChangelog(key)
	cmdutil.ExitIfError(err)

	if raw {
		out, err := json.MarshalIndent(cl, "", "  ")
		cmdutil.ExitIfError(err)
		fmt.Println(string(out))
		return
	}

	if len(cl.Histories) == 0 {
		fmt.Println("No changelog entries")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "DATE\tAUTHOR\tFIELD\tFROM\tTO\n")
	for _, h := range cl.Histories {
		for _, item := range h.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				h.Created, h.Author.DisplayName, item.Field, item.FromString, item.ToString)
		}
	}
	w.Flush()
}
