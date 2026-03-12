package edit

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/adf"
)

const helpText = `Edit updates an existing comment on an issue.`

// NewCmdCommentEdit is a comment edit command.
func NewCmdCommentEdit() *cobra.Command {
	cmd := cobra.Command{
		Use:   "edit ISSUE-KEY COMMENT-ID",
		Short: "Edit a comment on an issue",
		Long:  helpText,
		Args:  cobra.ExactArgs(2),
		Run:   edit,
	}

	cmd.Flags().StringP("body", "b", "", "New comment body (plain text)")
	cmd.Flags().String("body-adf", "", "New comment body in raw ADF JSON format")

	return &cmd
}

func edit(cmd *cobra.Command, args []string) {
	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])
	commentID := args[1]

	bodyStr, _ := cmd.Flags().GetString("body")
	bodyADF, _ := cmd.Flags().GetString("body-adf")

	if bodyStr == "" && bodyADF == "" {
		cmdutil.Failed("Either --body or --body-adf is required")
	}

	debug, _ := cmd.Flags().GetBool("debug")
	client := api.DefaultClient(debug)

	err := func() error {
		s := cmdutil.Info("Updating comment")
		defer s.Stop()

		if bodyADF != "" {
			var adfDoc adf.ADF
			if err := json.Unmarshal([]byte(bodyADF), &adfDoc); err != nil {
				cmdutil.Failed("Error: invalid ADF JSON: %s", err)
			}
			return client.EditIssueComment(issueKey, commentID, &adfDoc)
		}
		return client.EditIssueComment(issueKey, commentID, bodyStr)
	}()
	cmdutil.ExitIfError(err)

	cmdutil.Success("Comment %s updated on issue %q", commentID, issueKey)
	fmt.Printf("%s\n", cmdutil.GenerateServerBrowseURL(viper.GetString("server"), issueKey))
}
