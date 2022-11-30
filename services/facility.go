package services

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katakeda/lantrn-api-go/repositories"
)

func (s *Service) GetFacilities(c *gin.Context) {
	s.getFacilities(c)
}

func (s *Service) GetFacility(c *gin.Context) {
	s.getFacility(c)
}

func (s *Service) getFacilities(c *gin.Context) (err error) {
	defer func() {
		if err != nil {
			log.Println("Failed to get facilities |", err)
			c.JSON(http.StatusInternalServerError, "Something went wrong while getting facilities")
		}
	}()

	params := c.Request.URL.Query()
	facilities, err := s.repo.GetFacilities(c, repositories.GetFacilitiesFilter{
		Lat:  params.Get("lat"),
		Lng:  params.Get("lng"),
		Sort: params.Get("sort"),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch facilities | %w", err)
	}

	if len(facilities) <= 0 {
		log.Println("No facilities found")
		c.JSON(http.StatusNotFound, "No facilities found")
		return
	}

	c.JSON(http.StatusOK, facilities)

	return nil
}

func (s *Service) getFacility(c *gin.Context) (err error) {
	defer func() {
		if err != nil {
			log.Println("Failed to get facility |", err)
			c.JSON(http.StatusInternalServerError, "Something went wrong while getting facility")
		}
	}()

	id := c.Param("id")
	facility, err := s.repo.GetFacility(c, id)
	if err != nil {
		return fmt.Errorf("failed to get facility | %w", err)
	}

	c.JSON(http.StatusOK, facility)

	return nil
}
