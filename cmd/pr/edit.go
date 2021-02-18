package pr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ryclarke/cisco-batch-tool/call"
	"github.com/ryclarke/cisco-batch-tool/config"
	"github.com/ryclarke/cisco-batch-tool/utils"
)

var noAppendReviewers bool

// addEditCmd initializes the pr edit command
func addEditCmd() *cobra.Command {
	editCmd := &cobra.Command{
		Use:   "edit <repository> ...",
		Short: "Update existing pull requests",
		Args:  cobra.MinimumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			call.Do(args, call.Wrap(utils.ValidateBranch, editPR))
		},
	}

	editCmd.Flags().BoolVar(&noAppendReviewers, "no-append", false, "don't append to the reviewer list")

	return editCmd
}

func editPR(name string, ch chan<- string) error {
	pr, err := getPR(name)
	if err != nil {
		return err
	}

	if prTitle != "" {
		pr["title"] = prTitle
	}

	if prDescription != "" {
		pr["description"] = prDescription
	}

	// set or append reviewers to the PR and perform necessary modifications
	if noAppendReviewers {
		// if no-append is combined with reviewers, replace existing (otherwise ignore reviewers entirely)
		if len(viper.GetStringSlice(config.Reviewers)) > 0 {
			pr.SetReviewers(utils.LookupReviewers(name))
		}
	} else {
		pr.AddReviewers(utils.LookupReviewers(name))
	}

	delete(pr, "participants")
	delete(pr, "author")

	payload, err := json.Marshal(pr)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPut, utils.ApiPathID(name, pr.ID()), bytes.NewReader(payload))
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

	if resp.StatusCode > 399 {
		output, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("error %d: %s", resp.StatusCode, output)
	}

	ch <- fmt.Sprintf("Updated pull request (#%d) %s %v\n", pr.ID(), pr["title"].(string), pr.GetReviewers())

	return nil
}
