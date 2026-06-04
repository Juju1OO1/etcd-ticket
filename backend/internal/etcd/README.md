# internal/etcd

etcd 叢集管理與 client 封裝。

---

## etcd-cluster/

三節點 etcd 叢集，使用 docker-compose 啟動。

### 啟動指令

```bash
cd backend/internal/etcd/etcd-cluster
docker compose up -d
```

啟動成功應看到：

```
[+] Running 3/3
 ✔ Container etcd3  Started
 ✔ Container etcd2  Started
 ✔ Container etcd1  Started
```

### 關閉指令

```bash
docker compose down -v
```

### 問題排除

**權限問題（macOS / Linux）**

```
error while creating mount source path '...etcd3': permission denied
```

解法：

```bash
chmod -R 777 etcd1 etcd2 etcd3
```

---

## client.go

etcd client 的 singleton 封裝，連線到 `localhost:2379, 2381, 2382`。

### New()

取得 etcd client 實例，若未初始化則自動建立連線（lazy singleton）。

### Close()

程式結束時釋放連線，於 main.go 的 defer 呼叫。

提供基本 CRUD：
- `Get(ctx, key)` — 讀取單一 key
- `Put(ctx, key, value)` — 寫入 key-value
- `Delete(ctx, key)` — 刪除 key
