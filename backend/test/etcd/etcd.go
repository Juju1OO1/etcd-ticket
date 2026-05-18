package etcd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"etcd-ticket/internal/etcd"
	"etcd-ticket/internal/service"
)

const testArea = 1

func main() {
	defer etcd.Close()
	ctx := context.Background()

	fmt.Println("========== 測試開始 ==========")
	testInitEtcd(ctx)
	testConcurrentLockAndHold(ctx)
	testCheckout(ctx)
	testPaymentFailureAndSubstitution(ctx)
	fmt.Println("\n========== 所有測試完成 ==========")
}

// ─── 輔助函式 ────────────────────────────────────────────────

func section(name string) {
	fmt.Printf("\n--- %s ---\n", name)
}

func pass()              { fmt.Println("結果: PASS ✓") }
func fail(reason string) { fmt.Printf("結果: FAIL ✗ (%s)\n", reason) }

// 清除某區域所有非設定 key，確保每個測試從乾淨狀態開始
func resetAreaState(ctx context.Context, areaName string) {
	c := etcd.New()
	c.DeletePrefix(ctx, areaName+"/holder/")
	c.DeletePrefix(ctx, areaName+"/paying/")
	c.DeletePrefix(ctx, areaName+"/sold/")
}

// ─── 測試 1：InitEtcd ────────────────────────────────────────

func testInitEtcd(ctx context.Context) {
	section("測試 1：InitEtcd 初始化")

	const limit = 5
	areas := []service.AreaConfig{
		{Name: "area1", Limit: limit},
	}
	if err := service.InitEtcd(ctx, areas); err != nil {
		fail(fmt.Sprintf("InitEtcd 回傳錯誤: %v", err))
		return
	}

	c := etcd.New()
	limitVal, _, _ := c.Get(ctx, "area1/limit/limit")
	curLimit, _, _ := c.Get(ctx, "area1/limit/current_limit")
	status, _, _ := c.Get(ctx, "area1/status")

	expected := fmt.Sprint(limit)
	fmt.Printf("  limit         = %s (預期 %s)\n", limitVal, expected)
	fmt.Printf("  current_limit = %s (預期 %s)\n", curLimit, expected)
	fmt.Printf("  status        = %s (預期 on)\n", status)

	if limitVal == expected && curLimit == expected && status == "on" {
		pass()
	} else {
		fail("key 值與預期不符")
	}
}

// ─── 測試 2：並發搶票 ─────────────────────────────────────────

func testConcurrentLockAndHold(ctx context.Context) {
	section("測試 2：Lock_And_Hold — 10 人並發搶 3 張票")

	const (
		totalTickets = 3
		totalUsers   = 10
	)

	service.InitEtcd(ctx, []service.AreaConfig{{Name: "area1", Limit: totalTickets}})
	resetAreaState(ctx, "area1")

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		succeeded []string
		errCount  int
	)

	for i := 1; i <= totalUsers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			td := service.TicketData{
				UserName: fmt.Sprintf("user%02d", id),
				PhoneNum: fmt.Sprintf("09%08d", id),
				Area:     testArea,
			}
			userCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			got, err := service.Lock_And_Hold(userCtx, td)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				fmt.Printf("  [%s] 錯誤: %v\n", td.UserName, err)
				errCount++
				return
			}
			if got {
				succeeded = append(succeeded, td.UserName)
			}
		}(i)
	}
	wg.Wait()

	fmt.Printf("  搶票成功: %d 人 (預期 %d)\n", len(succeeded), totalTickets)
	fmt.Printf("  成功者: %v\n", succeeded)
	fmt.Printf("  錯誤數: %d (預期 0)\n", errCount)

	if len(succeeded) == totalTickets && errCount == 0 {
		pass()
	} else {
		fail(fmt.Sprintf("成功人數=%d，錯誤數=%d", len(succeeded), errCount))
	}
}

// ─── 測試 3：正常結帳 ─────────────────────────────────────────

func testCheckout(ctx context.Context) {
	section("測試 3：Checkout — 3 人全部結帳成功")

	const totalTickets = 3
	service.InitEtcd(ctx, []service.AreaConfig{{Name: "area1", Limit: totalTickets}})
	resetAreaState(ctx, "area1")

	users := []service.TicketData{
		{UserName: "alice", PhoneNum: "0911111111", Area: testArea},
		{UserName: "bob", PhoneNum: "0922222222", Area: testArea},
		{UserName: "carol", PhoneNum: "0933333333", Area: testArea},
	}

	// 依序搶票（sequential 即可，鎖會序列化）
	for _, td := range users {
		got, err := service.Lock_And_Hold(ctx, td)
		if err != nil || !got {
			fail(fmt.Sprintf("%s 搶票失敗 err=%v got=%v", td.UserName, err, got))
			return
		}
	}
	fmt.Printf("  搶票完成: %v\n", []string{"alice", "bob", "carol"})

	// 並發結帳
	var wg sync.WaitGroup
	for _, td := range users {
		wg.Add(1)
		go func(t service.TicketData) {
			defer wg.Done()
			checkoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := service.Checkout(checkoutCtx, t); err != nil {
				fmt.Printf("  [%s] 結帳失敗: %v\n", t.UserName, err)
			} else {
				fmt.Printf("  [%s] 結帳成功\n", t.UserName)
			}
		}(td)
	}
	wg.Wait()

	c := etcd.New()
	curLimit, _, _ := c.Get(ctx, "area1/limit/current_limit")
	status, _, _ := c.Get(ctx, "area1/status")
	sold, _ := c.GetPrefix(ctx, "area1/sold/")
	holders, _ := c.GetPrefix(ctx, "area1/holder/")

	fmt.Printf("  current_limit = %s (預期 0)\n", curLimit)
	fmt.Printf("  status        = %s (預期 off)\n", status)
	fmt.Printf("  sold 筆數     = %d (預期 %d)\n", len(sold), totalTickets)
	fmt.Printf("  殘留 holder   = %d (預期 0)\n", len(holders))

	if curLimit == "0" && status == "off" && len(sold) == totalTickets && len(holders) == 0 {
		pass()
	} else {
		fail("最終狀態與預期不符")
	}
}

// ─── 測試 4：付款失敗 + 遞補 ──────────────────────────────────

func testPaymentFailureAndSubstitution(ctx context.Context) {
	section("測試 4：付款失敗 + 遞補")
	fmt.Println("  情境: alice、bob 搶到 2 張票，bob 付款超時，carol 遞補")

	const totalTickets = 2
	service.InitEtcd(ctx, []service.AreaConfig{{Name: "area1", Limit: totalTickets}})
	resetAreaState(ctx, "area1")

	tdAlice := service.TicketData{UserName: "alice", PhoneNum: "0911111111", Area: testArea}
	tdBob := service.TicketData{UserName: "bob", PhoneNum: "0922222222", Area: testArea}
	tdCarol := service.TicketData{UserName: "carol", PhoneNum: "0933333333", Area: testArea}

	// alice 和 bob 搶到所有票
	for _, td := range []service.TicketData{tdAlice, tdBob} {
		got, err := service.Lock_And_Hold(ctx, td)
		if err != nil || !got {
			fail(fmt.Sprintf("%s 搶票失敗", td.UserName))
			return
		}
	}
	fmt.Println("  alice, bob 搶票成功（票已全部佔位）")

	// carol 搶不到（已滿）
	got, _ := service.Lock_And_Hold(ctx, tdCarol)
	if got {
		fail("carol 不應搶到票（票已佔滿）")
		return
	}
	fmt.Println("  carol 搶票失敗 ✓（票已佔滿，預期行為）")

	// 模擬 bob 付款超時：Lease 到期時 etcd 自動刪除 holder，這裡手動刪除模擬
	c := etcd.New()
	if err := c.Delete(ctx, "area1/holder/bob"); err != nil {
		fail(fmt.Sprintf("模擬 bob 超時失敗: %v", err))
		return
	}
	fmt.Println("  bob 付款超時，holder 已過期（模擬 Lease TTL）")

	// bob 嘗試結帳 → 應失敗
	err := service.Checkout(ctx, tdBob)
	if err == nil {
		fail("bob 結帳應失敗（holder 已過期）")
		return
	}
	fmt.Printf("  bob 結帳失敗 ✓（預期行為）: %v\n", err)

	// carol 遞補
	got, err = service.Lock_And_Hold(ctx, tdCarol)
	if err != nil || !got {
		fail(fmt.Sprintf("carol 遞補失敗 err=%v got=%v", err, got))
		return
	}
	fmt.Println("  carol 遞補成功 ✓")

	// alice 和 carol 結帳
	var wg sync.WaitGroup
	for _, td := range []service.TicketData{tdAlice, tdCarol} {
		wg.Add(1)
		go func(t service.TicketData) {
			defer wg.Done()
			checkoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := service.Checkout(checkoutCtx, t); err != nil {
				fmt.Printf("  [%s] 結帳失敗: %v\n", t.UserName, err)
			} else {
				fmt.Printf("  [%s] 結帳成功\n", t.UserName)
			}
		}(td)
	}
	wg.Wait()

	// 驗證最終狀態
	curLimit, _, _ := c.Get(ctx, "area1/limit/current_limit")
	status, _, _ := c.Get(ctx, "area1/status")
	sold, _ := c.GetPrefix(ctx, "area1/sold/")
	holders, _ := c.GetPrefix(ctx, "area1/holder/")
	_, bobSold := sold["area1/sold/bob"]

	fmt.Printf("  current_limit = %s (預期 0)\n", curLimit)
	fmt.Printf("  status        = %s (預期 off)\n", status)
	fmt.Printf("  sold 筆數     = %d (預期 %d)\n", len(sold), totalTickets)
	fmt.Printf("  bob 是否有 sold 紀錄 = %v (預期 false)\n", bobSold)
	fmt.Printf("  殘留 holder   = %d (預期 0)\n", len(holders))

	if curLimit == "0" && status == "off" && len(sold) == totalTickets && !bobSold && len(holders) == 0 {
		pass()
	} else {
		fail("最終狀態與預期不符")
	}
}
