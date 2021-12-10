package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func init() {
	rootCmd.AddCommand(docsCmd)
}

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate documentation",
	Long:   `Generating Markdown pages.`,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := doc.GenMarkdownTree(rootCmd, "./docs")

		if err == nil {
			fmt.Println("The documentation has been successfully generated.")
		}

		return err
	},
}
