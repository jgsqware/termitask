package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jgsqware/termitask/view"
	"github.com/jgsqware/termitask/view/git"
	"github.com/jgsqware/termitask/view/tasks"
	tui "github.com/marcusolsson/tui-go"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("termitask")
	viper.AddConfigPath("$HOME/.termitask")
	viper.AddConfigPath(".")
	viper.BindEnv("github_auth_token", "GITHUB_AUTH_TOKEN")
	err := viper.ReadInConfig()
	viper.WatchConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	ght := viper.GetString("github_auth_token")

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

	tb := tasks.NewTaskBox(*ui, "Regular Tasks", "Ctrl+r", "Ctrl+f", "Ctrl+c")
	repoB := []tui.Widget{}
	for _, v := range viper.GetStringSlice("repos") {
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
