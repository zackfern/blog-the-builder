package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func downloadArchive(archiveLink *url.URL) string {
	downloadDestination := fmt.Sprintf("%s-master.tar.gz", RepoName)

	// Create the file on disk
	output, err := os.Create(downloadDestination)
	handleError(err)
	defer output.Close()

	resp, err := http.Get(archiveLink.String())
	handleError(err)
	defer resp.Body.Close()

	written, err := io.Copy(output, resp.Body)
	handleError(err)

	fmt.Println("Downloaded", written, "bytes to", downloadDestination)
	return downloadDestination
}

func extractArchive(tarPath string) string {
	cmd := exec.Command("tar", "-xzf", tarPath)
	err := cmd.Run()
	handleError(err)

	matches, err := filepath.Glob(fmt.Sprintf("%s-%s-*", RepoOwner, RepoName))
	handleError(err)
	folderName := matches[0]

	fmt.Println("Extracted", tarPath, "to", folderName)
	return folderName
}
