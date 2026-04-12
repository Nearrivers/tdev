package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/Nearrivers/tdev/internal/config"
	"github.com/Nearrivers/tdev/internal/tmux"
	"github.com/Nearrivers/tdev/internal/ui"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <nom>",
	Short: "Ouvrir un projet dans tmux",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		store, err := config.New()
		if err != nil {
			return err
		}

		cfg, err := store.Load()
		if err != nil {
			return err
		}

		project, ok := store.Get(cfg, name)
		if !ok {
			return fmt.Errorf(ui.PrintError("projet %q introuvable. Lance `tdev list` pour voir les projets disponibles"), name)
		}

		return openNewSession(project)
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
