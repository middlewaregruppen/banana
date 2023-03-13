package create

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdCreate() *cobra.Command {
	c := &cobra.Command{
		Use:     "create",
		Aliases: []string{"init"},
		Short:   "Initialize a new kmaint configuration in the current directory",
		Long:    "",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fmt.Printf("Creating new config ...")
			return nil
		},
	}
	return c
}
