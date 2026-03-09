package service

import (
	"fmt"
	"log"

	"chaladshare_backend/internal/connect"
	recmodels "chaladshare_backend/internal/recommend/models"
	recrepo "chaladshare_backend/internal/recommend/repository"
)

const (
	minLikesForRecommend = 5
	seedLimit            = 5
	candidateLimit       = 800
	topK                 = 10
	boostSameCluster     = 0.05
	maxPerCluster        = 4
)

type RecommendService interface {
	RecomputeFromLikes(userID int) error
	OnLikeHook(userID int)
}

type svc struct {
	repo     recrepo.RecommendRepo
	aiClient *connect.Client
}

func NewRecommendService(repo recrepo.RecommendRepo, aiClient *connect.Client) RecommendService {
	return &svc{repo: repo, aiClient: aiClient}
}

func (s *svc) OnLikeHook(userID int) {
	go func() {
		if err := s.RecomputeFromLikes(userID); err != nil {
			log.Printf("[RECOMMEND] recompute error user=%d: %v", userID, err)
		}
	}()
}

func (s *svc) RecomputeFromLikes(userID int) error {
	if s.aiClient == nil {
		return fmt.Errorf("ai client is nil")
	}
	if userID <= 0 {
		return fmt.Errorf("invalid userID")
	}

	likeCount, err := s.repo.CountUserLikes(userID)
	if err != nil {
		return err
	}

	// ไม่มีไลก์ หรือยังไม่ถึงเกณฑ์ -> ล้าง recommendation เก่า
	if likeCount < minLikesForRecommend {
		return s.repo.ClearUserRecommendations(userID)
	}

	seeds, pairs, err := s.repo.ListSeedsFromLikes(userID, seedLimit)
	if err != nil {
		return err
	}
	if len(seeds) == 0 || len(pairs) == 0 {
		return s.repo.ClearUserRecommendations(userID)
	}

	cands, err := s.repo.ListCandidatesBySeedPairs(userID, pairs, candidateLimit)
	if err != nil {
		return err
	}
	if len(cands) == 0 {
		return s.repo.ClearUserRecommendations(userID)
	}

	req := recmodels.ColabRecommendFromLikedReq{
		Seeds:            seeds,
		Candidates:       cands,
		TopK:             topK,
		BoostSameCluster: boostSameCluster,
		MaxPerCluster:    maxPerCluster,
	}

	resp, err := s.aiClient.RecommendFromLiked(req)
	if err != nil {
		return err
	}

	// AI คืนว่าง -> ล้างของเก่า
	if resp == nil || len(resp.Recommendations) == 0 {
		return s.repo.ClearUserRecommendations(userID)
	}

	return s.repo.ReplaceUserRecommendations(userID, resp.Recommendations)
}
