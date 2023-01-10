package zru

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

func Run(cmd *exec.Cmd, name string, isInteractive bool, path string) {
	color.Cyan("Running %s", name)

	cmd.Dir = path

	/* Running in one go without any out logging */
	if isInteractive == true {
		_, err := cmd.CombinedOutput()
		if err != nil {
			color.Red("%s\n", err)
		}
		color.Green("[+] Success\n")
		return
	}

	color.Cyan("Running command interactively. Press SPACE to detach...")

	/* Getting output pipes */

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		color.Red("Error stdin pipe")
		log.Fatal(err)

		fmt.Println(err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		color.Red("Error getting stderr pipe")
		log.Fatal(err)

		fmt.Println(err)
		return
	}

	/* starting command & storing outputs in pipes */

	if err := cmd.Start(); err != nil {
		color.Red("Error starting command")
		log.Fatal(err)
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	/* read input with bufio */

	reader := bufio.NewReader(os.Stdin)
	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			color.Red("Error reading user input")
			log.Fatal(err)
		}
		if char == ' ' {
			/* detaching from process */
			cmd.Process.Release()
			return
		}

		stdout.Close()
		stderr.Close()

		color.Green("[+] Success!")
	}
}
