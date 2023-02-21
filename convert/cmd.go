package main

import (
	"github.com/spf13/cobra"
)


func NewCmd() (*cobra.Command) {

	cmd := &cobra.Command{
		
		
		Run: func(cmd *cobra.Command, args []string) error {
			fromPath := args[0]
			outPath := args[1]

			
		}
		
	}






}