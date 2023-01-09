package main

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/go-git/go-git/v5"

	d "gitlab.com/bogdzn/zru/deploy"
	r "gitlab.com/bogdzn/zru/release"
)

var (
	targetRemote   string
	targetBranch   string
	srcRemote      string
	srcBranch      string
	repositoryPath string
	newUrl         string
	oldUrl         string
	repoUrl        string
)

/*

   There are 3 commands that are supported by zru:

   - release: creates a branch /release from master, & pushes it to a new origin for software releases
   - deploy: renames URL in relevant places & deploys the application through docker-compose
   - certs: initiates ssl certficates for the application

*/

func main() {
	var version = "0.0.1"

	var releaseCmd = &cobra.Command{
		Use:     "release",
		Short:   "delivers code release across repositories",
		Long:    "creates target branch from source branch & pushes it to new remote/branch for software releases",
		Version: version,
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			color.Cyan("Running Release command")

			color.Cyan("Fetching update from both remotes...")
			r.FetchFromRemote(srcRemote, repositoryPath)
			r.FetchFromRemote(targetRemote, repositoryPath)

			color.Cyan("Switching to %s", srcBranch)
			r.SwitchBranch(srcRemote, srcBranch, repositoryPath)

			color.Cyan("Creating new %s branch on remote %s", targetBranch, targetRemote)
			r.SwitchBranch(targetRemote, targetBranch, repositoryPath)

			color.Cyan("Creating Release commit and pushing to %s", targetRemote)
			r.CommitAndPush(targetRemote, targetBranch, repositoryPath)

		},
	}

	var deployCmd = &cobra.Command{
		Use:     "deploy",
		Short:   "deploys the Zenith timesheet tool",
		Long:    "wraps around docker-compose to deploy the various services",
		Version: version,
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			/* fetch, switch to branch & stop docker */
			/* add option to completely nuke the repository if needed */

			color.Cyan("Running deploy command")
			d.ResetChanges(repositoryPath)
			d.SearchAndReplace(oldUrl, newUrl, repositoryPath)
			d.Deploy(repositoryPath)

		},
	}

	var cloneCmd = &cobra.Command{
		Use:     "clone",
		Short:   "clones a repository",
		Long:    "clones a repository recursively and print latest commit hash",
		Version: version,
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {

			r, err := git.PlainClone(repositoryPath, false, &git.CloneOptions{
				URL:               repoUrl,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})

			if err != nil {
				color.Red("Error when cloning repository")
				log.Fatal(err)
			}

			ref, err := r.Head()
			if err != nil {
				color.Red("Error when retrieving HEAD")
				log.Fatal(err)
			}

			commit, err := r.CommitObject(ref.Hash())
			if err != nil {
				color.Red("Error when retrieving latest commit")
				log.Fatal(err)
			}

			fmt.Println("[+] Latest commit:")
			fmt.Println(commit)
		},
	}

	releaseCmd.Flags().StringVarP(&targetRemote, "remote", "e", "customer", "git remote target")
	releaseCmd.Flags().StringVarP(&targetBranch, "target-branch", "t", "release", "branch to deliver release to")
	releaseCmd.Flags().StringVarP(&srcRemote, "src-remote", "r", "origin", "git remote source")
	releaseCmd.Flags().StringVarP(&srcBranch, "src-branch", "b", "master", "branch from which we deliver release")
	releaseCmd.Flags().StringVarP(&repositoryPath, "path", "p", ".", "repository path")

	deployCmd.Flags().StringVarP(&repositoryPath, "path", "p", ".", "repository path")
	deployCmd.Flags().StringVarP(&newUrl, "url", "u", "qa-timesheet.zenith-rh.com", "new deploy URL")
	deployCmd.Flags().StringVarP(&oldUrl, "original-url", "o", "timesheet.zenith-rh.com", "old deploy URL")

	cloneCmd.Flags().StringVarP(&repositoryPath, "path", "p", ".", "repository path")
	cloneCmd.Flags().StringVarP(&repoUrl, "url", "u", "https://gitlab.com/zenith-hr/TIMESHEET.git", "git clone url")

	cloneCmd.AddCommand(deployCmd, releaseCmd)

	if err := cloneCmd.Execute(); err != nil {
		panic(err)
	}
}
