package main

import (
	"fmt"
	"os"

	"cos/cmd"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "cos"}

	rootCmd.AddCommand(cmd.PasteCmd)
	rootCmd.AddCommand(cmd.UploadCmd)
	rootCmd.AddCommand(cmd.ListCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
