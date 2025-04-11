package bot

import (
	"context"
	"github.com/google/go-github/v55/github"
	"log"
)

const (
	gitHubOwner    = "kroexov"
	gitHubRepoName = "ludomania"
)

func getStarsCount() int {
	client := github.NewClient(nil)
	ctx := context.Background()
	repo, _, err := client.Repositories.Get(ctx, gitHubOwner, gitHubRepoName)
	if err != nil {
		log.Fatal(err)
	}

	stars := repo.GetStargazersCount()
	return stars
}
