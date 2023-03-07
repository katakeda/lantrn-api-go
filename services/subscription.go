package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katakeda/lantrn-api-go/repositories"
)

func (s *Service) GetSubscriptions(c *gin.Context) {
	s.getSubscriptions(c)
}

func (s *Service) CreateSubscription(c *gin.Context) {
	s.createSubscription(c)
}

func (s *Service) getSubscriptions(c *gin.Context) (err error) {
	defer func() {
		if err != nil {
			log.Println("Failed to get subscriptions |", err)
			c.JSON(http.StatusInternalServerError, "Something went wrong while getting subscriptions")
		}
	}()

	params := c.Request.URL.Query()
	response, err := s.repo.GetSubscriptions(c, repositories.GetSubscriptionsFilter{
		FacilityIds: params.Get("facility_ids"),
		Status:      params.Get("status"),
		Page:        params.Get("page"),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch subscriptions | %w", err)
	}

	if len(response.Data) <= 0 {
		log.Println("No subscriptions found")
		c.JSON(http.StatusNotFound, "No subscriptions found")
		return
	}

	c.JSON(http.StatusOK, response)

	return nil
}

func (s *Service) createSubscription(c *gin.Context) (err error) {
	defer func() {
		if err != nil {
			log.Println("Failed to create subscription |", err)
			c.JSON(http.StatusInternalServerError, "Something went wrong while creating subscription")
		}
	}()

	payload := repositories.CreateSubscriptionPayload{}
	if err := c.BindJSON(&payload); err != nil {
		return fmt.Errorf("failed to parse payload | %w", err)
	}

	ctx, _ := s.repo.BeginTxn(c)
	subscription, err := s.repo.CreateSubscription(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to create subscription | %w", err)
	}

	c.JSON(http.StatusOK, subscription)

	return s.repo.CommitTxn(ctx)
}
