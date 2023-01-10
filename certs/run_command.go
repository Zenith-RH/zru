package zru

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"

	"github.com/fatih/color"
)

func Run(cmd *exec.Cmd, name string, isInteractive bool, path string) {
	color.Cyan("Running %s", name)

	cmd.Dir = path

	/* Running in one go without any out logging */
	if isInteractive {
		_, err := cmd.CombinedOutput()
		if err != nil {
			color.Red("%s\n", err)
		}
		color.Green("[+] Success\n")
		return
	}

	// Get the stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		color.Red("Could not get StdoutPipe")
		log.Fatal(err)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		color.Red("Could not start command")
		log.Fatal(err)
	}

	// Use a scanner to read the output line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		color.Red("Could not start command")
		log.Fatal(err)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		color.Red("Could not start command")
		log.Fatal(err)
	}

	color.Green("[+] Success")
}
