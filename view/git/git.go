package git

import (
	"bytes"
	"context"
	"log"
	"os/exec"
	"regexp"

	"strings"

	"fmt"

	"github.com/google/go-github/github"
	"github.com/jgsqware/termitask/view"
	tui "github.com/marcusolsson/tui-go"
)

func githubIssues(client *github.Client, owner, repo string) []*github.Issue {

	ctx := context.Background()
	i, _, err := client.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{Sort: "created", ListOptions: github.ListOptions{Page: 1, PerPage: 20}})
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func githubPRs(client *github.Client, owner, repo string) []*github.PullRequest {

	ctx := context.Background()
	prs, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{ListOptions: github.ListOptions{Page: 1, PerPage: 5}})
	if err != nil {
		log.Fatal(err)
	}
	return prs
}

func currentBranch(path string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	return out.String()
}

func githubInfo(path string) (string, string, error) {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("cannot retrieve remote: %v", err)

	}
	var re = regexp.MustCompile(`(?m)origin[^@]*@github.com:([^\/]*)\/([^ ]*)`)

	if re.MatchString(out.String()) {
		m := re.FindStringSubmatch(out.String())

		return m[1], strings.TrimSuffix(m[2], ".git"), nil
	}
	return "", "", fmt.Errorf("Is not github project: %v", path)
}

func NewGitBox(ui view.UI, gitPath string) *tui.Box {

	lb := tui.NewVBox()
	lb.SetTitle(gitPath)
	lb.SetBorder(true)
	lb.SetSizePolicy(tui.Maximum, tui.Maximum)

	lb.Append(tui.NewLabel(""))
	addIssues(ui, lb, gitPath)
	lb.Append(tui.NewLabel(""))
	addPrs(ui, lb, gitPath)
	lb.Append(tui.NewSpacer())
	lb.Append(tui.NewLabel(""))
	addLog(lb, gitPath)
	lb.Append(tui.NewSpacer())
	ui.AddWidget(lb, "Ctrl+g")
	return lb

}

func addPrs(ui view.UI, lb *tui.Box, gitPath string) {
	owner, repo, err := githubInfo(gitPath)

	if err != nil {
		log.Fatal(err)
	}

	prs := githubPRs(ui.GithubClient, owner, repo)

	if len(prs) > 0 {
		vb := tui.NewVBox()
		vb.SetBorder(true)
		vb.SetTitle("Last PRs")

		for _, pr := range prs {
			n := tui.NewLabel(fmt.Sprintf("#%v", pr.GetNumber()))
			n.SetStyleName(view.StyleGitPRS)
			l := tui.NewLabel("<" + pr.User.GetLogin() + ">")
			l.SetStyleName(view.StyleGitAuthor)
			vb.Append(tui.NewHBox(
				n,
				tui.NewPadder(1, 0, tui.NewLabel(pr.GetTitle())),
				tui.NewPadder(1, 0, l),
				tui.NewSpacer(),
			))
		}
		lb.Append(vb)
	}
}

func addIssues(ui view.UI, lb *tui.Box, gitPath string) {
	owner, repo, err := githubInfo(gitPath)

	if err != nil {
		log.Fatal(err)
	}

	issues := githubIssues(ui.GithubClient, owner, repo)

	if len(issues) > 0 {
		vb := tui.NewVBox()
		vb.SetBorder(true)
		vb.SetTitle("Last Issues")

		c := 0
		for _, pr := range issues {
			if c < 5 {
				if pr.PullRequestLinks == nil {
					n := tui.NewLabel(fmt.Sprintf("#%v", pr.GetNumber()))
					n.SetStyleName(view.StyleGitIssues)
					l := tui.NewLabel("<" + pr.User.GetLogin() + ">")
					l.SetStyleName(view.StyleGitAuthor)
					vb.Append(tui.NewHBox(
						n,
						tui.NewPadder(1, 0, tui.NewLabel(pr.GetTitle())),
						tui.NewPadder(1, 0, l),
						tui.NewSpacer(),
					))
					c++
				}
			}
		}
		lb.Append(vb)
	}
}

func addLog(lb *tui.Box, gitPath string) {

	vb := tui.NewVBox()
	vb.SetTitle("Currently on " + currentBranch(gitPath))
	vb.SetBorder(true)
	cmd := exec.Command("git", "log", "--pretty=format:%h;;%s;;%cr;;%an", "--abbrev-commit", "-1")
	cmd.Dir = gitPath
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range strings.Split(out.String(), "\n") {
		c := strings.Split(v, ";;")

		id := tui.NewLabel(c[0])
		id.SetStyleName(view.StyleGitID)
		msg := tui.NewLabel(c[1])
		date := tui.NewLabel("(" + c[2] + ")")
		date.SetStyleName(view.StyleGitDate)
		author := tui.NewLabel("<" + c[3] + ">")
		author.SetStyleName(view.StyleGitAuthor)

		vb.Append(tui.NewHBox(
			id,
			tui.NewPadder(1, 0, msg),
			tui.NewPadder(1, 0, date),
			author,
			tui.NewSpacer(),
		))
	}
	lb.Append(vb)
}
