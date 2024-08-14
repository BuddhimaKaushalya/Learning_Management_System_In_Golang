package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DeleteFileByURL(fileURL string) error {
	// Extract the file path from the URL
	parts := strings.Split(fileURL, "/")
	println(parts)

	if len(parts) > 2 && parts[3] == "static" {
		parts[3] = "uploads"
	}
	newURl := strings.Join(parts, "/")
	println(newURl)

	//trim the server address and port 8080
	trimmedPath := filepath.Join("", filepath.Clean(strings.TrimPrefix(newURl, "http://172.17.249.61:8080")))
	println(trimmedPath)

	newpath := "." + trimmedPath
	// Delete the file
	err := os.Remove(newpath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found")
		}
		return err
	}
	println("success")

	return nil
}
