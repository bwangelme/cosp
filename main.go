package main

import (
	"fmt"
	"os"

	"cos/cmd"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "cosp"}

	rootCmd.AddCommand(cmd.PasteCmd)
	rootCmd.AddCommand(cmd.UploadCmd)
	rootCmd.AddCommand(cmd.ListCmd)
	rootCmd.AddCommand(cmd.DeleteCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
