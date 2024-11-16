package main

import (
	"context"
	"fmt"
	"path/filepath"
	"os"
	"crypto/sha1"
	"strconv"
	"io"

	"github.com/google/go-github/v66/github"
)

// https://gist.github.com/jaredhoward/f231391529efcd638bb7

const (
	owner    = "hclihn"
	repo     = "global_var_func_test_w_local_pkg"
	basePath = "test"
)

var client *github.Client

func main() {
	client = github.NewClient(nil).WithAuthToken(os.Getenv("token"))
	getContents("")
}

func getContents(path string) {
	fmt.Printf("\n\n")

	fileContent, directoryContent, resp, err := client.Repositories.GetContents(context.Background(), owner, repo, path, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("* fileContent: %#v\n", fileContent)
	fmt.Printf("* directoryContent: %#v\n", directoryContent)
	fmt.Printf("* response: %#v\n", resp)

	for _, c := range directoryContent {
		fmt.Println("Type, Path, Size, SHA:", *c.Type, *c.Path, *c.Size, *c.SHA)

		local := filepath.Join(basePath, *c.Path)
		fmt.Println("local:", local)

		switch *c.Type {
		case "file":
			fmt.Printf("-> File %q\n", *c.Path)
			if *c.Path == "mypkg/my_pkg.go" {
				downloadContents(c, local)
			}
		case "dir":
			fmt.Printf("-> Dir %q\n", *c.Path)
			getContents(filepath.Join(path, *c.Path))
		}
	}
}
func downloadContents(content *github.RepositoryContent, localPath string) {
	if content != nil && content.Content != nil {
		fmt.Println("content:", *content.Content)
	}

	rc, resp, err := client.Repositories.DownloadContents(context.Background(), owner, repo, *content.Path, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rc.Close()
	fmt.Printf("* response: %#v\n", resp)
	
	b, err := io.ReadAll(rc)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("File Content:\n%s\n\n", b)
	/*err = os.MkdirAll(filepath.Dir(localPath), 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Writing the file:", localPath)
	f, err := os.Create(localPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	n, err := f.Write(b)
	if err != nil {
		fmt.Println(err)
	}
	if n != *content.Size {
		fmt.Printf("number of bytes differ, %d vs %d\n", n, *content.Size)
	}*/
}

// calculateGitSHA1 computes the github sha1 from a slice of bytes.
// The bytes are prepended with: "blob " + filesize + "\0" before runing through sha1.
func calculateGitSHA1(contents []byte) []byte {
	contentLen := len(contents)
	blobSlice := []byte("blob " + strconv.Itoa(contentLen))
	blobSlice = append(blobSlice, '\x00')
	blobSlice = append(blobSlice, contents...)
	h := sha1.New()
	h.Write(blobSlice)
	bs := h.Sum(nil)
	return bs
}