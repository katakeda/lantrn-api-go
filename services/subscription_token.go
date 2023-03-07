package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katakeda/lantrn-api-go/repositories"
)

func (s *Service) GetSubscriptionTokens(c *gin.Context) {
	s.getSubscriptionTokens(c)
}

func (s *Service) CreateSubscriptionToken(c *gin.Context) {
	s.createSubscriptionToken(c)
}

func (s *Service) getSubscriptionTokens(c *gin.Context) (err error) {
	defer func() {
		if err != nil {
			log.Println("Failed to get subscription tokens |", err)
			c.JSON(http.StatusInternalServerError, "Something went wrong while getting subscription tokens")
		}
	}()

	params := c.Request.URL.Query()
	response, err := s.repo.GetSubscriptionTokens(c, repositories.GetSubscriptionTokensFilter{
		Token: params.Get("token"),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch subscription tokens | %w", err)
	}

	if len(response.Data) <= 0 {
		log.Println("No subscription tokens found")
		c.JSON(http.StatusNotFound, "No subscription tokens found")
		return
	}

	c.JSON(http.StatusOK, response)

	return nil
}

func (s *Service) createSubscriptionToken(c *gin.Context) (err error) {
	defer func() {
		if err != nil {
			log.Println("Failed to create subscription token |", err)
			c.JSON(http.StatusInternalServerError, "Something went wrong while creating subscription token")
		}
	}()

	payload := repositories.CreateSubscriptionTokenPayload{}
	if err := c.BindJSON(&payload); err != nil {
		return fmt.Errorf("failed to parse payload | %w", err)
	}

	ctx, _ := s.repo.BeginTxn(c)
	subscriptionToken, err := s.repo.CreateSubscriptionToken(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to create subscription token | %w", err)
	}

	c.JSON(http.StatusOK, subscriptionToken)

	return s.repo.CommitTxn(ctx)
}
