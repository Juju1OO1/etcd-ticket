package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"etcd-ticket/internal/etcd"
	"etcd-ticket/internal/service"
	"etcd-ticket/pkg/response"

	"github.com/gin-gonic/gin"
)

type ticketRequest struct {
	UserName string `json:"user_name"`
	PhoneNum string `json:"phone_num"`
	Area     int    `json:"area"`
	AreaID   int    `json:"area_id"`
}

type ticketResponse struct {
	UserName string `json:"user_name"`
	PhoneNum string `json:"phone_num"`
	Area     int    `json:"area"`
}

type reserveResponse struct {
	Reserved bool           `json:"reserved"`
	Ticket   ticketResponse `json:"ticket"`
}

type checkoutResponse struct {
	CheckedOut     bool           `json:"checked_out"`
	OrderPublished bool           `json:"order_published"`
	Ticket         ticketResponse `json:"ticket"`
}

type areaStatusResponse struct {
	AreaID    int    `json:"area_id"`
	Available int64  `json:"available"`
	Status    string `json:"status"`
}

type statusResponse struct {
	Areas []areaStatusResponse `json:"areas"`
}

var (
	reserveTicketFn    = service.Lock_And_Hold
	checkoutTicketFn   = service.Checkout
	publishOrderFn     = service.PublishOrder
	availableTicketsFn = service.GetAvailableTickets
	loadAreaStatusFn   = defaultLoadAreaStatus
)

// === Handler stubs ===
//
// 以下三個 handler 由成員 4（Booking Service（API 邏輯））負責填 body。
// 簽名固定不動，方便平行開工。
//
// 共通流程：
//   1. 解析 JSON request body
//   2. 組 service.TicketData{UserName, PhoneNum, Area}
//   3. 呼叫 service.Lock_And_Hold / Checkout / PublishOrder
//   4. 使用 response.OK / response.Fail 回應

// ReserveHandler — POST /api/tickets/reserve
// 對應 service.Lock_And_Hold（第一階段搶票，建立 holder + Lease 300s）
func ReserveHandler(c *gin.Context) {
	td, err := bindTicketRequest(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	got, err := reserveTicketFn(c.Request.Context(), td)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	if !got {
		response.Fail(c, http.StatusConflict, "目前無票或候位中")
		return
	}

	response.OK(c, reserveResponse{
		Reserved: true,
		Ticket:   toTicketResponse(td),
	})
}

// CheckoutHandler — POST /api/tickets/checkout
// 對應 service.Checkout（第二階段結帳 CAS Txn）+ service.PublishOrder（寫 Redis Stream）
func CheckoutHandler(c *gin.Context) {
	td, err := bindTicketRequest(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := checkoutTicketFn(c.Request.Context(), td); err != nil {
		writeServiceError(c, err)
		return
	}

	if err := publishOrderFn(c.Request.Context(), td); err != nil {
		response.Fail(c, http.StatusBadGateway, fmt.Sprintf("建立訂單失敗: %v", err))
		return
	}

	response.OK(c, checkoutResponse{
		CheckedOut:     true,
		OrderPublished: true,
		Ticket:         toTicketResponse(td),
	})
}

// StatusHandler — GET /api/tickets/status
// 列出各區剩餘票數與售票狀態（optional）
func StatusHandler(c *gin.Context) {
	areaIDs, err := parseStatusAreaIDs(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	areas := make([]areaStatusResponse, 0, len(areaIDs))
	ctx := c.Request.Context()
	for _, areaID := range areaIDs {
		available, err := availableTicketsFn(ctx, areaID)
		if err != nil {
			writeServiceError(c, err)
			return
		}

		status, err := loadAreaStatusFn(ctx, areaID)
		if err != nil {
			writeServiceError(c, err)
			return
		}

		areas = append(areas, areaStatusResponse{
			AreaID:    areaID,
			Available: available,
			Status:    status,
		})
	}

	response.OK(c, statusResponse{Areas: areas})
}

func bindTicketRequest(c *gin.Context) (service.TicketData, error) {
	var req ticketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return service.TicketData{}, fmt.Errorf("invalid request body: %w", err)
	}

	area := req.Area
	if area == 0 {
		area = req.AreaID
	}

	td := service.TicketData{
		UserName: strings.TrimSpace(req.UserName),
		PhoneNum: strings.TrimSpace(req.PhoneNum),
		Area:     area,
	}
	if td.UserName == "" || td.PhoneNum == "" || td.Area <= 0 {
		return service.TicketData{}, fmt.Errorf("user_name, phone_num, area are required")
	}

	return td, nil
}

func toTicketResponse(td service.TicketData) ticketResponse {
	return ticketResponse{
		UserName: td.UserName,
		PhoneNum: td.PhoneNum,
		Area:     td.Area,
	}
}

func writeServiceError(c *gin.Context, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "結帳時限已過"):
		response.Fail(c, http.StatusGone, msg)
	case strings.Contains(msg, "不存在"):
		response.Fail(c, http.StatusNotFound, msg)
	default:
		response.Fail(c, http.StatusInternalServerError, msg)
	}
}

func parseStatusAreaIDs(c *gin.Context) ([]int, error) {
	if values := c.QueryArray("area"); len(values) > 0 {
		return parseAreaValues(values)
	}
	if values := c.QueryArray("area_id"); len(values) > 0 {
		return parseAreaValues(values)
	}

	if raw := strings.TrimSpace(c.Query("areas")); raw != "" {
		return parseAreaValues(strings.Split(raw, ","))
	}
	if raw := strings.TrimSpace(c.Query("area_ids")); raw != "" {
		return parseAreaValues(strings.Split(raw, ","))
	}

	return []int{1, 2}, nil
}

func parseAreaValues(values []string) ([]int, error) {
	areas := make([]int, 0, len(values))
	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			item := strings.TrimSpace(part)
			if item == "" {
				continue
			}

			areaID, err := strconv.Atoi(item)
			if err != nil || areaID <= 0 {
				return nil, fmt.Errorf("invalid area id: %s", item)
			}
			areas = append(areas, areaID)
		}
	}

	if len(areas) == 0 {
		return nil, fmt.Errorf("at least one area id is required")
	}

	return areas, nil
}

func defaultLoadAreaStatus(ctx context.Context, areaID int) (string, error) {
	area := "area" + strconv.Itoa(areaID)
	statusKey := area + "/status"

	status, exist, err := etcd.New().Get(ctx, statusKey)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", fmt.Errorf("%s 不存在", statusKey)
	}

	return status, nil
}
