package repository

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/golang/glog"
)

var (
	// GithubBaseURL
	GithubBaseURL = "https://github.com/"
)

// NewGitClient returns a git client. e.g. repo=test/hello
func NewGitClient(repo string) (*GitClient, error) {
	// it is used to store mirror. e.g. tmpMirrorDir=/tmp/mirror170610160
	tmpMirrorDir, err := ioutil.TempDir("", "mirror")
	if err != nil {
		glog.Errorf("Failed to create tmp mirror dir %s: %v", repo, err)
		return nil, err
	}

	return &GitClient{
		LocalMirrorDir: tmpMirrorDir,
		BaseURL:        GithubBaseURL,
		Repo:           repo,
	}, nil
}

// GitClient define
type GitClient struct {
	// local mirror dir
	LocalMirrorDir string
	// local repository dir
	LocalRepoDir string
	// github base url
	BaseURL string
	// repository org and name
	Repo string
}

// CloneMirror clones a mirror into local tmp folder
func (r *GitClient) CloneMirror() error {
	// e.g. localMirror=/tmp/mirror170610160/test/hello
	localMirror := filepath.Join(r.LocalMirrorDir, r.Repo)
	// e.g. remote=https://github.com/test/hello
	remote := fmt.Sprintf("%s%s", r.BaseURL, r.Repo)
	glog.Infof("Clone localMirror: %s from remote: %s", localMirror, remote)

	// check if localMirror is exsiting
	_, err := os.Stat(localMirror)
	if err != nil {
		if os.IsNotExist(err) {
			glog.Infof("Clone mirror: %s", remote)
			// localMirror is not exsiting
			err = os.MkdirAll(filepath.Dir(localMirror), os.ModePerm)
			if err != nil {
				if !os.IsExist(err) {
					// mkdir error
					glog.Errorf("Failed to mkdir %s: %v", localMirror, err)
					return err
				}
			}

			// clone mirror
			bs, err := exec.Command("git", "clone", "--mirror", remote, localMirror).CombinedOutput()
			if err != nil {
				// clone mirror error
				glog.Errorf("Failed to clone mirror %s error: %v", string(bs), err)
				return err
			}
		} else {
			// other error
			glog.Errorf("Failed to stat %s: %v", localMirror, err)
			return err
		}
	} else {
		glog.Infof("Fetch mirror: %s", r.Repo)
		c := exec.Command("git", "fetch")
		c.Dir = localMirror
		bs, err := c.CombinedOutput()
		if err != nil {
			// fetch error
			glog.Errorf("Failed to fetch mirror %s error: %v", string(bs), err)
			return err
		}
	}

	return nil
}

// CloneRepo clones a repository into local tmp folder
func (r *GitClient) CloneRepo() error {
	// e.g. localMirror=/tmp/mirror170610160/test/hello
	localMirror := filepath.Join(r.LocalMirrorDir, r.Repo)

	// clone mirror
	err := r.CloneMirror()
	if err != nil {
		// clone mirror error
		glog.Errorf("Failed to clone mirror: %v", err)
		return err
	}

	// it is used to store repository. e.g. tmpRepoDir=/tmp/repo170610160
	r.LocalRepoDir, err = ioutil.TempDir("", "repo")
	if err != nil {
		glog.Errorf("Failed to create tmp repository dir: %v", err)
		return err
	}

	// clone repository
	glog.Infof("Clone repository from localMirror: %s to LocalRepoDir: %s", localMirror, r.LocalRepoDir)
	bs, err := exec.Command("git", "clone", localMirror, r.LocalRepoDir).CombinedOutput()
	if err != nil {
		// clone repository error
		glog.Errorf("Failed to clone repository from localMirror %s error: %v", string(bs), err)
		return err
	}

	return nil
}

// RemoveMirror removes tmp local mirror dir
func (r *GitClient) RemoveMirror() error {
	// remove local mirror dir
	glog.Infof("Remove mirror: %s", r.LocalMirrorDir)
	err := os.RemoveAll(r.LocalMirrorDir)
	if err != nil {
		// remove error
		glog.Errorf("Failed to remove mirror: %v", err)
		return err
	}
	return nil
}

// RemoveRepo removes tmp local repository dir
func (r *GitClient) RemoveRepo() error {
	// remove local repo dir
	glog.Infof("Remove repository: %s", r.LocalRepoDir)
	err := os.RemoveAll(r.LocalRepoDir)
	if err != nil {
		// remove error
		glog.Errorf("Failed to remove repository: %v", err)
		return err
	}
	return nil
}

// CheckOut runs the command: git checkout
func (r *GitClient) CheckOut(branch string) error {
	glog.Infof("Checkout: %s", branch)
	bs, err := exec.Command("git", "checkout", branch).CombinedOutput()
	if err != nil {
		// git checkout error
		glog.Errorf("Failed to git checkout %s: %v", string(bs), err)
		return err
	}
	return err
}

// Diff runs the command: git diff
func (r *GitClient) Diff(head string, sha string) (changes []string, err error) {
	glog.Infof("Diff head %s sha %s", head, sha)
	bs, err := exec.Command("git", "diff", head, sha, "--name-only").CombinedOutput()
	if err != nil {
		// git diff error
		glog.Errorf("Failed to git diff %s: %v", string(bs), err)
		return nil, err
	}
	scan := bufio.NewScanner(bytes.NewReader(bs))
	scan.Split(bufio.ScanLines)
	for scan.Scan() {
		changes = append(changes, scan.Text())
	}
	return changes, err
}
