package api

import (
	"etcd-ticket/pkg/response"

	"github.com/gin-gonic/gin"
)

// === Handler stubs ===
//
// 以下三個 handler 由成員 4（Booking Service - API 邏輯）負責填 body。
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
	response.Fail(c, 501, "TODO by member4: Lock_And_Hold")
}

// CheckoutHandler — POST /api/tickets/checkout
// 對應 service.Checkout（第二階段結帳 CAS Txn）+ service.PublishOrder（寫 Redis Stream）
func CheckoutHandler(c *gin.Context) {
	response.Fail(c, 501, "TODO by member4: Checkout + PublishOrder")
}

// StatusHandler — GET /api/tickets/status
// 列出各區剩餘票數與售票狀態（optional）
func StatusHandler(c *gin.Context) {
	response.Fail(c, 501, "TODO by member4: list area status")
}
