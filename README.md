# tdev

Gestionnaire de workspaces tmux pour devs fullstack.

## Installation

```bash
go install github.com/user/tdev@latest
```

Ou depuis les sources :

```bash
git clone ...
cd tdev
go build -o tdev .
mv tdev /usr/local/bin/
```

## Usage

```bash
# Enregistrer un projet (chemins absolus ou noms zoxide)
tdev add monapp ~/projets/monapp-front ~/projets/monapp-back
tdev add monapp monapp-front monapp-back   # avec zoxide

# Ouvrir dans une nouvelle session tmux
tdev open monapp

# Rechercher et ouvrir les projets connus dans la session tmux courante
tdev open  

Cette commande lance une session si aucune n'est en cours

# Lister les projets
tdev list
tdev ls

# Supprimer un projet
tdev remove monapp
tdev rm monapp
```

## Structure des windows créées

```sh
Session "monapp"
├── Window 1 : front
│   ├── Pane 0 (haut)        → éditeur / commandes principales
│   ├── Pane 1 (bas gauche)  → npm run dev
│   └── Pane 2 (bas droite)  → installs de packages
└── Window 2 : back
    ├── Pane 0 (haut)        → éditeur / commandes principales
    ├── Pane 1 (bas gauche)  → serveur
    └── Pane 2 (bas droite)  → installs de packages
```

## Structure du projet

```sh
tdev/
├── main.go
├── go.mod
├── cmd/
│   ├── root.go      # cobra root + init
│   ├── add.go       # tdev add
│   ├── open.go      # tdev open
│   ├── list.go      # tdev list
│   └── remove.go    # tdev remove
└── internal/
    ├── config/
    │   └── config.go  # lecture/écriture ~/.config/tdev/projects.toml
    ├── tmux/
    │   └── tmux.go    # appels tmux via os/exec
    └── ui/
        └── styles.go  # styles lipgloss
```
