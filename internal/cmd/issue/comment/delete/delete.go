package delete

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
)

const helpText = `Delete removes a comment from an issue.`

// NewCmdCommentDelete is a comment delete command.
func NewCmdCommentDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ISSUE-KEY COMMENT-ID",
		Short: "Delete a comment from an issue",
		Long:  helpText,
		Args:  cobra.ExactArgs(2),
		Run:   del,
	}
}

func del(cmd *cobra.Command, args []string) {
	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])
	commentID := args[1]

	debug, _ := cmd.Flags().GetBool("debug")
	client := api.DefaultClient(debug)

	err := func() error {
		s := cmdutil.Info("Deleting comment")
		defer s.Stop()

		return client.DeleteIssueComment(issueKey, commentID)
	}()
	cmdutil.ExitIfError(err)

	cmdutil.Success("Comment %s deleted from issue %q", commentID, issueKey)
	fmt.Printf("%s\n", cmdutil.GenerateServerBrowseURL(viper.GetString("server"), issueKey))
}
