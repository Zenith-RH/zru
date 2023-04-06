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

	c "github.com/zenith-rh/zru/certs"
	d "github.com/zenith-rh/zru/deploy"
	r "github.com/zenith-rh/zru/release"
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

   There are 5 commands that are supported by zru:

   - release: creates a branch /release from master, & pushes it to a new origin for software releases
   - deploy: renames URL in relevant places & deploys the application through docker compose
   - certs: initiates ssl certficates for the application
   - clone: git clones a bare repository
   - nuke: destroys an infrastructure, and resets it

*/

func main() {
	var version = "0.0.6"

	var releaseCmd = &cobra.Command{
		Use:     "release",
		Short:   "delivers code release across repositories",
		Long:    "creates target branch from source branch & pushes it to new remote/branch for software releases",
		Version: version,
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			color.Cyan("Running Release command")

			color.Cyan("\nFetching update from both remotes...")
			r.FetchFromRemote(srcRemote, repositoryPath)
			r.FetchFromRemote(targetRemote, repositoryPath)

			toRun := exec.Command("git", "pull")
			c.RunHeadless(toRun, "git pull", repositoryPath)

			toRun = exec.Command("git", "switch", srcBranch)
			c.RunHeadless(toRun, "git switch srcBranch", repositoryPath)

			toRun = exec.Command("git", "switch", "-c", targetBranch)
			c.RunHeadless(toRun, "git switch -c targetBranch", repositoryPath)

			toRun = exec.Command("git", "push", "-u", targetRemote, targetBranch, "-f")
			c.RunHeadless(toRun, "git push targetRemote targetBranch -f", repositoryPath)
		},
	}

	var deployCmd = &cobra.Command{
		Use:     "deploy",
		Short:   "deploys the Zenith timesheet tool",
		Long:    "wraps around docker compose to deploy the various services",
		Version: version,
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			color.Cyan("Running deploy command")

			d.ResetChanges(repositoryPath)

			command := exec.Command("git", "pull")
			c.RunHeadless(command, "git pull", repositoryPath)

			d.SearchAndReplace(oldUrl, newUrl, repositoryPath)

			command = exec.Command("docker", "compose", "down")
			c.RunHeadless(command, "docker compose down", repositoryPath)

			/* Fixes timeout on build on QA vps: https://stackoverflow.com/a/69432587 */
			os.Setenv("DOCKER_BUILDKIT", "0")
			os.Setenv("COMPOSE_DOCKER_CLI_BUILD", "0")

			toRun := exec.Command("docker", "compose", "up", "--build", "--force-recreate", "-d")
			c.Run(toRun, "docker compose up --build -d", repositoryPath)

			color.Green("\nDeployment done\n\tCurrent logs:\n")
			toRun = exec.Command("docker", "compose", "logs", "-f", "-t")
			c.Run(toRun, "docker compose logs -f -t", repositoryPath)

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

			d.ResetChanges(repositoryPath)
			d.SearchAndReplace(oldUrl, domain, repositoryPath)

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
			toRun := exec.Command("docker", "compose", "run", "--rm", "--entrypoint", entrypoint, "certbot")
			c.RunHeadless(toRun, "docker compose run create key", repositoryPath)

			color.Green("Booting up nginx")
			toRun = exec.Command("docker", "compose", "up", "--force-recreate", "-d", "nginx")
			c.RunHeadless(toRun, "docker compose up nginx -d", repositoryPath)

			color.Green("Deleting dummy certificates")
			entrypoint = fmt.Sprintf("rm -Rf /etc/letsencrypt/live/%s && rm -Rf /etc/letsencrypt/archive/%s && rm -Rf /etc/letsencrypt/renewal/%s.conf", domain, domain, domain)
			toRun = exec.Command("docker", "compose", "run", "--rm", "--entrypoint", entrypoint, "certbot")
			c.RunHeadless(toRun, "docker compose rm certs", repositoryPath)

			color.Green("Requesting real certificates")
			entrypoint = fmt.Sprintf("certbot certonly --webroot -w /var/www/certbot --email %s -d %s --rsa-key-size 4096 --agree-tos --force-renewal --non-interactive", email, domain)
			toRun = exec.Command("docker", "compose", "run", "--rm", "--entrypoint", entrypoint, "certbot")
			c.RunHeadless(toRun, "docker compose run certbot", repositoryPath)

			color.Green("Reloading nginx")
			toRun = exec.Command("docker", "compose", "exec", "nginx", "nginx", "-s", "reload")
			c.RunHeadless(toRun, "docker compose exec nginx nginx -s reload", repositoryPath)

			color.Green("SSL setup done -- rebuilding application")
			toRun = exec.Command("docker", "compose", "down")
			c.RunHeadless(toRun, "docker compose down --rmi local", repositoryPath)
			toRun = exec.Command("docker", "compose", "up", "--build", "-d")
			c.Run(toRun, "docker compose up --build -d", repositoryPath)

			color.Green("\nDeployment done\n\tCurrent logs:\n")
			toRun = exec.Command("docker", "compose", "logs", "-f", "-t")
			c.Run(toRun, "docker compose logs -f -t", repositoryPath)
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

	var nukeCmd = &cobra.Command{
		Use:     "nuke",
		Short:   "nukes docker volumes, images, & everything",
		Long:    "destroys all deployment config for dev purposes",
		Version: version,
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			color.Cyan("Running nuke command")

			color.Green("Stopping and removing all docker images")

			toRun := exec.Command("docker", "compose", "down")
			c.RunHeadless(toRun, "docker compose down", repositoryPath)
			toRun = exec.Command("docker", "rmi", "-f", "$(docker images -aq)")
			c.RunHeadless(toRun, "docker rmi -f $(docker images -aq)", repositoryPath)

			color.Green("Removing volumes and networks")
			toRun = exec.Command("docker", "volume", "rm", "$(docker volume ls -q)")
			c.RunHeadless(toRun, "docker volume prune -f", repositoryPath)
			toRun = exec.Command("docker", "network", "prune", "-f")
			c.RunHeadless(toRun, "docker network prune -f", repositoryPath)

			color.Green("[+] Done")
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
	cloneCmd.Flags().StringVarP(&repoUrl, "url", "u", "https://github.com/zenith-rh/timesheet.git", "git clone url")

	certsCmd.Flags().StringVarP(&domain, "url", "u", "timesheet.zenith-rh.com", "domain of your new environment")
	certsCmd.Flags().StringVarP(&email, "email", "e", "backoffice@zenith-rh.com", "renewal email")

	cloneCmd.AddCommand(deployCmd, releaseCmd, certsCmd, nukeCmd)

	if err := cloneCmd.Execute(); err != nil {
		panic(err)
	}
}
