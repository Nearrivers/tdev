package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/Nearrivers/tdev/internal/config"
	"github.com/Nearrivers/tdev/internal/tmux"
	"github.com/Nearrivers/tdev/internal/ui"
	"github.com/spf13/cobra"

	"github.com/koki-develop/go-fzf"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Ouvrir un projet dans tmux",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := config.New()
		if err != nil {
			return err
		}

		cfg, err := store.Load()
		if err != nil {
			return err
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
			fzf.WithInputPlaceholder("Rechercher un projet..."),
		)
		if err != nil {
			return err
		}

		idxs, err := f.Find(cfg.Projects, func(i int) string {
			return fmt.Sprintf("%s | Front: %s | Back: %s", cfg.Projects[i].Name, cfg.Projects[i].Front, cfg.Projects[i].Back)
		})
		if err != nil {
			return err
		}

		for _, i := range idxs {
			fmt.Println(cfg.Projects[i])
		}

		if len(idxs) == 0 {
			return fmt.Errorf("%s", ui.PrintError("Lance `tdev list` pour voir les projets disponibles"))
		}

		if len(idxs) != 1 {
			return errors.New("un seul projet doit être sélectionné")
		}

		project := cfg.Projects[idxs[0]]
		return openNewSession(&project)
	},
}

func openNewSession(p *config.Project) error {
	session := p.Name

	if tmux.IsInsideTmux() {
		return openAppend(p)
	}

	if tmux.SessionExists(session) {
		fmt.Println(ui.PrintWarning(fmt.Sprintf("Session %q déjà ouverte, attachement...", session)))
		return attachSession(session)
	}

	// Crée la session en arrière-plan
	if err := exec.Command("tmux", "new-session", "-d", "-s", session).Run(); err != nil {
		return fmt.Errorf("impossible de créer la session: %w", err)
	}

	fmt.Println(ui.PrintSuccess(fmt.Sprintf("Session %q créée", session)))

	if err := tmux.CreateProjectWindows(session, p.Name, p.Front, p.Back); err != nil {
		return err
	}

	// Supprime la window vide créée par défaut (index 1 avant nos windows)
	exec.Command("tmux", "kill-window", "-t", session+":1").Run()

	return attachSession(session)
}

func openAppend(p *config.Project) error {
	session, err := tmux.CurrentSession()
	if err != nil {
		return err
	}

	// Si le nom de la session est un nombre, on la renome car le nombre peut coincider
	// avec l'index d'une fenêtre et cause des problèmes
	if _, err := strconv.Atoi(session); err == nil {
		s, err := tmux.RenameSession(session, p.Name)
		if err != nil {
			return err
		}

		session = s
	}

	fmt.Println(ui.PrintSuccess(fmt.Sprintf("Ajout de %q dans la session %q", p.Name, session)))

	// Préfixe le nom des windows pour éviter les conflits
	return tmux.CreateProjectWindows(session, p.Name, p.Front, p.Back)
}

func attachSession(session string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", session)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
