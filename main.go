package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func checkGitTracked(fileName string) bool {
	cmd := exec.Command("git", "ls-files", "--error-unmatch", fileName)
	return cmd.Run() == nil
}

func getFileCommits(fileName string) ([]string, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%H", "--", fileName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting commit history: %v", err)
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	return commits, nil
}

func saveFileVersion(commit, fileName, directory string, version int) error {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", commit, fileName))
	content, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error getting file content from commit %s: %v", commit, err)
	}

	outPath := filepath.Join(directory, fmt.Sprintf("%s_version_%d", fileName, version))
	err = os.WriteFile(outPath, content, 0644)
	if err != nil {
		return fmt.Errorf("error writing file version: %v", err)
	}

	fmt.Printf("Saved version %d of '%s' (commit %s)\n", version, fileName, commit)
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: natsukashii <file_name>")
		os.Exit(1)
	}

	fileName := os.Args[1]
	directory := "natsukashii"

	if !checkGitTracked(fileName) {
		fmt.Printf("Error: File '%s' is not tracked by git.\n", fileName)
		os.Exit(1)
	}

	commits, err := getFileCommits(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	commitCount := len(commits)
	fmt.Printf("The file '%s' has been modified in %d commits.\n", fileName, commitCount)

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Do you want to proceed and save all versions in the '%s' directory? (y/n) ", directory)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	if strings.TrimSpace(strings.ToLower(confirmation)) != "y" {
		fmt.Println("Aborting.")
		os.Exit(0)
	}

	err = os.MkdirAll(directory, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	for i, commit := range commits {
		err := saveFileVersion(commit, fileName, directory, i+1)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Printf("All versions of '%s' have been saved in the '%s' directory.\n", fileName, directory)
}

