// Package cmd
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tdev",
	Short: "Gestionnaires d'espaces de travail tmux pour les devs fullstack",
	Long:  "tdev - ouvre les projets front+back dans tmux en une commande",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(openCmd, listCmd, addCmd, deleteCmd)
}
