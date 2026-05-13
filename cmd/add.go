package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Nearrivers/tdev/internal/config"
	"github.com/Nearrivers/tdev/internal/ui"
	"github.com/koki-develop/go-fzf"
	"github.com/spf13/cobra"
)

const hereFlag = "--here"

var addCmd = &cobra.Command{
	Use:   "add <nom> [chemin-front] [chemin-back]",
	Short: "Déclarer un nouveau projet",
	Long: `Ajoute un nouveau projet à la liste des projets connus.

Les chemins front et back sont optionnels. Si omis, une interface fzf
permet de sélectionner un répertoire parmi ~/Projects et ~/Documents.

Si un chemin est "--here", le répertoire courant sera utilisé.

Sinon, si le chemin commence par "/", il est considéré comme absolu.
Si zoxide est installé, le CLI tentera de résoudre les chemins relatifs.
Dans le dernier cas, le chemin sera construit : $HOME/chemin-donné.`,
	Args: cobra.RangeArgs(1, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		frontPath, err := resolveOrPrompt(1, args, "Front")
		if err != nil {
			return err
		}
		backPath, err := resolveOrPrompt(2, args, "Back")
		if err != nil {
			return err
		}

		store, err := config.New()
		if err != nil {
			return err
		}

		cfg, err := store.Load()
		if err != nil {
			return err
		}

		_, exists := store.Get(cfg, name)
		if err := store.Add(cfg, config.Project{Name: name, Front: frontPath, Back: backPath}); err != nil && err != config.ErrProjectAlreadyExists {
			return err
		}

		action := "ajouté"
		if exists {
			action = "mis à jour"
		}

		fmt.Println(ui.PrintSuccess(fmt.Sprintf("Projet %q %s", name, action)))
		fmt.Println(ui.PrintInfo("  front", frontPath))
		fmt.Println(ui.PrintInfo("  back ", backPath))
		return nil
	},
}

// resolvePath tente zoxide, sinon retourne le chemin tel quel
func resolvePath(input string) (string, error) {
	// Chemin absolu → on le garde
	if strings.HasPrefix(input, "/") || strings.HasPrefix(input, "~") {
		return input, nil
	}

	// Tente zoxide
	if _, err := exec.LookPath("zoxide"); err == nil {
		out, err := exec.Command("zoxide", "query", input).Output()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}

	// Fallback : chemin relatif depuis home
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := home + "/" + input

	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	return path, nil
}

// resolveOrPrompt résout un chemin explicitement fourni, utilise le répertoire
// courant si "--here", ou lance une sélection fzf si l'argument est absent.
func resolveOrPrompt(argIndex int, args []string, label string) (string, error) {
	if argIndex >= len(args) {
		return selectPathWithFzf(label)
	}
	arg := args[argIndex]
	if arg == hereFlag {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("impossible de déterminer le répertoire courant: %w", err)
		}
		return wd, nil
	}
	path, err := resolvePath(arg)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(ui.PrintError(fmt.Sprintf("Le chemin vers le projet %s pointe vers un répertoire inexistant", label)))
			return "", nil
		}
		return "", fmt.Errorf("%s: %w", strings.ToLower(label), err)
	}
	return path, nil
}

// selectPathWithFzf lance une interface fzf pour sélectionner un répertoire.
func selectPathWithFzf(label string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	paths := collectDirectories([]string{
		filepath.Join(home, "Projets"),
		filepath.Join(home, "Documents"),
	})

	if len(paths) == 0 {
		return "", fmt.Errorf("aucun répertoire trouvé dans ~/Projects et ~/Documents")
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
		fzf.WithInputPlaceholder(fmt.Sprintf("Chemin %s — rechercher un répertoire...", label)),
	)
	if err != nil {
		return "", err
	}

	idxs, err := f.Find(paths, func(i int) string {
		return paths[i]
	})
	if err != nil {
		return "", fmt.Errorf("sélection fzf annulée pour le chemin %s", label)
	}

	if len(idxs) == 0 {
		return "", fmt.Errorf("aucun chemin %s sélectionné", strings.ToLower(label))
	}

	return paths[idxs[0]], nil
}

// collectDirectories retourne les répertoires enfants directs des racines
// fournies, sans descendre dans les sous-répertoires.
func collectDirectories(roots []string) []string {
	var dirs []string
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				dirs = append(dirs, filepath.Join(root, entry.Name()))
			}
		}
	}
	return dirs
}
