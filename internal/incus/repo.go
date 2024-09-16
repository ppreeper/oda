package oda

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func RepoBaseClone() error {
	dirs := GetDirs()
	urlbase := "https://github.com/odoo/"
	repos := []string{"odoo", "enterprise"}
	for _, repo := range repos {
		username, token := GetGitHubUsernameToken()
		if err := CloneUrlDir(
			urlbase+repo,
			dirs.Repo, repo,
			username, token,
		); err != nil {
			return fmt.Errorf("odoo repo %s clone failed %w", repo, err)
		}
	}
	return nil
}

func RepoBaseUpdate() error {
	dirs := GetDirs()
	repos := []string{"odoo", "enterprise"}
	for _, repo := range repos {

		repoDir := filepath.Join(dirs.Repo, repo)

		repoHeadShortCode, err := RepoHeadShortCode(repo)
		if err != nil {
			return fmt.Errorf("repoHeadShortCode %w", err)
		}

		fetch := exec.Command("git", "fetch", "origin")
		fetch.Dir = repoDir
		if err := fetch.Run(); err != nil {
			return fmt.Errorf("git fetch origin on %s %w", repo, err)
		}

		checkout := exec.Command("git", "checkout", repoHeadShortCode)
		checkout.Dir = repoDir
		if err := checkout.Run(); err != nil {
			return fmt.Errorf("git checkout on %s %w", repo, err)
		}

		pull := exec.Command("git", "pull", "origin", repoHeadShortCode)
		pull.Dir = repoDir
		if err := pull.Run(); err != nil {
			return fmt.Errorf("git pull on %s %w", repo, err)
		}
	}
	return nil
}

func RepoBranchClone() error {
	repoShorts, _ := RepoShortCodes("odoo")
	versions := GetCurrentOdooRepos()
	for _, version := range versions {
		repoShorts = removeValue(repoShorts, version)
	}
	versionOptions := []huh.Option[string]{}
	for _, version := range repoShorts {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}
	var version string
	var create bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Available Odoo Branches").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Clone Branch?").
				Value(&create),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("odoo version form error %w", err)
	}

	dirs := GetDirs()
	repos := []string{"odoo", "enterprise"}

	for _, repo := range repos {
		source := filepath.Join(dirs.Repo, repo)
		dest := filepath.Join(dirs.Repo, version, repo)
		if err := CopyDirectory(source, dest); err != nil {
			return fmt.Errorf("copy directory %s to %s failed %w", source, dest, err)
		}

		fetcher := exec.Command("git", "fetch", "origin")
		fetcher.Dir = dest
		if err := fetcher.Run(); err != nil {
			return fmt.Errorf("git fetch origin on %s %w", repo, err)
		}

		checkout := exec.Command("git", "checkout", version)
		checkout.Dir = dest
		if err := checkout.Run(); err != nil {
			return fmt.Errorf("git checkout on %s %w", repo, err)
		}

		pull := exec.Command("git", "pull", "origin", version)
		pull.Dir = dest
		if err := pull.Run(); err != nil {
			return fmt.Errorf("git pull on %s %w", repo, err)
		}
	}
	return nil
}

func RepoBranchUpdate() error {
	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}
	var version string
	var create bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Available Odoo Branches").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Update Branch?").
				Value(&create),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("odoo version form error %w", err)
	}

	dirs := GetDirs()
	repos := []string{"odoo", "enterprise"}

	for _, repo := range repos {
		dest := filepath.Join(dirs.Repo, version, repo)

		fetcher := exec.Command("git", "fetch", "origin")
		fetcher.Dir = dest
		if err := fetcher.Run(); err != nil {
			return fmt.Errorf("git fetch origin on %s %w", repo, err)
		}

		checkout := exec.Command("git", "checkout", version)
		checkout.Dir = dest
		if err := checkout.Run(); err != nil {
			return fmt.Errorf("git checkout on %s %w", repo, err)
		}

		pull := exec.Command("git", "pull", "origin", version)
		pull.Dir = dest
		if err := pull.Run(); err != nil {
			return fmt.Errorf("git pull on %s %w", repo, err)
		}
	}
	return nil
}

func RepoHeadShortCode(repo string) (string, error) {
	dirs := GetDirs()
	repoDir := filepath.Join(dirs.Repo, repo)

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
	dirs := GetDirs()
	repoDir := filepath.Join(dirs.Repo, repo)

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
