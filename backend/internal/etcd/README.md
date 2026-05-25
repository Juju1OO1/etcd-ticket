## 啟動步驟

```
// 切到目錄
cd backend/internal/etcd/etcd-cluster

// 啟動
sudo docker-compose up -d
```

啟動成功：

```
[+] Running 3/3
 ✔ Container etcd3  Started                                                0.4s
 ✔ Container etcd2  Started                                                0.4s
 ✔ Container etcd1  Started                                                0.5s
```

```
// 結束
sudo docker-compose down -v
```

## 問題排除

1. 在起 etcd 要寫三個 etcd file 可能遇到權限問題

```
Error response from daemon: error while creating mount source path '/Users/lizhiyun/Desktop/Projects/etcd-ticket/backend/internal/etcd/etcd-cluster/etcd3': chown /Users/lizhiyun/Desktop/Projects/etcd-ticket/backend/internal/etcd/etcd-cluster/etcd3: permission denied
```

解法：修改權限

```
chmod -R 777 etcd1 etcd2 etcd3
```
