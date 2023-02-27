package main

import (
	"fmt"
	"os"

	"github.com/notional-labs/demerklizator"
	"github.com/spf13/cobra"
)

var version = "0.0.1"
var rootCmd = &cobra.Command{
	Use:     "demerk",
	Version: version,
	Short:   "demerk - a tool to convert iavl merklized data to normal db data",

	RunE: func(cmd *cobra.Command, args []string) error {

		rootDir := args[0]
		outRootDir := args[1]
		//convert fromPath before parsing to from field
		applicationDBPath := demerklizator.ApplicationDBPathFromRootDir(rootDir)
		outApplicationDBPath := demerklizator.ApplicationDBPathFromRootDir(outRootDir)
		err := demerklizator.MigrateLatestStateDataToDBStores(applicationDBPath, outApplicationDBPath)
		if err != nil {
			return err
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}
