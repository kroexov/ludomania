package bot

import (
	"context"
	"github.com/google/go-github/v55/github"
	"log"
)

func getStarsCount() int {
	client := github.NewClient(nil)
	ctx := context.Background()
	owner := "kroexov"
	repoName := "ludomania"

	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		log.Fatal(err)
	}

	stars := repo.GetStargazersCount()
	return stars
}
