package zru

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

func ResetChanges(path string) {
	r, err := git.PlainOpen(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	head, err := r.Head()
	if err != nil {
		color.Red("Could not get HEAD commit")
		log.Fatal(err)
	}

	wt, err := r.Worktree()
	if err != nil {
		color.Red("Could not resolve git worktree")
		log.Fatal(err)
	}

	err = wt.Reset(&git.ResetOptions{
		Commit: head.Hash(),
		Mode:   git.HardReset,
	})
	if err != nil {
		color.Red("Could not run git reset --hard")
		log.Fatal(err)
	}

	fmt.Println("[+] Changes reset to the latest commit")
}

func contains(slice []string, element string) bool {
	for _, e := range slice {
		if strings.Contains(e, element) {
			return true
		}
	}
	return false
}

func SearchAndReplace(search string, replace string, dir string) {

	color.Cyan("Running search and replace -> Searching for %s and replacing with %s", search, replace)
	blacklist := []string{
		".env",
		"zenith_api_config.json",
	}

	dirBlacklist := []string{
		"data",
		"node_modules",
		"build",
		".git",
	}

	// Get the absolute path of the current file
	absPath, err := filepath.Abs(dir)
	if err != nil {
		color.Red("Could not get absolute path")
		fmt.Println(err)
		return
	}

	// Use filepath.Walk to traverse the directory tree
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if contains(dirBlacklist, info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Mode().IsRegular() && !contains(blacklist, info.Name()) {
			file, err := ioutil.ReadFile(path)
			if err != nil {
				color.Red("Could not read file")
				log.Fatalln(err)
			}

			lines := strings.Split(string(file), "\n")
			for i, line := range lines {
				if strings.Contains(line, search) {
					lines[i] = strings.Replace(line, search, replace, -1)
				}
			}

			output := strings.Join(lines, "\n")
			err = ioutil.WriteFile(path, []byte(output), 0644)
			if err != nil {
				color.Red("Could not write to file")
				log.Fatalln(err)
			}

		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	color.Green("Git status:")
	cmd := exec.Command("git", "status")

	cmd.Dir = absPath
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(output))
}
