package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/ryclarke/cisco-batch-tool/config"
)

// GenPR generates a PR payload for the Bitbucket v1 API
func GenPR(name, title string, reviewers []string) string {
	project := viper.GetString(config.GitProject)

	// generate list of reviewers
	revs := make([]string, 0, len(reviewers))
	for _, rev := range reviewers {
		if rev != "" {
			revs = append(revs, fmt.Sprintf(config.PrReviewerTmpl, rev))
		}
	}

	return fmt.Sprintf(config.PrTmpl, title,
		viper.GetString(config.Branch), fmt.Sprintf(config.PrRepoTmpl, name, project),
		viper.GetString(config.SourceBranch), fmt.Sprintf(config.PrRepoTmpl, name, project),
		strings.Join(revs, ","),
	)
}

// PR data fetched from the BitBucket v1 API
type PR map[string]interface{}

// ID of the PR
func (pr PR) ID() int {
	if id, ok := pr["id"]; ok {
		return int(id.(float64))
	}

	return 0
}

// Version of the PR
func (pr PR) Version() int {
	if version, ok := pr["version"]; ok {
		return int(version.(float64))
	}

	return 0
}

// AddReviewers appends the given list of reviewers to the PR
func (pr PR) AddReviewers(reviewers []string) {
	for _, rev := range reviewers {
		pr["reviewers"] = append(pr["reviewers"].([]interface{}), map[string]interface{}{
			"user": map[string]interface{}{"name": rev},
		})
	}
}

// GetReviewers returns the list of reviewers for the PR
func (pr PR) GetReviewers() []string {
	revs := pr["reviewers"].([]interface{})
	output := make([]string, len(revs))

	for i, rev := range revs {
		output[i] = rev.(map[string]interface{})["user"].(map[string]interface{})["name"].(string)
	}

	return output
}
