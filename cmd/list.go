package cmd

import (
	"fmt"

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

		fmt.Println(ui.Title.Render(fmt.Sprintf(" tdev — %d projet(s) ", len(cfg.Projects))))

		for _, p := range cfg.Projects {
			card := lipgloss.JoinVertical(lipgloss.Left,
				ui.Bold.Render(p.Name),
				ui.BadgeFront.Render("front")+" "+ui.Muted.Render(p.Front),
				ui.BadgeBack.Render("back ")+" "+ui.Muted.Render(p.Back),
			)
			fmt.Println(ui.ProjectCard.Render(card))
		}
		return nil
	},
}
