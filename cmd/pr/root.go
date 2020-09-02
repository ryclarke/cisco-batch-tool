package pr

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ryclarke/cisco-batch-tool/config"
	"github.com/ryclarke/cisco-batch-tool/utils"
)

// Cmd configures the root pr command along with all subcommands and flags
func Cmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "pr [cmd] <repository> ...",
		Short: "Manage pull requests using the BitBucket v1 API",
		Args:  cobra.MinimumNArgs(1),
	}

	rootCmd.PersistentFlags().StringSliceP("reviewer", "r", nil, "pull request reviewer (cecid)")
	viper.BindPFlag(config.Reviewers, rootCmd.PersistentFlags().Lookup("reviewer"))

	defaultCmd := addNewCmd()

	rootCmd.AddCommand(
		defaultCmd,
		addEditCmd(),
		addMergeCmd(),
	)
	rootCmd.Run = defaultCmd.Run

	return rootCmd
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
