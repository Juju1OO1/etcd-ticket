package service

// 引入etcd client進行操作
// func init_etcd:初始化etcd的售票內容，票的狀態
// 限制每個區域購票人數的邏輯、限制檢視limit的 mutex
// 購票與修改etcd狀態transcation

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"etcd-ticket/internal/etcd"

	clientv3 "go.etcd.io/etcd/client/v3"

	"go.etcd.io/etcd/client/v3/concurrency"
)

/*
// ── etcd key 結構 ────────────────────────────────────────────────
//
// area{n}/holder/{userID}   → value: 搶票時間戳或 session token
//   - 代表「正在結帳」的佔位，同時存在的數量上限 = current_limit
//   - 【重要】建立時必須綁定 etcd Lease（TTL = 結帳時限），
//     超時後 etcd 自動刪除該 key，不需要手動清理逾時者
//
// area{n}/limit/current_limit → value: 剩餘可賣票數（初始 = limit）
//   - 只在用戶「成功結帳」後才 -1，holder 超時不影響此值
//   - 修改時需用 Txn + ModRevision CAS，防止並發寫衝突
//
// area{n}/limit/limit → value: 該區總票數（唯讀，不變）
//
// area{n}/lock → 競爭用的 mutex key
//   - 建議用 etcd concurrency.Mutex（內建 Lease），
//     確保取得鎖的程序崩潰後鎖會自動釋放，不會死鎖
//   - 保護「讀 holder 數量 → 建立 holder key」這段 check-then-act
//
// area{n}/status → "on" | "off"
//   - 初始為 "on"；current_limit 降至 0 時設為 "off"
//   - 前端可直接讀此 key 判斷是否售完，不用每次算 holder 數
//
// area{n}/sold/user1,user3(購買成功人的資訊)
//   - 初始為空
//   - 在購買完成後寫入。
//
// ── 搶票流程 ─────────────────────────────────────────────────────
//
// 1. 用戶請求 → 競爭 area{n}/lock（阻塞等待或直接回傳「請稍後」）
// 2. 取得鎖後 → GetPrefix(area{n}/holder/) 取得目前 holder 數量
// 3a. holder 數 < current_limit
//       → Put area{n}/holder/{userID}（帶 Lease TTL）
//       → 釋放鎖
//       → 回傳「搶票成功，進入結帳」
// 3b. holder 數 >= current_limit
//       → 釋放鎖
//       → 回傳「目前無票或候位中」
// 4a. 結帳成功（時限內完成付款）
//       → Txn: Delete area{n}/holder/{userID}
//              + Put current_limit = current_limit - 1
//       → 若 current_limit 降至 0，Put area{n}/status = "off"
// 4b. 結帳超時
//       → Lease 到期，etcd 自動刪除 area{n}/holder/{userID}
//       → 名額空出，其他競爭者可搶
//
// ── 注意事項 ─────────────────────────────────────────────────────
//
// - 步驟 2~3 必須在持鎖期間完成，釋放鎖後才讓下一位進入，
//   避免多人同時讀到相同 holder 數而超賣（TOCTOU race）
// - lock 的 Lease TTL 要短於 opTimeout，確保服務異常時鎖能自動釋放
// - init_etcd 應在服務啟動時執行一次，之後 current_limit/status 由
//   交易流程維護，不應重複初始化
*/

// 1.初始化後臺的程式碼，需要建立區域的設定，並初始化ETCD
// 2.
// 區域設定
type AreaConfig struct {
	Name  string // 區域名稱
	Limit int    // 該區票數
}

// InitEtcd 在服務啟動時呼叫一次，寫入各區域的初始狀態。
// 若 etcd 中已有舊資料，會直接覆蓋（重置系統狀態）。
func InitEtcd(ctx context.Context, areas []AreaConfig) error {
	client := etcd.New()

	for _, area := range areas {
		limitStr := strconv.Itoa(area.Limit)

		//初始化區域的前綴
		if err := client.Put(ctx, area.Name, "area 成功添加"); err != nil {
			return fmt.Errorf("init %s/: %w", area.Name, err)
		}

		// area{n}/limit/limit — 總票數（唯讀基準值）
		if err := client.Put(ctx, area.Name+"/limit/limit", limitStr); err != nil {
			return fmt.Errorf("init %s/limit/limit: %w", area.Name, err)
		}

		// area{n}/limit/current_limit — 剩餘可賣票數，初始 = limit
		if err := client.Put(ctx, area.Name+"/limit/current_limit", limitStr); err != nil {
			return fmt.Errorf("init %s/limit/current_limit: %w", area.Name, err)
		}

		// area{n}/status — 售票狀態，初始為開放
		if err := client.Put(ctx, area.Name+"/status", "on"); err != nil {
			return fmt.Errorf("init %s/status: %w", area.Name, err)
		}
		fmt.Printf("[INIT] %s ready (limit=%d, current_limit=%d, status=on)\n", area.Name, area.Limit, area.Limit)
	}
	fmt.Println("[INIT] all areas initialized")
	return nil
}

type TicketData struct {
	UserName string
	PhoneNum string
	Area     int
}

const checkoutTTL = 300 // 結帳時限，單位：秒
const payingTTL = 30    // paying 標記的存活時間，覆蓋 Checkout 重試所需的最大時間

// ctx代表競爭鎖的人，沒有重刷就繼續等鎖，有重刷就不在對列等待。
func Lock_And_Hold(ctx context.Context, td TicketData) (bool, error) {
	area := "area" + strconv.Itoa(td.Area)
	fmt.Printf("[RESERVE] %s trying to reserve %s\n", td.UserName, area)

	client := etcd.New()
	session, err := concurrency.NewSession(client.Cli, concurrency.WithTTL(10))
	if err != nil {
		return false, fmt.Errorf("建立 session 失敗: %w", err)
	}
	defer session.Close()

	mutex_key := area + "/lock"
	mutex := concurrency.NewMutex(session, mutex_key)

	if err := mutex.Lock(ctx); err != nil {
		// 拿鎖失敗（ctx取消、etcd 掛了）
		return false, err
	}
	fmt.Printf("[RESERVE] %s acquired lock %s\n", td.UserName, mutex_key)
	// 以下為鎖的持有
	ticket_prefix := area + "/holder"
	paying_prefix := area + "/paying"
	current_limit_key := area + "/limit/current_limit"

	holderCount, err := client.CountPrefix(ctx, ticket_prefix)
	if err != nil {
		//錯誤處理
		mutex.Unlock(ctx)
		return false, err
	}
	payingCount, err := client.CountPrefix(ctx, paying_prefix)
	if err != nil {
		mutex.Unlock(ctx)
		return false, err
	}
	// 當下holder(正在購買，付款畫面)
	// paying代表當下正在修改etcd的資料庫，已經有付款
	count := holderCount + payingCount
	cur_limit, exist, err := client.Get(ctx, current_limit_key)
	if err != nil {
		//錯誤處裡
		mutex.Unlock(ctx)
		return false, err
	}
	if !exist {
		//沒有該區域
		mutex.Unlock(ctx)
		return false, fmt.Errorf("%s 不存在", area)
	}
	i, err := strconv.ParseInt(cur_limit, 10, 64)
	if err != nil {
		//錯誤處裡
		mutex.Unlock(ctx)
		return false, err
	}
	fmt.Printf("[RESERVE] %s state: holder=%d paying=%d current_limit=%d\n", area, holderCount, payingCount, i)

	got := false
	if count < i {
		// 如果當下有空位，建立holder與該holder的時間上限，
		// 通知前端跳轉照結帳頁面，TTL也要傳到前端顯示剩餘時間
		lease, err := client.Cli.Grant(ctx, checkoutTTL)
		if err != nil {
			mutex.Unlock(ctx)
			return false, err
		}
		// holder key是用來辨識目前誰進入了結帳環節。
		holderKey := ticket_prefix + "/" + td.UserName
		_, err = client.Cli.Put(ctx, holderKey, strconv.FormatInt(time.Now().Unix(), 10), clientv3.WithLease(lease.ID))
		if err != nil {
			mutex.Unlock(ctx)
			return false, err
		}
		fmt.Printf("[RESERVE] ✓ %s reserved, created holder/%s (Lease %ds)\n", td.UserName, td.UserName, checkoutTTL)
		got = true
	} else {
		fmt.Printf("[RESERVE] ✗ %s no available slot (occupied=%d limit=%d)\n", td.UserName, count, i)
	}
	// 當下沒有空位，釋放這個檢查鎖。

	// 以上為鎖的持有
	mutex.Unlock(ctx)
	fmt.Printf("[RESERVE] %s released lock\n", td.UserName)
	return got, nil
}

// Checkout 在用戶付款完成後呼叫，原子性地刪除 holder 並扣減 current_limit。
// 若 holder 已過期（超時）回傳錯誤；若 current_limit 被並發修改則自動重試。
func Checkout(ctx context.Context, td TicketData) error {
	area := "area" + strconv.Itoa(td.Area)
	fmt.Printf("[CHECKOUT] %s starting checkout %s\n", td.UserName, area)

	client := etcd.New()

	holderKey := area + "/holder/" + td.UserName
	currentLimitKey := area + "/limit/current_limit"
	statusKey := area + "/status"

	payingKey := area + "/paying/" + td.UserName
	soldKey := area + "/sold/" + td.UserName

	// 原子操作：驗證 holder 存在的同時建立 paying 標記
	// 消除「驗證到建立」之間的空窗，防止 holder 在此期間剛好過期
	payingLease, err := client.Cli.Grant(ctx, payingTTL)
	if err != nil {
		return fmt.Errorf("建立 paying lease 失敗: %w", err)
	}
	verifyResp, err := client.Cli.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(holderKey), ">", 0)).
		Then(clientv3.OpPut(payingKey, "1", clientv3.WithLease(payingLease.ID))).
		Commit()
	if err != nil {
		return fmt.Errorf("建立 paying 狀態失敗: %w", err)
	}
	if !verifyResp.Succeeded {
		// holder 不存在，代表結帳時限已過
		fmt.Printf("[CHECKOUT] ✗ %s holder expired, checkout window passed\n", td.UserName)
		return fmt.Errorf("結帳時限已過，請重新搶票")
	}
	fmt.Printf("[CHECKOUT] ✓ %s holder verified, created paying/%s (Lease %ds)\n", td.UserName, td.UserName, payingTTL)

	// 迴圈重試修改limit。
	retries := 0
	for {
		// 讀取目前 current_limit，用來做 CAS 條件
		curLimitStr, exist, err := client.Get(ctx, currentLimitKey)
		if err != nil {
			return fmt.Errorf("讀取 current_limit 失敗: %w", err)
		}
		if !exist {
			return fmt.Errorf("area %s 不存在", area)
		}

		curLimit, err := strconv.ParseInt(curLimitStr, 10, 64)
		if err != nil {
			return fmt.Errorf("解析 current_limit 失敗: %w", err)
		}

		newLimitStr := strconv.FormatInt(curLimit-1, 10)
		fmt.Printf("[CHECKOUT] CAS attempt: %d -> %d\n", curLimit, curLimit-1)

		ops := []clientv3.Op{
			clientv3.OpDelete(holderKey), // holder 已過期時為 no-op，不影響 Txn 結果
			clientv3.OpDelete(payingKey), // 結帳完成，清除 paying 標記
			clientv3.OpPut(currentLimitKey, newLimitStr),
			clientv3.OpPut(soldKey, td.PhoneNum), // 與 current_limit 同一 Txn，Watcher 可用 ModRevision 比對判斷結帳成功
		}
		soldOut := curLimit-1 == 0
		if soldOut {
			// 票賣完，同一個 Txn 內順帶關閉售票
			ops = append(ops, clientv3.OpPut(statusKey, "off"))
		}

		// If: current_limit 未被並發修改
		// Then: 刪除 holder + 刪除 paying + 更新 current_limit（+ 可能關閉售票）
		txnResp, err := client.Cli.Txn(ctx).
			If(clientv3.Compare(clientv3.Value(currentLimitKey), "=", curLimitStr)).
			Then(ops...).
			Commit()
		if err != nil {
			return fmt.Errorf("結帳 txn 失敗: %w", err)
		}

		if txnResp.Succeeded {
			if soldOut {
				fmt.Printf("[CHECKOUT] ✓ CAS success, %s SOLD OUT, status=off (retries=%d)\n", area, retries)
			} else {
				fmt.Printf("[CHECKOUT] ✓ CAS success (retries=%d)\n", retries)
			}
			return nil
		}
		retries++
		fmt.Printf("[CHECKOUT] ↻ CAS conflict, retrying...\n")
		// current_limit 被並發修改，重新讀取後重試
	}
}
