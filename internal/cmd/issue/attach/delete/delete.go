package delete

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
)

const helpText = `Delete removes an attachment by its ID.`

// NewCmdAttachDelete is an attach delete command.
func NewCmdAttachDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ATTACHMENT-ID",
		Short: "Delete an attachment",
		Long:  helpText,
		Args:  cobra.ExactArgs(1),
		Run:   deleteAttachment,
	}
}

func deleteAttachment(cmd *cobra.Command, args []string) {
	attachmentID := args[0]

	debug, _ := cmd.Flags().GetBool("debug")
	client := api.DefaultClient(debug)

	err := func() error {
		s := cmdutil.Info("Deleting attachment")
		defer s.Stop()

		return client.DeleteAttachment(attachmentID)
	}()
	cmdutil.ExitIfError(err)

	cmdutil.Success("Attachment %s deleted", attachmentID)
}
