package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func adminInit() error {
	HOME, _ := os.UserHomeDir()
	CONFIG, _ := os.UserConfigDir()

	// Repo
	REPO := filepath.Join(HOME, "workspace/repos/odoo")
	if _, err := os.Stat(REPO); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(REPO), 0o755)
	}

	// Project
	PROJECT := filepath.Join(HOME, "workspace/odoo")
	if _, err := os.Stat(filepath.Join(PROJECT, "backups")); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(PROJECT, "backups"), 0o755)
	}

	// ODA Config
	ODOOCONF := filepath.Join(CONFIG, "oda", "oda.conf")
	if _, err := os.Stat(ODOOCONF); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(CONFIG, "oda"), 0o755)
		f, err := os.Create(ODOOCONF)
		if err != nil {
			return err
		}
		defer f.Close()
		f.WriteString("REPO=" + REPO + "\n")
		f.WriteString("PROJECT=" + PROJECT + "\n")
		f.WriteString("ODOOBASE=ghcr.io/ppreeper/odoobase" + "\n")
		f.WriteString("BR_ADDR=10.250.250.10" + "\n")
		f.WriteString("DB_HOST=db.local" + "\n")
		f.WriteString("DB_PORT=5432" + "\n")
		f.WriteString("DB_PASS=postgres" + "\n")
		f.WriteString("DB_USERNAME=odoodev" + "\n")
		f.WriteString("DB_USERPASS=odooodoo" + "\n")
	}
	return nil
}

func repoBaseClone() error {
	dirs := GetDirs()
	urlbase := "https://github.com/odoo/"
	repos := []string{"odoo", "enterprise"}
	for _, repo := range repos {
		username, token := getGitHubUsernameToken()
		if err := cloneUrlDir(
			urlbase+repo,
			dirs.Repo, repo,
			username, token,
		); err != nil {
			return fmt.Errorf("odoo repo %s clone failed %w", repo, err)
		}
	}
	return nil
}

func repoBaseUpdate() error {
	dirs := GetDirs()
	repos := []string{"odoo", "enterprise"}
	for _, repo := range repos {

		repoDir := filepath.Join(dirs.Repo, repo)

		repoHeadShortCode, err := repoHeadShortCode(repo)
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

func repoBranchClone() error {
	repoShorts, _ := repoShortCodes("odoo")
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

func repoBranchUpdate() error {
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

func repoHeadShortCode(repo string) (string, error) {
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

func repoShortCodes(repo string) ([]string, error) {
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
