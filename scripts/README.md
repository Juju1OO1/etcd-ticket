# etcd Cluster 故障轉移 (Raft) Demo

## 用途

驗證 etcd 3 節點 cluster 在 **Leader 節點掛掉時**，能透過 **Raft 演算法自動重選 Leader**，系統仍可繼續對外服務（搶票邏輯仍然正確、不超賣）。

這對應**成員 8 的職責**：etcd Cluster 架設與故障轉移 (Raft)。

---

## 前置：cluster 已跑起來

```powershell
docker-compose -f backend\internal\etcd\etcd-cluster\docker-compose.yaml up -d
```

確認 3 個 container 都 Up：

```powershell
docker ps
```

---

## Demo 完整流程（建議照這個順序）

### Step 1: 確認 cluster 健康，找出當前 Leader

```powershell
.\scripts\show-cluster.ps1
```

預期看到 3 個節點都 Running，其中一個 `IS LEADER = true`。

---

### Step 2: 跑搶票測試（基準線）

```powershell
cd backend
go run .\test\etcd\etcd.go
```

預期 4 個測試全部 PASS — 證明**正常情況下系統工作正常**。

---

### Step 3: Kill 當前 Leader ⭐ 最關鍵

```powershell
cd ..
.\scripts\kill-leader.ps1
```

腳本會自動：
1. 找出當前 Leader
2. `docker kill` 它
3. 等 4 秒讓 Raft 重新選舉
4. 顯示新的 cluster 狀態（會看到原 Leader 變成 Stopped，新 Leader 由剩餘 2 節點之一接任）

---

### Step 4: 再跑一次搶票測試 

```powershell
cd backend
go run .\test\etcd\etcd.go
```

**預期：4 個測試仍然全部 PASS** — 證明 cluster 在 Leader 掛掉後仍可正常服務。

這就是 Raft 的核心價值：**容錯 + 強一致**。

---

### Step 5: 恢復 cluster

```powershell
cd ..
.\scripts\restart-cluster.ps1
```

把被 kill 的節點重新啟動，回復到完整 3 節點。

---

## Demo 時要講解的觀念

### Q1: 為什麼 Leader 死了 cluster 還能用？

**Raft 的「過半同意」原則**：
- 3 節點 cluster 的 quorum = `ceil(3/2) + 1` = 2
- 只要 ≥ 2 個節點存活，cluster 仍可接受寫入
- 1 個 Leader 死掉 → 剩 2 個（仍 ≥ quorum）→ 重新選出 Leader

### Q2: 重新選舉怎麼運作？

1. **Heartbeat 超時偵測**：剩餘節點偵測到 Leader 失聯（預設 ~1 秒）
2. **發起選舉**：其中一個轉成 Candidate，把 Term +1，呼叫 RequestVote RPC
3. **投票**：另一個節點同意（自己投自己 + 收到 1 票 = 2/2 過半）
4. **成為 Leader**：發送 AppendEntries 通知所有節點
5. **客戶端切換**：Go client 端有 3 個 endpoint，會自動 retry 新 Leader

整個過程大約 1~3 秒。

### Q3: 那 3 個都死了會怎樣？

**1 死**：cluster 還能用（quorum 2/2 OK）
**2 死**：cluster 失去 quorum，**不能接受寫入**（但仍能讀已寫入的資料）
**3 死**：cluster 完全離線

### Q4: 為什麼搶票邏輯仍然正確？

因為 `Lock_And_Hold` 和 `Checkout` 的所有 etcd 操作都會：
- 自動 retry 到新 Leader（client 內建）
- 受 ModRevision CAS 保護 → 並發下不超賣
- holder / paying 的 Lease 仍由新 Leader 維護

**短暫的選舉空窗期間（~1 秒）會有少數 timeout**，但程式碼裡的 retry 邏輯會把它們補回來。

---

## 如果某一步不對

- **Step 1 看不到 Leader**：cluster 可能還在啟動，等 5 秒再跑
- **Step 3 找不到 Leader**：可能剛好在選舉中，等幾秒再跑
- **Step 4 測試 FAIL**：檢查 cluster 是否還有 ≥ 2 節點存活
- **container Exited 但 docker ps 看不到**：用 `docker ps -a` 查 stopped 的

---

## 報告金句

> 「我們用 3 節點 etcd cluster + Raft 演算法實作高可用搶票系統。Demo 中 kill 掉當前 Leader，剩餘 2 節點透過 RequestVote 在約 1~3 秒內重新選出新 Leader，整個搶票邏輯持續運作且不超賣。這展示了 Raft 在 CAP 中選擇 **CP（強一致 + 分區容錯）** 的設計：容忍少數節點故障，保證任何時刻看到的庫存值都是全局一致的。」
