package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var CfgFile string

const (
	EnvGopath = "gopath"

	GitUser      = "git.user"
	GitHost      = "git.host"
	GitProject   = "git.project"
	SourceBranch = "git.default-branch"

	// User, Host, Project, Repo
	CloneSSHURLTmpl = "ssh://%s@%s/%s/%s.git"

	SortRepos        = "repos.sort"
	RepoAliases      = "repos.aliases"
	DefaultReviewers = "repos.reviewers"
	CatalogCacheFile = "repos.cache.filename"
	CatalogCacheTTL  = "repos.cache.ttl"

	CommitAmend   = "commit.amend"
	CommitMessage = "commit.message"

	Branch    = "branch"
	Reviewers = "reviewers"
	AuthToken = "auth-token"
	UseSync   = "sync"

	ChannelBuffer = "channels.buffer-size"

	// Bitbucket v1 API PR template - Host, Project, Repo
	ApiPathTmpl = "https://%s/rest/api/1.0/projects/%s/repos/%s/pull-requests"
	PrTmpl      = `{
	"title": "%s",
	"description": "%s",
	"fromRef": {
		"id": "refs/heads/%s",
		"repository": %s
	},
	"toRef": {
		"id": "refs/heads/%s",
		"repository": %s
	},
	"reviewers": [%s]
}`
	PrRepoTmpl = `{
			"slug": "%s",
			"project": {"key": "%s"}
		}`
	PrReviewerTmpl = `{
		"user": {"name": "%s"}
	}`
)

// Init reads in config file and ENV variables if set.
func Init() {
	// Default user for SSH clone.
	viper.SetDefault(GitUser, "git")

	viper.SetDefault(SourceBranch, "develop")
	viper.SetDefault(SortRepos, true)
	viper.SetDefault(UseSync, false)
	viper.SetDefault(CatalogCacheFile, ".catalog")
	viper.SetDefault(CatalogCacheTTL, "24h")

	viper.SetDefault(ChannelBuffer, 100)

	// default reviewers in the form `repo: [reviewers...]`
	viper.SetDefault(DefaultReviewers, map[string][]string{})

	// aliases in the form `alias: [repos...]`
	viper.SetDefault(RepoAliases, map[string][]string{})

	if gopath, err := exec.Command("go", "env", "GOPATH").Output(); err == nil {
		viper.SetDefault(EnvGopath, strings.TrimSpace(string(gopath)))
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		viper.SetConfigName("batch-tool")

		// Search in the working directory
		viper.AddConfigPath(".")

		// Search in system configuration (Linux/Darwin only)
		viper.AddConfigPath("/usr/local/etc/")

		// Search in the executable's directory
		if ex, err := os.Executable(); err == nil {
			viper.AddConfigPath(filepath.Dir(ex))
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %v\n\n", viper.ConfigFileUsed())
	}
}
