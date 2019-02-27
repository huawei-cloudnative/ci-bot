package repository

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-github/github"

	"gopkg.in/yaml.v2"
)

var (
	// OwnersFileName
	OwnersFileName = "OWNERS"
)

// NewRepository returns an repository instance
func NewRepository(client *github.Client, repository string) (*Repository, error) {
	// e.g. https://github.com/test/hello
	glog.Infof("New repository : %s", repository)
	if repository == "" {
		return nil, errors.New("Paramter repository is empty")
	}
	// get org and repo
	strs := strings.Split(strings.TrimPrefix(repository, GithubBaseURL), "/")
	if len(strs) >= 2 {
		// e.g. test
		org := strs[0]
		// e.g. hello
		repo := strs[1]
		glog.Infof("New repository org: %s repo: %s", org, repo)
		return &Repository{
			GithubClient: client,
			Org:          org,
			Repo:         repo,
		}, nil
	} else {
		return nil, errors.New("Failed to get org and repo")
	}
}

// OwnersFile defines the content format of owners file
type OwnersFile struct {
	Approvers []string `yaml:"approvers,omitempty"`
	Reviewers []string `yaml:"reviewers,omitempty"`
}

// Interface defines for Owners
type Interface interface {
	// Init repository
	Init() error
	// Clear repository
	Clear() error

	// LoadOwners loads an owners list
	LoadOwners(branch string) error

	// GetApproversFilePath returns the OWNERS file path that contains approvers
	GetApproversFilePath(path string) string
	// GetReviewersFilePath returns the OWNERS file path that contains reviewers
	GetReviewersFilePath(path string) string

	// GetClosestApprovers returns a set of users who are the closest approvers
	GetClosestApprovers(path string) map[string]string
	// GetClosestReviewers returns a set of users who are the closest reviewers
	GetClosestReviewers(path string) map[string]string

	// GetAllApprovers returns all the approvers including parent owners file
	GetAllApprovers(path string) map[string]string
	// GetAllReviewers returns all the reviewers including parent owners file
	GetAllReviewers(path string) map[string]string
}

var _ Interface = &Repository{}

// Repository cache
type Repository struct {
	GithubClient *github.Client
	GitClient    *GitClient

	Org  string
	Repo string

	Sha     string
	RepoDir string

	Approvers map[string]map[string]string
	Reviewers map[string]map[string]string
}

// Init repository
func (o *Repository) Init() error {
	glog.Info("Init repository started.")
	// init repository e.g. test/hello
	gitClient, err := NewGitClient(fmt.Sprintf("%s/%s", o.Org, o.Repo))
	if err != nil {
		glog.Errorf("Failed to new git client: %v", err)
		return err
	}

	// clone mirror
	err = gitClient.CloneMirror()
	if err != nil {
		glog.Errorf("Failed to clone mirror: %v", err)
		return err
	}

	// init
	o.GitClient = gitClient
	return nil
}

// Clear repository
func (o *Repository) Clear() error {
	glog.Info("Clear repository started.")
	// clear mirror
	if o.GitClient != nil {
		err := o.GitClient.RemoveMirror()
		if err != nil {
			glog.Errorf("Failed to remove mirror: %v", err)
			return err
		}
	}
	return nil
}

// LoadOwners loads an owners list
func (o *Repository) LoadOwners(branch string) error {
	// e.g. org=test repo=hello branch=master
	glog.Infof("Load owners started. org: %s repo: %s branch: %s", o.Org, o.Repo, branch)

	// get ref of the repository
	ref, _, err := o.GithubClient.Git.GetRef(
		context.Background(),
		o.Org,
		o.Repo,
		fmt.Sprintf("heads/%s", branch))
	if err != nil {
		glog.Errorf("Failed to get ref: %v", err)
		return err
	}

	// get sha of the ref
	sha := ""
	if ref != nil {
		if ref.Object != nil {
			sha = *ref.Object.SHA
		}
	}
	glog.Infof("Repository sha: %s sha: %s", o.Sha, sha)

	// repository is not changed yet
	if o.Sha != "" && o.Sha == sha {
		glog.Info("Repository is not changed yet")
		return nil
	}

	// clone repository
	err = o.GitClient.CloneRepo()
	if err != nil {
		glog.Errorf("Failed to clone repository: %v", err)
		return err
	}
	defer o.GitClient.RemoveRepo()

	// diff repository
	if o.Sha != "" {
		// get changes between different sha
		changes, err := o.GitClient.Diff(sha, o.Sha)
		if err != nil {
			glog.Errorf("Failed to diff %s with %s", sha, o.Sha)
			return err
		}
		// check if owners files are changed
		ownersChanged := false
		for _, change := range changes {
			if strings.HasSuffix(change, OwnersFileName) {
				ownersChanged = true
				break
			}
		}
		// Owners files are not changed
		if !ownersChanged {
			glog.Info("Owners files are not changed.")
			return nil
		}
	}

	// checkout branch
	err = o.GitClient.CheckOut(branch)
	if err != nil {
		glog.Errorf("Failed to checkout %s", branch)
		return nil
	}

	// Walk
	o.RepoDir = o.GitClient.LocalRepoDir
	o.Sha = sha
	o.Approvers = nil
	o.Reviewers = nil
	return filepath.Walk(o.RepoDir, o.walkFunc)
}

// GetApproversFilePath returns the OWNERS file path that contains approvers
func (o *Repository) GetApproversFilePath(path string) string {
	return o.getOwnersFilePath(path, o.Approvers)
}

// GetReviewersFilePath returns the OWNERS file path that contains reviewers
func (o *Repository) GetReviewersFilePath(path string) string {
	return o.getOwnersFilePath(path, o.Reviewers)
}

// GetClosestApprovers returns a set of users who are the closest approvers
func (o *Repository) GetClosestApprovers(path string) map[string]string {
	return o.getOwners(path, o.Approvers, true)
}

// GetClosestReviewers returns a set of users who are the closest reviewers
func (o *Repository) GetClosestReviewers(path string) map[string]string {
	return o.getOwners(path, o.Reviewers, true)
}

// GetAllApprovers returns all the approvers including parent owners file
func (o *Repository) GetAllApprovers(path string) map[string]string {
	return o.getOwners(path, o.Approvers, false)
}

// GetAllReviewers returns all the reviewers including parent owners file
func (o *Repository) GetAllReviewers(path string) map[string]string {
	return o.getOwners(path, o.Reviewers, false)
}

// filterFilePath format the file path
func (o *Repository) filterFilePath(path string) string {
	if path == "." {
		return ""
	} else {
		return strings.TrimSuffix(path, "/")
	}
}

// walkFunc reads the owners file and loads them into approvers or reviewers
func (o *Repository) walkFunc(path string, info os.FileInfo, err error) error {
	// skip
	if err != nil {
		glog.Error("Error while walking OWNERS files")
		return nil
	}
	if info == nil {
		glog.Error("FileInfo is nil")
		return nil
	}

	// check if it is OWNERS file
	filename := filepath.Base(path)
	if !info.Mode().IsRegular() {
		return nil
	}
	if filename != OwnersFileName {
		return nil
	}

	// read content from OWNERS file
	glog.Infof("WalkFunc path: %s", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Errorf("Failed to read the OWNERS file: %s", path)
		return nil
	}

	relPath, err := filepath.Rel(o.RepoDir, path)
	if err != nil {
		glog.Errorf("Unable to find relative path between repodir: %s and path: %s", o.RepoDir, path)
		return err
	}
	glog.Infof("WalkFunc relPath: %s", relPath)
	relPathDir := o.filterFilePath(filepath.Dir(relPath))
	glog.Infof("WalkFunc relPathDir: %s", relPathDir)

	// unmarshal
	ownersfile := OwnersFile{}
	err = yaml.Unmarshal(b, &ownersfile)
	if err != nil {
		glog.Errorf("Failed to unmarshal %s", path)
		return err
	}

	if len(ownersfile.Approvers) > 0 {
		if o.Approvers == nil {
			o.Approvers = make(map[string]map[string]string)
		}
		if o.Approvers[relPathDir] == nil {
			o.Approvers[relPathDir] = make(map[string]string)
		}
		for _, v := range ownersfile.Approvers {
			o.Approvers[relPathDir][v] = v
		}
	}
	if len(ownersfile.Reviewers) > 0 {
		if o.Reviewers == nil {
			o.Reviewers = make(map[string]map[string]string)
		}
		if o.Reviewers[relPathDir] == nil {
			o.Reviewers[relPathDir] = make(map[string]string)
		}
		for _, v := range ownersfile.Reviewers {
			o.Reviewers[relPathDir][v] = v
		}
	}

	return nil
}

// getOwnersFilePath returns the closest owners file path
func (o *Repository) getOwnersFilePath(path string, maps map[string]map[string]string) string {
	d := path
	if maps != nil {
		for {
			glog.Infof("Get owners file path: %s", d)
			// check if the directory is in maps
			if o, ok := maps[d]; ok {
				// the owners file contains approvers or reviewers
				if len(o) > 0 {
					return d
				}
			}
			// break if the path is root
			if d == "" {
				break
			}
			// get the directory of file or directory path
			d = o.filterFilePath(filepath.Dir(d))
		}
	}
	return ""
}

// getOwners returns a set of users for the requested file.
func (o *Repository) getOwners(path string, maps map[string]map[string]string, closest bool) map[string]string {
	d := path
	out := make(map[string]string)
	if maps != nil {
		for {
			glog.Infof("Get owners: %s closest: %v", d, closest)
			// check if the directory is in maps
			if o, ok := maps[d]; ok {
				for k, v := range o {
					out[k] = v
				}
			}
			// the closest owners are gotten
			if closest && len(out) > 0 {
				break
			}
			// break if the path is root
			if d == "" {
				break
			}
			// get the directory of file or directory path
			d = o.filterFilePath(filepath.Dir(d))
		}
	}
	return out
}
