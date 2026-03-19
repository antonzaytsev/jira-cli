package linktypes

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
)

const (
	helpText = `Link-types lists available issue link types.`
	examples = `$ jira issue link-types

# Output as JSON
$ jira issue link-types --raw`
)

// NewCmdLinkTypes is a link-types command.
func NewCmdLinkTypes() *cobra.Command {
	cmd := cobra.Command{
		Use:     "link-types",
		Short:   "List available issue link types",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"lt"},
		Run:     linkTypes,
	}

	cmd.Flags().Bool("raw", false, "Output raw JSON response")

	return &cmd
}

func linkTypes(cmd *cobra.Command, _ []string) {
	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	raw, err := cmd.Flags().GetBool("raw")
	cmdutil.ExitIfError(err)

	client := api.DefaultClient(debug)

	types, err := client.GetIssueLinkTypes()
	cmdutil.ExitIfError(err)

	if raw {
		out, err := json.MarshalIndent(types, "", "  ")
		cmdutil.ExitIfError(err)
		fmt.Println(string(out))
		return
	}

	if len(types) == 0 {
		fmt.Println("No link types available")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "ID\tNAME\tINWARD\tOUTWARD\n")
	for _, lt := range types {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", lt.ID, lt.Name, lt.Inward, lt.Outward)
	}
	w.Flush()
}
