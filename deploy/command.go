package zru

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"

	"github.com/fatih/color"
)

func Deploy(path string) {
	cmd := exec.Command("docker-compose", "up", "--build")
	cmd.Dir = path

	color.Cyan("Running docker-compose up --build")

	// Get the stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		return
	}

	// Use a scanner to read the output line by line
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	color.Cyan("Deployment successful")
}
