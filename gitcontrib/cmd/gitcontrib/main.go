package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type statistics struct {
	commits int
	filesChanged int
	insertions int
	deletions int
	name string
	email string
}

func (s *statistics) String() string  {
	return fmt.Sprintf("Name: %s\nEmail: %s\nCommits: %d\nFiles Changed: %d\nInsertions: %d\nDeletions: %d\nDelta: %d\n", s.name, s.email, s.commits, s.filesChanged, s.insertions, s.deletions, s.insertions-s.deletions)
}

func main() {
	var repoPath string
	flag.StringVar(&repoPath, "repo", "", "path of repo to pull statistics for.")
	flag.Parse()
	c := exec.Command("git", "-C", repoPath, "log", "--no-merges", "--shortstat", "--format=name:<%an>,email:<%ae>")
	fmt.Println(strings.Join(c.Args, " "))
	so, err := c.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := c.Start(); err != nil {
		panic(err)
	}
	r := bufio.NewReader(so)
	scanner := bufio.NewScanner(r)
	results := make(map[string]*statistics, 1000)
	var currentStat *statistics
	var ok bool
	for scanner.Scan() {
		data := bytes.TrimSpace(scanner.Bytes())
		if len(data) == 0 {
			continue
		}
		if bytes.HasPrefix(data, []byte("name")) {
			split := bytes.Split(data, []byte(","))
			authorName := string(bytes.Split(split[0], []byte(":"))[1])
			authorEmail := string(bytes.Split(split[1], []byte(":"))[1])
			currentStat, ok = results[authorEmail]
			if !ok {
				currentStat = &statistics{
					name:authorName,
					email:authorEmail,
				}
				results[authorEmail] = currentStat
			}
			currentStat.commits ++
		} else if bytes.Contains(data, []byte("files changed")){
			err = currentStat.read(data)
			if err != nil {
				panic(err)
			}
		}
	}
	err = c.Wait()
	if err != nil {
		panic(err)
	}
	for _, s := range results {
		fmt.Println(s)
	}
}

func (s *statistics) read( data []byte) error  {
	const (
		filesChanged = "files changed"
		insertions = "insertions(+)"
		deletions = "deletions(-)"
	)
	split := bytes.Split(data, []byte(","))
	for _, sp := range split {
		vals := bytes.SplitN(bytes.TrimSpace(sp), []byte(" "), 2)
		tipe := string(bytes.TrimSpace(vals[1]))
		conv, err := strconv.Atoi(string(vals[0]))
		if err != nil {
			return err
		}
		switch tipe {
		case filesChanged:
			s.filesChanged += conv
		case insertions:
			s.insertions += conv
		case deletions:
			s.deletions += conv
		}
	}
	return nil
}
