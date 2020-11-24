package pr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ryclarke/cisco-batch-tool/call"
	"github.com/ryclarke/cisco-batch-tool/config"
	"github.com/ryclarke/cisco-batch-tool/utils"
)

var allReviewers bool

// addNewCmd initializes the pr new command
func addNewCmd() *cobra.Command {
	newCmd := &cobra.Command{
		Use:   "new <repository> ...",
		Short: "Submit new pull requests",
		Args:  cobra.MinimumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			call.Do(args, call.Wrap(utils.ValidateBranch, newPR))
		},
	}

	newCmd.Flags().BoolVarP(&allReviewers, "all-reviewers", "a", false, "use all provided reviewers for a new PR")

	return newCmd
}

func newPR(name string, ch chan<- string) error {
	branch, err := utils.LookupBranch(name)
	if err != nil {
		return err
	}

	// default PR title is branch name
	if prTitle == "" {
		prTitle = branch
	}

	reviewers := utils.LookupReviewers(name)
	if len(reviewers) == 0 {
		// append placeholder to prevent NPE below
		reviewers = append(reviewers, "")
	}

	// remove all but the first reviewer by default
	if !allReviewers && len(reviewers) > 1 {
		reviewers = reviewers[:1]
	}

	payload := utils.GenPR(name, prTitle, prDescription, reviewers)

	request, err := http.NewRequest(http.MethodPost, utils.ApiPath(name), strings.NewReader(payload))
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viper.GetString(config.AuthToken)))
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	output, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 399 {
		return fmt.Errorf("error %d: %s", resp.StatusCode, output)
	}

	pr := struct {
		ID int `json:"id"`
	}{}
	if err := json.NewDecoder(strings.NewReader(string(output))).Decode(&pr); err != nil {
		return err
	}

	ch <- fmt.Sprintf("New pull request (#%d) %s %v\n", pr.ID, branch, reviewers)

	return nil
}
