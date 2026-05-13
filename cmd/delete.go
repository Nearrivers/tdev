package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Nearrivers/tdev/internal/config"
	"github.com/Nearrivers/tdev/internal/ui"
	"github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Supprimer un projet",
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

		f, err := fzf.New(
			fzf.WithLimit(1),
			fzf.WithCaseSensitive(false),
			fzf.WithKeyMap(fzf.KeyMap{
				Up:     []string{"up", "ctrl+p"},
				Down:   []string{"down", "ctrl+n"},
				Choose: []string{"enter"},
				Abort:  []string{"esc"},
			}),
			fzf.WithInputPosition(fzf.InputPositionBottom),
			fzf.WithInputPlaceholder("Sélectionner le projet à supprimer..."),
		)
		if err != nil {
			return err
		}

		idxs, err := f.Find(cfg.Projects, func(i int) string {
			p := cfg.Projects[i]
			return fmt.Sprintf("%s | Front: %s | Back: %s", p.Name, shorten(p.Front), shorten(p.Back))
		})
		if err != nil {
			return fmt.Errorf("sélection annulée")
		}

		if len(idxs) == 0 {
			return fmt.Errorf("aucun projet sélectionné")
		}

		project := cfg.Projects[idxs[0]]

		fmt.Println()
		fmt.Println("Projet à supprimer :")
		fmt.Printf("  Nom:   %s\n", ui.Bold.Render(project.Name))
		fmt.Printf("  Front: %s\n", project.Front)
		fmt.Printf("  Back:  %s\n", project.Back)
		fmt.Println()

		fmt.Print(ui.PrintWarning("Confirmer la suppression ? (y/N) : "))

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		response = strings.TrimSpace(response)
		if response != "y" && response != "Y" {
			fmt.Println(ui.Muted.Render("Annulation."))
			return nil
		}

		if err := store.Remove(cfg, project.Name); err != nil {
			return err
		}

		fmt.Println(ui.PrintSuccess(fmt.Sprintf("Projet %q supprimé", project.Name)))
		return nil
	},
}
