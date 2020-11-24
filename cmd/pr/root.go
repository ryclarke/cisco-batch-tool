package pr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ryclarke/cisco-batch-tool/call"
	"github.com/ryclarke/cisco-batch-tool/config"
	"github.com/ryclarke/cisco-batch-tool/utils"
)

var (
	prTitle       string
	prDescription string
)

// Cmd configures the root pr command along with all subcommands and flags
func Cmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "pr [cmd] <repository> ...",
		Short: "Manage pull requests using the BitBucket v1 API",
		Args:  cobra.MinimumNArgs(1),
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return utils.ValidateRequiredConfig(config.AuthToken)
		},
		Run: func(_ *cobra.Command, args []string) {
			call.Do(args, call.Wrap(utils.ValidateBranch, getPRCmd))
		},
	}

	rootCmd.PersistentFlags().StringVarP(&prTitle, "title", "t", "", "pull request title")
	rootCmd.PersistentFlags().StringVarP(&prDescription, "description", "d", "", "pull request description")

	rootCmd.PersistentFlags().StringSliceP("reviewer", "r", nil, "pull request reviewer (cecid)")
	viper.BindPFlag(config.Reviewers, rootCmd.PersistentFlags().Lookup("reviewer"))

	rootCmd.AddCommand(
		addNewCmd(),
		addEditCmd(),
		addMergeCmd(),
	)

	return rootCmd
}

func getPRCmd(name string, ch chan<- string) error {
	pr, err := getPR(name)
	if err != nil {
		return err
	}

	ch <- fmt.Sprintf("(PR #%d) %s %v\n", pr.ID(), pr["title"].(string), pr.GetReviewers())
	if pr["description"].(string) != "" {
		ch <- fmt.Sprintln(pr["description"].(string))
	}

	return nil
}

func getPR(name string) (utils.PR, error) {
	branch, err := utils.LookupBranch(name)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?direction=outgoing&at=refs/heads/%s", utils.ApiPath(name), branch), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viper.GetString(config.AuthToken)))
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// results will be returned in an array property called "values"
	raw := struct {
		Values []utils.PR `json:"values"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if len(raw.Values) == 0 {
		return nil, fmt.Errorf("No pull requests found for %s", branch)
	}

	// return the first PR in the results (this will be the most recent)
	return raw.Values[0], nil
}
