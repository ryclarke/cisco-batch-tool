package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"github.com/ryclarke/cisco-batch-tool/config"
)

var Catalog = make(map[string]Repository)
var Labels = make(map[string][]string)

type Repository struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Public      bool     `json:"public"`
	Project     string   `json:"project_name"`
	Labels      []string `json:"labels,omitempty"`
}

func InitRepositoryCatalog() error {
	if len(Catalog) > 0 {
		return nil
	}

	if err := loadCatalogCache(); err != nil {
		fmt.Print(err.Error())
	} else {
		return nil
	}

	return fetchRepositoryData()
}

type repositoryCache struct {
	UpdatedAt    time.Time             `json:"updated_at"`
	Repositories map[string]Repository `json:"repositories"`
}

type repositoryListResp struct {
	Values []Repository `json:"values"`
}

type labelListResp struct {
	Values []struct {
		Name string `json:"name"`
	} `json:"values"`
}

func loadCatalogCache() error {
	file, err := os.Open(catalogCachePath())
	if err != nil {
		return fmt.Errorf("local cache of repository catalog is missing or invalid - fetching remote info")
	}

	defer file.Close()

	var cached repositoryCache
	if err := json.NewDecoder(file).Decode(&cached); err != nil {
		return err
	}

	if time.Since(cached.UpdatedAt) > viper.GetDuration(config.CatalogCacheTTL) {
		return fmt.Errorf("local cache of repository catalog is too old - fetching remote info")
	}

	Catalog = cached.Repositories

	for _, repo := range Catalog {
		for _, label := range repo.Labels {
			Labels[label] = append(Labels[label], repo.Name)
		}
	}

	return nil
}

func saveCatalogCache() error {
	cache := repositoryCache{
		UpdatedAt:    time.Now().UTC(),
		Repositories: Catalog,
	}

	data, err := json.Marshal(&cache)
	if err != nil {
		return err
	}

	return os.WriteFile(catalogCachePath(), data, 0644)
}

func fetchRepositoryData() error {
	project := viper.GetString(config.GitProject)

	output, err := apiGET(fmt.Sprintf("https://%s/rest/api/1.0/projects/%s/repos?limit=1000", viper.GetString(config.GitHost), project))
	if err != nil {
		return err
	}

	var resp repositoryListResp
	if err := json.Unmarshal(output, &resp); err != nil {
		return err
	}

	for _, repo := range resp.Values {
		repo.Project = project

		labels, err := getLabels(project, repo.Name)
		if err != nil {
			return err
		}

		repo.Labels = labels

		Catalog[repo.Name] = repo

		for _, label := range repo.Labels {
			Labels[label] = append(Labels[label], repo.Name)
		}
	}

	return saveCatalogCache()
}

func catalogCachePath() string {
	return filepath.Join(viper.GetString(config.EnvGopath), "src",
		viper.GetString(config.GitHost),
		viper.GetString(config.GitProject),
		viper.GetString(config.CatalogCacheFile),
	)
}

func getLabels(project, repo string) ([]string, error) {
	output, err := apiGET(fmt.Sprintf("https://%s/rest/api/1.0/projects/%s/repos/%s/labels?limit=100", viper.GetString(config.GitHost), project, repo))
	if err != nil {
		return nil, err
	}

	var resp labelListResp
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, err
	}

	// Flatten the API response to extract the list of labels
	var labels = make([]string, len(resp.Values))
	for i, val := range resp.Values {
		labels[i] = val.Name
	}

	return labels, nil
}

func apiGET(path string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", viper.GetString(config.AuthToken)))

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	output, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 399 {
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, output)
	}

	return output, nil
}
