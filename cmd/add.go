package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Nearrivers/tdev/internal/config"
	"github.com/Nearrivers/tdev/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <nom> <chemin-front> <chemin-back>",
	Short: "Déclarer un nouveau projet",
	Long: `Ajoute une nouveau projet à la liste des projets connus.

Si le chemin commence par un "/", alors tdev part du principe que le chemin est absolu.

Si zoxide est installé, le CLI va d'abord tenter de regarder si les chemins fournis
n'existent pas déjà.

Dans le dernier cas, le chemin fourni sera construit de la manière suivante : $HOME/chemin-donné.`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, front, back := args[0], args[1], args[2]

		// Résolution des chemins via zoxide si disponible
		frontPath, err := resolvePath(front)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(ui.PrintError("Le chemin vers le projet Front pointe vers un répertoire inexistant"))
				return nil
			}
			return fmt.Errorf("front: %w", err)
		}
		backPath, err := resolvePath(back)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(ui.PrintError("Le chemin vers le projet Back pointe vers un répertoire inexistant"))
				return nil
			}
			return fmt.Errorf("back: %w", err)
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
