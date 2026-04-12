// Package tmux
package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// run exécute une commande tmux et retourne une erreur si elle échoue
func run(args ...string) error {
	cmd := exec.Command("tmux", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// output exécute une commande tmux et retourne sa sortie
func output(args ...string) (string, error) {
	out, err := exec.Command("tmux", args...).Output()
	return string(out), err
}

// IsInsideTmux retourne true si on est dans une session tmux active
func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// CurrentSession retourne le nom de la session tmux courante
func CurrentSession() (string, error) {
	out, err := output("display-message", "-p", "#S")
	if err != nil {
		return "", err
	}
	// Trim newline
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out, nil
}

func RenameSession(currentName, newName string) (string, error) {
	out, err := output("rename-session", "-t", currentName, newName)
	if err != nil {
		return "", err
	}

	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out, nil
}

// SessionExists vérifie si une session existe
func SessionExists(name string) bool {
	err := exec.Command("tmux", "has-session", "-t", name).Run()
	return err == nil
}

// CreateProjectWindows crée les 2 windows (front + back) avec leurs panes
// dans la session donnée.
func CreateProjectWindows(session, projectName, frontPath, backPath string) error {
	if _, err := strconv.Atoi(string(projectName[0])); err == nil {
		projectName = "P" + projectName
	}
	// ── Window Frontend ──────────────────────────────────────────
	winFront, err := initDevEnv(session, projectName+"|front", frontPath)

	if err != nil {
		return err
	}

	// ── Window Backend ───────────────────────────────────────────
	if _, err = initDevEnv(session, projectName+"|back", backPath); err != nil {
		return err
	}

	// Focus sur la window front
	return run("select-window", "-t", winFront)
}

func trim(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}

func initDevEnv(session, name, path string) (string, error) {
	if err := run("new-window", "-t", session, "-n", name); err != nil {
		return "", fmt.Errorf("new-window %s: %w", name, err)
	}

	windowName, err := output("display-message", "-p", "#{window_index}")
	if err != nil {
		return "", fmt.Errorf("get window index: %w", err)
	}

	windowName = trim(windowName)

	if err := run("send-keys", "-t", windowName, "cd '"+path+"'", "Enter"); err != nil {
		return "", err
	}
	if err := run("send-keys", "-t", windowName, "nvim", "Enter"); err != nil {
		return "", err
	}
	if err := run("split-window", "-v", "-t", windowName); err != nil {
		return "", err
	}
	if err := run("select-pane", "-D"); err != nil {
		return "", err
	}
	if err := run("send-keys", "-t", windowName+".2", "cd '"+path+"'", "Enter"); err != nil {
		return "", err
	}
	if err := run("split-window", "-h", "-t", windowName+".2"); err != nil {
		return "", err
	}
	if err := run("send-keys", "-t", windowName+".3", "cd '"+path+"'", "Enter"); err != nil {
		return "", err
	}
	if err := run("select-pane", "-U"); err != nil {
		return "", err
	}

	return windowName, nil
}
