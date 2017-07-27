package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jgsqware/termitask/view"
	"github.com/jgsqware/termitask/view/git"
	"github.com/jgsqware/termitask/view/tasks"
	tui "github.com/marcusolsson/tui-go"
)

var repos = []string{"/Users/juliengarciagonzalez/workspace/go/src/github.com/giantswarm/opsctl",
	"/Users/juliengarciagonzalez/workspace/go/src/github.com/giantswarm/leanix-exporter",
	"/Users/juliengarciagonzalez/workspace/go/src/github.com/jgsqware/clairctl"}

func main() {

	ght := os.Getenv("GITHUB_AUTH_TOKEN")

	if ght == "" {
		fmt.Println("GITHUB_AUTH_TOKEN no set")
		os.Exit(1)
	}

	f, err := os.OpenFile("exec.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
	ui := view.NewUI(ght)

	tb := tasks.NewTaskBox(*ui, "Regular Tasks", "Ctrl+r", "Ctrl+f")
	repoB := []tui.Widget{}
	for _, v := range repos {
		repoB = append(repoB, git.NewGitBox(*ui, v))
	}
	t := tui.SimpleFocusChain{}
	t.Set(ui.GetWidgets()...)
	ui.SetFocusChain(&t)
	ui.Append(tui.NewVBox(repoB...))
	ui.Append(tb)
	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
