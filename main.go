package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed templates/*
var templateFS embed.FS

type Commit struct {
	Hash    string
	Content string
	Date    string
}

type PageData struct {
	FileName string
	Commits  []Commit
	Active   string
}

func promptForConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/n) ", prompt)
	confirmation, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(strings.ToLower(confirmation)) == "y"
}

func checkGitTracked(fileName string) bool {
	cmd := exec.Command("git", "ls-files", "--error-unmatch", fileName)
	return cmd.Run() == nil
}

func getFileCommits(fileName string) ([]Commit, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%H|%aI", "--", fileName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting commit history: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}

		hash := parts[0]
		date := parts[1]

		cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", hash, fileName))
		content, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("error getting file content from commit %s: %v", hash, err)
		}

		commits = append(commits, Commit{
			Hash:    hash,
			Content: string(content),
			Date:    date,
		})
	}

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

func startServer(fileName string, commits []Commit) error {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		activeHash := r.URL.Query().Get("commit")
		if activeHash == "" && len(commits) > 0 {
			activeHash = commits[0].Hash
		}

		data := PageData{
			FileName: fileName,
			Commits:  commits,
			Active:   activeHash,
		}

		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	port := ":1987"
	fmt.Printf("Starting server at http://localhost%s\n", port)
	return http.ListenAndServe(port, nil)
}

func main() {
	serveFlag := flag.Bool("s", false, "Start web server to view file history")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: natsukashii [-s] <file_name>")
		os.Exit(1)
	}

	fileName := args[0]
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

	if *serveFlag {
		if !promptForConfirmation("Do you want to start the web server to view file history?") {
			fmt.Println("Aborting.")
			os.Exit(0)
		}
		err = startServer(fileName, commits)
		if err != nil {
			fmt.Printf("Error starting server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if !promptForConfirmation(fmt.Sprintf("Do you want to proceed and save all versions in the '%s' directory?", directory)) {
		fmt.Println("Aborting.")
		os.Exit(0)
	}

	err = os.MkdirAll(directory, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	for i, commit := range commits {
		err := saveFileVersion(commit.Hash, fileName, directory, i+1)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Printf("All versions of '%s' have been saved in the '%s' directory.\n", fileName, directory)
}

