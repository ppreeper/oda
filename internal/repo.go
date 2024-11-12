package internal

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/charmbracelet/huh"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/ppreeper/oda/config"
	"github.com/ppreeper/oda/ui"
)

// RepoBaseClone
// setup odoo,enterprise,design-themes,industry
// source repositories in repos/odoo directory
func (o *ODA) RepoBaseClone() error {
	odaConf, _ := config.LoadOdaConfig()
	repoDir := odaConf.Dirs.Repo
	latestBranch := config.GetBranchLatest()

	for _, repo := range latestBranch.Repos {
		fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Cloning", repo))
		username, token := GetGitHubUsernameToken()
		odooURL, err := url.JoinPath(config.OdooBaseURL, repo)
		if err != nil {
			return fmt.Errorf("url.JoinPath %w", err)
		}
		err = CloneUrlDir(
			odooURL,
			repoDir, repo,
			username, token,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, ui.ErrorStyle.Render("odoo repo %s clone err %v"), repo, err)
		}
		fmt.Fprintln(os.Stderr, "")
	}

	return nil
}

// RepoBaseUpdate
// update odoo,enterprise,design-themes,industry
// source repositories in repos/odoo directory
func (o *ODA) RepoBaseUpdate() error {
	odaConf, _ := config.LoadOdaConfig()
	repoDir := odaConf.Dirs.Repo
	latestBranch := config.GetBranchLatest()

	for _, repo := range latestBranch.Repos {
		fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Updating", repo))

		dest := filepath.Join(repoDir, repo)

		repoHeadShortCode, err := RepoHeadShortCode(repo)
		if err != nil {
			return fmt.Errorf("repoHeadShortCode %w", err)
		}

		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render("git", "fetch", "origin"))
		fetch := exec.Command("git", "fetch", "origin")
		fetch.Dir = dest
		if err := fetch.Run(); err != nil {
			return fmt.Errorf("git fetch origin on %s %w", repo, err)
		}

		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render("git", "checkout", repoHeadShortCode))
		checkout := exec.Command("git", "checkout", repoHeadShortCode)
		checkout.Dir = dest
		if err := checkout.Run(); err != nil {
			return fmt.Errorf("git checkout on %s %w", repo, err)
		}

		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render("git", "pull", "origin", repoHeadShortCode))
		pull := exec.Command("git", "pull", "origin", repoHeadShortCode)
		pull.Dir = dest
		if err := pull.Run(); err != nil {
			return fmt.Errorf("git pull on %s %v", repo, err)
		}

		fmt.Fprintln(os.Stderr, "")
	}

	return nil
}

// RepoBranchClone
// setup odoo,enterprise,design-themes,industry
// branch repositories in repos/odoo directory
// repos/odoo/<branch> ex 17.0 > 17.0
// repos/odoo/<branch> ex 17.2 > saas-17.2
// mkdir -p repos/odoo/<branch>
// copy repos/odoo/{odoo,enterprise,design-themes,industry}
// to repos/odoo/<branch>/{odoo,enterprise,design-themes,industry}
// go to each directory and git checkout <branch>
func (o *ODA) RepoBranchClone() error {
	var version string
	var create bool

	repoShorts, _ := RepoShortCodes("odoo")
	versions := GetCurrentOdooRepos()
	for _, version := range versions {
		repoShorts = removeValue(repoShorts, version)
	}
	versionOptions := []huh.Option[string]{}
	for _, version := range repoShorts {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}

	if len(versionOptions) == 0 {
		return fmt.Errorf("no more branches to clone")
	}

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

	if !create {
		return nil
	}

	odaConf, _ := config.LoadOdaConfig()
	repoDir := odaConf.Dirs.Repo
	branch := config.GetVersion(version)
	for _, repo := range branch.Repos {
		fmt.Fprintln(os.Stderr, ui.StepStyle.Render("Cloning", repo))
		source := filepath.Join(repoDir, repo)
		dest := filepath.Join(repoDir, version, repo)
		fmt.Fprintf(os.Stderr, ui.SubStepStyle.Render("copying from %s base")+"\n", repo)
		if err := CopyDirectory(source, dest); err != nil {
			return fmt.Errorf("copy directory %s to %s failed %w", source, dest, err)
		}

		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render("git", "fetch", "origin"))
		fetcher := exec.Command("git", "fetch", "origin")
		fetcher.Dir = dest
		fetcher.Stdout = os.Stdout
		if err := fetcher.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "git fetch origin on %s %v", repo, err)
		}

		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render("git", "checkout", version))
		checkout := exec.Command("git", "checkout", version)
		checkout.Dir = dest
		checkout.Stdout = os.Stdout
		if err := checkout.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "git checkout on %s %v", repo, err)
		}

		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render("git", "pull", "origin", version))
		pull := exec.Command("git", "pull", "origin", version)
		pull.Dir = dest
		pull.Stdout = os.Stdout
		if err := pull.Run(); err != nil {
			fmt.Fprintf(os.Stderr, ui.ErrorStyle.Render("git pull on %s %v"), repo, err)
		}
		fmt.Fprintln(os.Stderr, "")
	}

	return nil
}

// RepoBranchUpdate
// update odoo,enterprise,design-themes,industry
// source repositories in repos/odoo/<branch> directory
func (o *ODA) RepoBranchUpdate() error {
	fmt.Println("RepoBranchUpdate")

	versions := GetCurrentOdooRepos()
	versionOptions := []huh.Option[string]{}
	for _, version := range versions {
		versionOptions = append(versionOptions, huh.NewOption(version, version))
	}
	var version string
	var update bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Available Odoo Branches").
				Options(versionOptions...).
				Value(&version),

			huh.NewConfirm().
				Title("Update Branch?").
				Value(&update),
		),
	)
	if err := form.Run(); err != nil {
		return fmt.Errorf("odoo version form error %e", err)
	}
	if !update {
		return nil
	}

	odaConf, _ := config.LoadOdaConfig()
	repoDir := odaConf.Dirs.Repo

	odooBranch := config.GetVersion(version)

	for _, repo := range odooBranch.Repos {
		dest := filepath.Join(repoDir, version, repo)
		fmt.Fprintf(os.Stderr, ui.StepStyle.Render("Updating %s %s")+"\n", repo, version)
		fmt.Fprintln(os.Stderr, ui.SubStepStyle.Render("git", "pull", "--rebase"))
		pull := exec.Command("git", "pull", "--rebase")
		pull.Dir = dest
		if err := pull.Run(); err != nil {
			fmt.Fprintf(os.Stderr, ui.ErrorStyle.Render("git pull on %s %v")+"\n", repo, err)
		}
		fmt.Fprintln(os.Stderr, "")
	}

	return nil
}

// ####################################

func CloneUrlDir(url, baseDir, cloneDir, username, token string) error {
	targetDir := filepath.Join(baseDir, cloneDir)
	_, err := os.Stat(filepath.Join(targetDir, ".git"))
	if os.IsNotExist(err) {
		os.MkdirAll(baseDir, 0o755)
		_, err = git.PlainClone(targetDir, false, &git.CloneOptions{
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
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return "", err
	}

	repoDir := filepath.Join(odaConf.Dirs.Repo, repo)

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
	odaConf, err := config.LoadOdaConfig()
	if err != nil {
		return []string{}, err
	}

	repoDir := filepath.Join(odaConf.Dirs.Repo, repo)

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

// From: https://stackoverflow.com/questions/51779243/copy-a-folder-in-go

func CopyDirectory(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return fmt.Errorf("failed to list directory: '%s', error: '%s'", scrDir, err.Error())
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to get file info for '%s'", sourcePath)
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0o755); err != nil {
				return fmt.Errorf("failed to create directory: '%s', error: '%s'", destPath, err.Error())
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return fmt.Errorf("failed to copy directory: '%s', error: '%s'", sourcePath, err.Error())
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return fmt.Errorf("failed to copy symlink: '%s', error: '%s'", sourcePath, err.Error())
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return fmt.Errorf("failed to copy file: '%s', error: '%s'", sourcePath, err.Error())
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return fmt.Errorf("failed to change owner for '%s'", destPath)
		}

		fInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info for '%s'", sourcePath)
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fInfo.Mode()); err != nil {
				return fmt.Errorf("failed to change mode for '%s'", destPath)
			}
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("failed to create file: '%s', error: '%s'", dstFile, err.Error())
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open file: '%s', error: '%s'", srcFile, err.Error())
	}

	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("failed to copy file: '%s', error: '%s'", srcFile, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return fmt.Errorf("failed to read symlink: '%s', error: '%s'", source, err.Error())
	}
	return os.Symlink(link, dest)
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}
