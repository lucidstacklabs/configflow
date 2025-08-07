package environment

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucidstacklabs/configflow/internal/pkg/actor"
	"github.com/lucidstacklabs/configflow/internal/pkg/auth"
)

type Handler struct {
	router        *gin.Engine
	authenticator *auth.Authenticator
	service       *Service
}

func NewHandler(router *gin.Engine, authenticator *auth.Authenticator, service *Service) *Handler {
	return &Handler{router: router, authenticator: authenticator, service: service}
}

func (h *Handler) Register() {

	h.router.POST("/admin/api/v1/environments", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, time.Second*5)
		defer cancel()

		aa, err := h.authenticator.ValidateAdminContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		var req CreateRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		environment, err := h.service.Create(ctx, &req, actor.TypeAdmin, aa.ID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		c.JSON(http.StatusCreated, environment)
	})

	h.router.GET("/admin/api/v1/environments", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, time.Second*5)
		defer cancel()

		_, err := h.authenticator.ValidateAdminContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		page := c.DefaultQuery("page", "0")
		size := c.DefaultQuery("size", "50")
		pageInt, err := strconv.ParseInt(page, 10, 64)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		sizeInt, err := strconv.ParseInt(size, 10, 64)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		environments, err := h.service.List(ctx, pageInt, sizeInt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, environments)
	})

	h.router.GET("/admin/api/v1/environments/:environmentID", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, time.Second*5)
		defer cancel()

		_, err := h.authenticator.ValidateAdminContext(c)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		environmentID := c.Param("environmentID")

		environment, err := h.service.Get(ctx, environmentID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, environment)
	})

	h.router.PUT("/admin/api/v1/environments/:environmentID", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, time.Second*5)
		defer cancel()

		_, err := h.authenticator.ValidateAdminContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		environmentID := c.Param("environmentID")

		var req UpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		environment, err := h.service.Update(ctx, environmentID, &req)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, environment)
	})

	h.router.DELETE("/admin/api/v1/environments/:environmentID", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, time.Second*5)
		defer cancel()

		_, err := h.authenticator.ValidateAdminContext(c)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		environmentID := c.Param("environmentID")

		environment, err := h.service.Delete(ctx, environmentID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, environment)
	})
}
