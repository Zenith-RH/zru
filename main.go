package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/go-git/go-git/v5"

	c "gitlab.com/bogdzn/zru/certs"
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
	domain         string
	email          string
)

/*

   There are 3 commands that are supported by zru:

   - release: creates a branch /release from master, & pushes it to a new origin for software releases
   - deploy: renames URL in relevant places & deploys the application through docker-compose
   - certs: initiates ssl certficates for the application

*/

func main() {
	var version = "0.0.2"

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
			color.Cyan("Running deploy command")

			d.ResetChanges(repositoryPath)

			command := exec.Command("git", "pull")
			c.Run(command, "git pull", false, repositoryPath)

			d.SearchAndReplace(oldUrl, newUrl, repositoryPath)

			command = exec.Command("docker-compose", "down")
			c.Run(command, "docker-compose down", false, repositoryPath)

			command = exec.Command("docker-compose", "up", "--build", "-d")
			c.Run(command, "docker-compose up", true, repositoryPath)

		},
	}

	var certsCmd = &cobra.Command{
		Use:     "certs",
		Short:   "fetches new SSL certificates",
		Long:    "To be used on initial deployments -- generates new SSL certificates for your new env",
		Version: version,
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			color.Cyan("Running certs command")

			color.Green("Downloading recommended TLS files")
			c.GetFile(
				"https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf",
				"data/certbot/conf/", "options-ssl-nginx.conf")
			c.GetFile(
				"https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem",
				"data/certbot/conf/", "ssl-dhparams.pem")

			color.Green("Creating dummy certificates")

			keysPath := filepath.Join("/etc/letsencrypt/live/", domain)
			privPath := filepath.Join(keysPath, "privkey.pem")
			fullchainPath := filepath.Join(keysPath, "fullchain.pem")

			pathInVolume := filepath.Join("data/certbot/conf/live/", domain)
			err := os.MkdirAll(pathInVolume, 0777)
			if err != nil {
				color.Red("Could not create directory %s", pathInVolume)
				log.Fatal(err)
			}

			entrypoint := fmt.Sprintf("openssl req -x509 -nodes -newkey rsa:4096 -days 1 -keyout '%s' -out '%s' -subj '/CN=localhost'", privPath, fullchainPath)
			toRun := exec.Command("docker-compose", "run", "--rm", "--entrypoint", entrypoint, "certbot")
			c.Run(toRun, "docker-compose run create key", false, repositoryPath)

			color.Green("Booting up nginx")
			toRun = exec.Command("docker-compose", "up", "--force-recreate", "-d", "nginx")
			c.Run(toRun, "docker-compose up -d nginx", false, repositoryPath)

			color.Green("Deleting dummy certificates")
			entrypoint = fmt.Sprintf("rm -Rf /etc/letsencrypt/live/%s && rm -Rf /etc/letsencrypt/archive/%s && rm -Rf /etc/letsencrypt/renewal/%s.conf", domain, domain, domain)
			toRun = exec.Command("docker-compose", "run", "--rm", "--entrypoint", entrypoint, "certbot")
			c.Run(toRun, "docker-compose rm certs", false, repositoryPath)

			color.Green("Requesting real certificates")
			entrypoint = fmt.Sprintf("certbot certonly --webroot -w /var/www/certbot --email %s -d %s --rsa-key-size 4096 --agree-tos --force-renewal", email, domain)
			toRun = exec.Command("echo", "yes", "|", "docker-compose", "run", "--rm", "--entrypoint", entrypoint, "certbot")
			c.Run(toRun, "docker-compose run certbot", false, repositoryPath)

			color.Green("Reloading nginx")
			toRun = exec.Command("docker-compose", "exec", "nginx", "nginx", "-s", "reload")
			c.Run(toRun, "docker-compose run nginx -s reload", true, repositoryPath)
		},
	}

	var cloneCmd = &cobra.Command{
		Use:     "clone",
		Short:   "clones a repository",
		Long:    "clones a repository recursively and print latest commit hash",
		Version: version,
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			color.Cyan("Running clone command")

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

	certsCmd.Flags().StringVarP(&domain, "url", "u", "timesheet.zenith-rh.com", "domain of your new environment")
	certsCmd.Flags().StringVarP(&email, "email", "e", "backoffice@zenith-rh.com", "renewal email")

	cloneCmd.AddCommand(deployCmd, releaseCmd, certsCmd)

	if err := cloneCmd.Execute(); err != nil {
		panic(err)
	}
}
