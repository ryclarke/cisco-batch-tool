package pr

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ryclarke/cisco-batch-tool/call"
	"github.com/ryclarke/cisco-batch-tool/config"
	"github.com/ryclarke/cisco-batch-tool/utils"
)

// addMergeCmd initializes the pr merge command
func addMergeCmd() *cobra.Command {
	mergeCmd := &cobra.Command{
		Use:   "merge <repository> ...",
		Short: "Merge accepted pull requests",
		Args:  cobra.MinimumNArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			call.Do(args, call.Wrap(utils.ValidateBranch, mergePR))
		},
	}

	return mergeCmd
}

func mergePR(name string, ch chan<- string) error {
	pr, err := getPR(name)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/merge?version=%d", utils.ApiPathID(name, pr.ID()), pr.Version()), nil)
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

	ch <- fmt.Sprintf("Merged pull request (#%d) %s\n", pr.ID(), pr["title"].(string))

	return nil
}
