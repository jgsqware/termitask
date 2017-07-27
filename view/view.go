package view

import (
	"context"
	"log"

	"github.com/boltdb/bolt"
	"github.com/google/go-github/github"
	tui "github.com/marcusolsson/tui-go"
	"golang.org/x/oauth2"
)

type UI struct {
	tui.UI

	widgets      []tui.Widget
	root         *tui.Box
	GithubClient *github.Client
	Db           *bolt.DB
}

func (vc *UI) AddWidget(w tui.Widget, keybinding string) {
	vc.widgets = append(vc.widgets, w)
	vc.UI.SetKeybinding(keybinding, func() {
		vc.UnFocusedAll()
		(w).SetFocused(true)
	})
}

func (vc *UI) Append(w tui.Widget) {
	vc.root.Append(w)
}

func (vc UI) UnFocusedAll() {
	for _, v := range vc.widgets {
		v.SetFocused(false)
	}
}

func (vc UI) GetWidgets() []tui.Widget {
	return vc.widgets
}

func initDB(name string) *bolt.DB {
	db, err := bolt.Open(name, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func NewUI(githubAuthToken string) *UI {

	root := tui.NewHBox()
	ui := UI{
		root:         root,
		UI:           tui.New(root),
		widgets:      []tui.Widget{},
		GithubClient: initGithubClient(githubAuthToken),
		Db:           initDB("my.db"),
	}
	t := tui.DefaultTheme
	t.SetStyle("label."+StyleGitID, tui.Style{Fg: tui.ColorRed})
	t.SetStyle("label."+StyleGitDate, tui.Style{Fg: tui.ColorGreen})
	t.SetStyle("label."+StyleGitAuthor, tui.Style{Fg: tui.ColorBlue})
	t.SetStyle("label."+StyleGitIssues, tui.Style{Fg: tui.ColorRed})
	t.SetStyle("label."+StyleGitPRS, tui.Style{Fg: tui.ColorYellow})
	ui.SetTheme(t)

	return &ui
}

func initGithubClient(accessToken string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

const (
	StyleGitID     = "git.id"
	StyleGitDate   = "git.date"
	StyleGitAuthor = "git.author"
	StyleGitIssues = "git.issues"
	StyleGitPRS    = "git.prs"
)
