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

func RunHeadless(cmd *exec.Cmd, name string, path string) {
	color.Cyan("Running %s", name)

	cmd.Dir = path

	_, err := cmd.CombinedOutput()
	if err != nil {
		color.Red("%s\n", err)
	}
	color.Green("[+] Success\n")

}

func Run(cmd *exec.Cmd, name string, path string) {
	color.Cyan("Running %s", name)
	cmd.Dir = path

	color.Cyan("Running command interactively. Press ENTER to detach...")

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

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-done:
			color.Green("[+] Command finished successfully.")
			return
		default:
		}

		char, _, err := reader.ReadRune()
		if err != nil {
			color.Red("Error reading user input")
			log.Fatal(err)
		}
		if char == ' ' {
			cmd.Process.Release()
			return
		}

		stdout.Close()
		stderr.Close()
	}
}
