package delete

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Delete deletes an issue. To delete a task with subtasks, use '--cascade' flag.`
	examples = `$ jira issue delete ISSUE-1

# Delete task along with all of its subtasks
$ jira issue delete ISSUE-1 --cascade`
)

// NewCmdDelete is a delete command.
func NewCmdDelete() *cobra.Command {
	cmd := cobra.Command{
		Use:     "delete ISSUE-KEY",
		Short:   "Delete an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"remove", "rm", "del"},
		Annotations: map[string]string{
			"help:args": `ISSUE-KEY	Issue key, eg: ISSUE-1`,
		},
		Run: del,
	}

	cmd.Flags().Bool("cascade", false, "Delete issue along with its subtasks")
	cmd.Flags().Bool("yes", false, "Confirm deletion without interactive prompt")

	return &cmd
}

func del(cmd *cobra.Command, args []string) {
	project := viper.GetString("project.key")
	params := parseArgsAndFlags(cmd.Flags(), args, project)
	client := api.DefaultClient(params.debug)
	mc := deleteCmd{
		client:      client,
		transitions: nil,
		params:      params,
	}

	cmdutil.ExitIfError(mc.setIssueKey(project))

	if !params.yes {
		if !cmdutil.IsInteractive() {
			cmdutil.Failed("Use --yes flag to confirm deletion in non-interactive mode")
		}
		var confirm bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Are you sure you want to delete %q?", mc.params.key),
		}
		if err := survey.AskOne(prompt, &confirm); err != nil {
			cmdutil.ExitIfError(err)
		}
		if !confirm {
			cmdutil.Failed("Action aborted")
		}
	}

	err := func() error {
		s := cmdutil.Info(fmt.Sprintf("Removing issue %q", mc.params.key))
		defer s.Stop()

		return client.DeleteIssue(mc.params.key, mc.params.cascade)
	}()
	cmdutil.ExitIfError(err)

	cmdutil.Success(fmt.Sprintf("Issue %q removed successfully", mc.params.key))
}

type deleteParams struct {
	key     string
	cascade bool
	yes     bool
	debug   bool
}

func parseArgsAndFlags(flags query.FlagParser, args []string, project string) *deleteParams {
	var key string

	nargs := len(args)
	if nargs >= 1 {
		key = cmdutil.GetJiraIssueKey(project, args[0])
	}

	cascade, err := flags.GetBool("cascade")
	cmdutil.ExitIfError(err)

	yes, err := flags.GetBool("yes")
	cmdutil.ExitIfError(err)

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	return &deleteParams{
		key:     key,
		cascade: cascade,
		yes:     yes,
		debug:   debug,
	}
}

type deleteCmd struct {
	client      *jira.Client
	transitions []*jira.Transition
	params      *deleteParams
}

func (mc *deleteCmd) setIssueKey(project string) error {
	if mc.params.key != "" {
		return nil
	}
	if !cmdutil.IsInteractive() {
		return cmdutil.ErrNonInteractive
	}

	var ans string

	qs := &survey.Question{
		Name:     "key",
		Prompt:   &survey.Input{Message: "Issue key"},
		Validate: survey.Required,
	}
	if err := survey.Ask([]*survey.Question{qs}, &ans); err != nil {
		return err
	}
	mc.params.key = cmdutil.GetJiraIssueKey(project, ans)

	return nil
}
