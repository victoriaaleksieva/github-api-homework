package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type User struct {
	Login       string    `json:"login"`
	Name        string    `json:"name"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	PublicRepos int       `json:"public_repos"`
	Followers   int       `json:"followers"`
	Following   int       `json:"following"`
	CreatedAt   time.Time `json:"created_at"`
}
type Repos []struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

var client = &http.Client{Timeout: 10 * time.Second}

func fetch(url string, target any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "homework-go")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func readFromFile(input string) []string {
	file, err := os.Open(input)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	output := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			output = append(output, word)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return output
}
func fetchUser(username string) (User, error) {
	var user User
	url := fmt.Sprintf("https://api.github.com/users/%s", username)
	err := fetch(url, &user)

	return user, err
}
func fetchRepos(username string) (Repos, error) {
	var repo Repos
	url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
	err := fetch(url, &repo)
	return repo, err
}

func fetchLang(username string, repo string) (map[string]int, error) {
	var languages map[string]int
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/languages", username, repo)
	err := fetch(url, &languages)
	return languages, err
}
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
func formatLangs(langs map[string]int) string {
	if len(langs) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(langs))
	for lang := range langs {
		keys = append(keys, lang)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, lang := range keys {
		parts = append(parts, fmt.Sprintf("%s (%d)", lang, langs[lang]))
	}
	return strings.Join(parts, ", ")
}
func printUser(u User) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "LOGIN\tNAME\tLOCATION\tREPOS\tFOLLOWERS\tCREATED")
	fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\n",
		u.Login, orDash(u.Name), orDash(u.Location),
		u.PublicRepos, u.Followers, u.CreatedAt.Year())
	w.Flush()
}
func main() {

	userArr := readFromFile("test")

	for _, name := range userArr {
		u, err := fetchUser(name)
		if err != nil {
			fmt.Println("skipping", name+":", err)
			continue
		}
		fmt.Println()
		printUser(u)

		repos, err := fetchRepos(name)
		if err != nil {
			fmt.Println("could not fetch repos for", name+":", err)
			continue
		}

		fmt.Printf("\nRepositories of %s (%d)\n", name, len(repos))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "REPO\tLANGUAGES")
		for _, repo := range repos {
			langs, err := fetchLang(name, repo.Name)
			if err != nil {
				fmt.Fprintf(w, "%s\t(error: %v)\n", repo.Name, err)
				continue
			}
			fmt.Fprintf(w, "%s\t%s\n", repo.Name, formatLangs(langs))
		}
		w.Flush()
	}

}
