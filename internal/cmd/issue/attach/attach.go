package attach

import (
	"github.com/spf13/cobra"

	attachadd "github.com/ankitpokhrel/jira-cli/internal/cmd/issue/attach/add"
	attachdelete "github.com/ankitpokhrel/jira-cli/internal/cmd/issue/attach/delete"
)

const helpText = `Attach command helps you manage issue attachments. See available commands below.`

// NewCmdAttach is an attachment command.
func NewCmdAttach() *cobra.Command {
	cmd := cobra.Command{
		Use:     "attach",
		Short:   "Manage issue attachments",
		Long:    helpText,
		Aliases: []string{"attachment"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		attachadd.NewCmdAttachAdd(),
		attachdelete.NewCmdAttachDelete(),
	)

	return &cmd
}
