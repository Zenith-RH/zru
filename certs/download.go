package zru

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetFile(url string, dirpath string, filename string) {
	err := os.MkdirAll(dirpath, os.ModeAppend)
	if err != nil {
		color.Red("Could not create directory %s", dirpath)
		log.Fatal(err)
	}

	fullpath := filepath.Join(dirpath, filename)
	err = downloadFile(fullpath, url)
	if err != nil {
		color.Red("Could not download file %s", url)
		log.Fatal(err)
	}
}
