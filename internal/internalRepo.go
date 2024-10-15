package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/viper"
)

func CloneUrlDir(url, baseDir, cloneDir, username, token string) error {
	_, err := os.Stat(filepath.Join(baseDir, cloneDir, ".git"))
	if os.IsNotExist(err) {
		os.MkdirAll(baseDir, 0o755)
		_, err = git.PlainClone(filepath.Join(baseDir, cloneDir), false, &git.CloneOptions{
			URL:      url,
			Progress: os.Stdout,
			Auth: &http.BasicAuth{
				Username: username,
				Password: token,
			},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error cloning repo %v\n", err)
		}
	}
	return nil
}

func RepoHeadShortCode(repo string) (string, error) {
	repoDir := filepath.Join(viper.GetString("dirs.repo"), repo)

	r, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", err
	}

	refs, err := r.References()
	if err != nil {
		return "", err
	}
	var refList string
	if err := refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.SymbolicReference {
			refList = ref.Target().Short()
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("refs.ForEach on %s %w", repo, err)
	}
	return refList, nil
}

func RepoShortCodes(repo string) ([]string, error) {
	repoDir := filepath.Join(viper.GetString("dirs.repo"), repo)

	r, err := git.PlainOpen(repoDir)
	if err != nil {
		return []string{}, err
	}

	refs, err := r.References()
	if err != nil {
		return []string{}, err
	}
	refList := []float64{}
	if err := refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.SymbolicReference &&
			strings.HasPrefix(ref.Name().Short(), "origin") {
			refSplit := strings.Split(ref.Name().Short(), "/")
			shortName := refSplit[len(refSplit)-1]
			if !strings.HasPrefix(shortName, "master") &&
				!strings.HasPrefix(shortName, "staging") &&
				!strings.HasPrefix(shortName, "saas") &&
				!strings.HasPrefix(shortName, "tmp") {
				val, _ := strconv.ParseFloat(shortName, 32)
				refList = append(refList, val)
			}
		}
		return nil
	}); err != nil {
		return []string{}, fmt.Errorf("refs.ForEach on %s %w", repo, err)
	}
	slices.Sort(refList)
	slices.Reverse(refList)
	shortRefs := []string{}
	for _, ref := range refList[0:4] {
		shortRefs = append(shortRefs, strconv.FormatFloat(ref, 'f', 1, 64))
	}
	return shortRefs, nil
}
