package bot

import (
	"context"
	"github.com/google/go-github/v55/github"
)

type GithubService struct {
	client *github.Client
	owner  string
	repo   string
}

func NewGithubService(owner, repo string) *GithubService {
	return &GithubService{
		client: github.NewClient(nil),
		owner:  owner,
		repo:   repo,
	}
}

func (s *GithubService) GetStarsCount(ctx context.Context) (int, error) {
	r, _, err := s.client.Repositories.Get(ctx, s.owner, s.repo)
	if err != nil {
		return 0, err
	}
	return r.GetStargazersCount(), nil
}
