package services

import "github.com/whotterre/push_microservice/internal/repository"

type PushService interface {

}

type pushService struct {
	pushRepo repository.PushRepository
}

func NewPushService(pushRepo repository.PushRepository) PushService {
	return &pushService{
		pushRepo: pushRepo,
	}
} 
