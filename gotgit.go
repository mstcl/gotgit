// Package gotgit is the main thang
package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type repo struct {
	name     string
	path     string
	cmtCnt   int
	uChanges changes
	sChanges changes
}

type changes struct {
	added    int
	modified int
	deleted  int
	others   int
}

func getGitDirs() []string {
	out, err := exec.Command("/usr/bin/sh", "./scripts/find_git_dirs.sh").Output()

	if err != nil {
		fmt.Printf("%s", err)
	}
	output := string(out)
	dirs := strings.Split(output, "\n")
	return dirs
}

func fetchRepo(r *git.Repository) {
	fetchRes := r.Fetch(&git.FetchOptions{RemoteName: "origin"})
	if fetchRes != nil && fetchRes.Error() != git.NoErrAlreadyUpToDate.Error() {
		fmt.Println(fetchRes.Error())
	}
}

// Get the status, return status plus status error
func getStats(path string) git.Status {
	r, err := git.PlainOpen(path)
	if err != nil {
		fmt.Printf("%s", err)
		return nil
	}
	wTree, err := r.Worktree()
	if err != nil {
		fmt.Printf("%s", err)
		return nil
	}
	status, err := wTree.Status()
	if err != nil {
		fmt.Printf("%s", err)
		return nil
	}
	return status
}

// Extract status code from 0th (1st) column: status of staging area (worktree)
func extractChanges(status git.Status, c changes, col int) changes {
	files := strings.Split(status.String(), "\n")
	for idx := 0; idx < len(files)-1; idx++ {
		sCode := string(files[idx][col])
		switch sCode {
		case "?":
			continue
		case "A":
			c.added++
		case "D":
			c.deleted++
		case "M":
			c.modified++
		default:
			c.others++
		}
	}
	return c
}

func getRef(r *git.Repository) *plumbing.Reference {
	ref, err := r.Head()
	if err != nil {
		fmt.Printf("%s", err)
	}
	return ref
}

func getTotCmts(r *git.Repository, hash plumbing.Hash) int {
	cmtIter, err := r.Log(&git.LogOptions{From: hash})
	if err != nil {
		fmt.Printf("%s", err)
		return 0
	}
	var cmtCnt int
	err = cmtIter.ForEach(func(c *object.Commit) error {
		cmtCnt++
		return nil
	})
	if err != nil {
		fmt.Printf("%s", err)
		return 0
	}
	return cmtCnt
}

func getRepos(dirsSorted []string) []repo {
	repos := make([]repo, len(dirsSorted))
	for idx := 0; idx < len(repos); idx++ {
		dirSlc := strings.Split(dirsSorted[idx], "/")
		// path := strings.Join([]string{"~", strings.Join(dirSlc[3:], "/")}, "/")
		name := dirSlc[len(dirSlc)-1]
		fmt.Println(name)
		repos[idx] = repo{
			name:     name,
			path:     dirsSorted[idx],
		}
	}
	return repos
}

func processRepo(r repo) repo {
		gr, err := git.PlainOpen(r.path)
		if err != nil {
			fmt.Printf("%s", err)
		}
		ref := getRef(gr)
		u := changes{added: 0, modified: 0, deleted: 0, others: 0}
		s := changes{added: 0, modified: 0, deleted: 0, others: 0}
		status := getStats(r.path)
		if status != nil {
			s = extractChanges(status, u, 0)
			u = extractChanges(status, u, 1)
		}
		totCmts := getTotCmts(gr, ref.Hash())
		r.cmtCnt = totCmts
		r.uChanges = u
		r.sChanges = s
	return r
}
