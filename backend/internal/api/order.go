package api

import (
	"net/http"
	"strings"

	"etcd-ticket/internal/repository"
	"etcd-ticket/pkg/response"

	"github.com/gin-gonic/gin"
)

// OrdersHandler — GET /api/orders?user_name=xxx
func OrdersHandler(c *gin.Context) {
	userName := strings.TrimSpace(c.Query("user_name"))
	if userName == "" {
		response.Fail(c, http.StatusBadRequest, "user_name is required")
		return
	}

	orders, err := repository.GetOrdersByUser(c.Request.Context(), userName)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, orders)
}
