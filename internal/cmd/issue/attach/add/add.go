package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const helpText = `Add uploads a file attachment to an issue.`

// NewCmdAttachAdd is an attach add command.
func NewCmdAttachAdd() *cobra.Command {
	return &cobra.Command{
		Use:   "add ISSUE-KEY FILE_PATH",
		Short: "Upload an attachment to an issue",
		Long:  helpText,
		Args:  cobra.ExactArgs(2),
		Run:   addAttachment,
	}
}

func addAttachment(cmd *cobra.Command, args []string) {
	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])
	filePath := args[1]

	debug, _ := cmd.Flags().GetBool("debug")
	client := api.DefaultClient(debug)

	var attachments []*jira.Attachment

	err := func() error {
		s := cmdutil.Info("Uploading attachment")
		defer s.Stop()

		var err error
		attachments, err = client.AddAttachment(issueKey, filePath)
		return err
	}()
	cmdutil.ExitIfError(err)

	for _, a := range attachments {
		cmdutil.Success("Attachment %q uploaded to %s (id: %s, size: %d bytes)", a.Filename, issueKey, a.ID, a.Size)
	}
	fmt.Printf("%s\n", cmdutil.GenerateServerBrowseURL(viper.GetString("server"), issueKey))
}
