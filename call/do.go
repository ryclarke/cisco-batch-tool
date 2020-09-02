package call

import (
	"fmt"
	"sort"

	"github.com/spf13/viper"

	"github.com/ryclarke/cisco-batch-tool/config"
)

// Do executes the provided Wrapper on each repository, operating
// asynchronously by default. Repository aliases are also expanded
// here to allow for configurable repository grouping.
func Do(repos []string, fwrap Wrapper) {
	repos = processArguments(repos)

	// initialize channel set
	ch := make([]chan string, len(repos))
	for i := range repos {
		ch[i] = make(chan string, viper.GetInt(config.ChannelBuffer))
	}

	if viper.GetBool(config.UseSync) {
		// execute workers and print output synchronously
		for i, repo := range repos {
			go fwrap(repo, ch[i])

			for msg := range ch[i] {
				fmt.Println(msg)
			}
		}

		return
	}

	// start asynchronous workers
	for i, repo := range repos {
		go fwrap(repo, ch[i])
	}

	// batch and print ordered output
	for i := range repos {
		for msg := range ch[i] {
			fmt.Println(msg)
		}
	}
}

// DoAsync always operates asynchronously regardless of configuration
func DoAsync(repos []string, fwrap Wrapper) {
	viper.Set(config.UseSync, false)
	Do(repos, fwrap)
}

// DoSync always operates synchronously regardless of configuration
func DoSync(repos []string, fwrap Wrapper) {
	viper.Set(config.UseSync, true)
	Do(repos, fwrap)
}

func processArguments(args []string) []string {
	aliases := viper.GetStringMapStringSlice(config.RepoAliases)
	repos := make([]string, 0, len(args))

	// Expand any defined repository group aliases
	for _, repo := range args {
		if list := aliases[repo]; len(list) > 0 {
			repos = append(repos, list...)
		} else {
			repos = append(repos, repo)
		}
	}

	// Remove duplicate repository entries
	repoMap := make(map[string]struct{}, len(repos))

	for _, repo := range repos {
		repoMap[repo] = struct{}{}
	}

	repos = make([]string, len(repoMap))

	var i int

	for repo := range repoMap {
		repos[i] = repo
		i++
	}

	// Sort the repositories alphabetically
	if viper.GetBool(config.SortRepos) {
		sort.Strings(repos)
	}

	return repos
}
