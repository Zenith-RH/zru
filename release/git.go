package zru

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func worktree(path string) *git.Worktree {

	// We instantiate a new repository targeting the given path (the .git folder)
	r := repo(path)

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		fmt.Println("Could not open worktree")
		os.Exit(1)
	}

	return w
}

func repo(path string) *git.Repository {
	// We instantiate a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(path)
	if err != nil {
		fmt.Println("Could not find git repository in path: " + path)
		os.Exit(1)
	}

	return r
}

func SwitchBranch(remote string, branch string, path string) {
	wk := worktree(path)

	var err = wk.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName(remote, branch),
		Create: true,
	})

	if err != nil {
		color.Red("Could not checkout to new branch")
		log.Fatal(err)
	}
}

func CommitAndPush(remote string, branch string, path string) {
	r := repo(path)
	w := worktree(path)

	// Add and commit the changes
	_, err := w.Add(".")
	if err != nil {
		color.Red("Could not add the changes")
		log.Fatal(err)
	}

	status, err := w.Status()
	if err != nil {
		color.Red("Could not get git status")
		log.Fatal(err)
	}

	fmt.Println(status)

	_, err = w.Commit("sending new release", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "zenith-release-utility",
			Email: "release@timesheet-rh.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		color.Red("Could not commit")
		log.Fatal(err)
	}

	// Push the changes to the remote
	err = r.Push(&git.PushOptions{
		RemoteName: remote,
	})
	if err != nil {
		color.Red("Could not push to new remote")
		log.Fatal(err)
	}

	hash, err := r.ResolveRevision(plumbing.Revision("HEAD"))
	if err != nil {
		color.Red("Could not get commit hash")
		log.Fatal(err)
	}

	fmt.Println("[+] Commit hash:")
	fmt.Print(hash)
}

func FetchFromRemote(remote string, path string) {

	color.Cyan("Fetching from remote %s", remote)

	r := repo(path)
	opts := &git.FetchOptions{
		RemoteName: remote,
	}

	r.Fetch(opts)
}
