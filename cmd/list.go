package cmd

import (
	"fmt"
	"strings"

	"github.com/Nearrivers/tdev/internal/config"
	"github.com/Nearrivers/tdev/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lister tous les projets enregistrés",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := config.New()
		if err != nil {
			return err
		}

		cfg, err := store.Load()
		if err != nil {
			return err
		}

		if len(cfg.Projects) == 0 {
			fmt.Println(ui.Muted.Render("Aucun projet enregistré. Lance `tdev add` pour commencer."))
			return nil
		}

		fmt.Println()

		title := lipgloss.NewStyle().Bold(true).MarginLeft(2).Render("tdev — projets")
		fmt.Println(title)
		fmt.Println()

		nameW, frontW, backW := 6, 5, 5
		for _, p := range cfg.Projects {
			if l := len(p.Name); l > nameW {
				nameW = l
			}
			if l := len(shorten(p.Front)); l > frontW {
				frontW = l
			}
			if l := len(shorten(p.Back)); l > backW {
				backW = l
			}
		}

		rowFmt := "  %-*s  %-*s  %s"
		header := fmt.Sprintf(rowFmt, nameW, "NOM", frontW, "FRONT", "BACK")
		fmt.Println(ui.Muted.Render(header))
		fmt.Println(ui.Muted.Render("  " + strings.Repeat("─", nameW) + "  " + strings.Repeat("─", frontW) + "  " + strings.Repeat("─", backW)))

		for _, p := range cfg.Projects {
			fmt.Printf(rowFmt+"\n", nameW, p.Name, frontW, shorten(p.Front), shorten(p.Back))
		}
		fmt.Println()
		return nil
	},
}
