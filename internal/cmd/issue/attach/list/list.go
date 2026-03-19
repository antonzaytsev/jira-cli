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
	helpText = `List shows attachments of an issue.`
	examples = `$ jira issue attach list ISSUE-1

# Output as JSON
$ jira issue attach list ISSUE-1 --raw`
)

// NewCmdAttachList is an attachment list command.
func NewCmdAttachList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list ISSUE-KEY",
		Short:   "List attachments of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"ls"},
		Args:    cobra.MinimumNArgs(1),
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Run: listAttachments,
	}

	cmd.Flags().Bool("raw", false, "Output raw JSON response")

	return &cmd
}

func listAttachments(cmd *cobra.Command, args []string) {
	project := viper.GetString("project.key")
	key := cmdutil.GetJiraIssueKey(project, args[0])

	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	raw, err := cmd.Flags().GetBool("raw")
	cmdutil.ExitIfError(err)

	client := api.DefaultClient(debug)

	issue, err := api.ProxyGetIssueRaw(client, key, "attachment")
	cmdutil.ExitIfError(err)

	if raw {
		fmt.Println(issue)
		return
	}

	var parsed struct {
		Fields struct {
			Attachment []struct {
				ID       string `json:"id"`
				Filename string `json:"filename"`
				Size     int64  `json:"size"`
				MimeType string `json:"mimeType"`
				Created  string `json:"created"`
				Author   struct {
					DisplayName string `json:"displayName"`
				} `json:"author"`
			} `json:"attachment"`
		} `json:"fields"`
	}

	if err := json.Unmarshal([]byte(issue), &parsed); err != nil {
		cmdutil.ExitIfError(fmt.Errorf("failed to parse response: %w", err))
	}

	attachments := parsed.Fields.Attachment
	if len(attachments) == 0 {
		fmt.Println("No attachments found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "ID\tFILENAME\tSIZE\tTYPE\tAUTHOR\tCREATED\n")
	for _, a := range attachments {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			a.ID, a.Filename, a.Size, a.MimeType, a.Author.DisplayName, a.Created)
	}
	w.Flush()
}
