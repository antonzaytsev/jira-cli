package list

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/adf"
)

const helpText = `List shows all comments on an issue.`

// NewCmdCommentList is a comment list command.
func NewCmdCommentList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list ISSUE-KEY",
		Short:   "List comments on an issue",
		Long:    helpText,
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
		Run:     listComments,
	}

	cmd.Flags().Bool("raw", false, "Output raw JSON response")
	cmd.Flags().Bool("plain", false, "Output plain text (no ADF rendering)")

	return &cmd
}

func listComments(cmd *cobra.Command, args []string) {
	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])

	debug, _ := cmd.Flags().GetBool("debug")
	raw, _ := cmd.Flags().GetBool("raw")

	client := api.DefaultClient(debug)

	result, err := client.ListIssueComments(issueKey)
	cmdutil.ExitIfError(err)

	if raw {
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
		return
	}

	if result.Total == 0 {
		fmt.Printf("No comments on %s\n", issueKey)
		return
	}

	for _, c := range result.Comments {
		fmt.Fprintf(os.Stdout, "# Comment %s by %s (%s)\n", c.ID, c.Author.DisplayName, cmdutil.FormatDateTimeHuman(c.Created, "2006-01-02T15:04:05.000-0700"))

		switch body := c.Body.(type) {
		case string:
			fmt.Fprintln(os.Stdout, body)
		default:
			js, err := json.Marshal(body)
			if err != nil {
				fmt.Fprintln(os.Stdout, "[unable to render comment body]")
				continue
			}
			var doc adf.ADF
			if err := json.Unmarshal(js, &doc); err != nil {
				fmt.Fprintln(os.Stdout, string(js))
				continue
			}
			md := adf.NewTranslator(&doc, adf.NewMarkdownTranslator()).Translate()
			fmt.Fprintln(os.Stdout, md)
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d comments\n", result.Total)
}
