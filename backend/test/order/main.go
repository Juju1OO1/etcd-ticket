package main

import (
    "context"
    "etcd-ticket/internal/db"
    "etcd-ticket/internal/mq"
    "etcd-ticket/internal/service"
    "fmt"
    "time"
)

func main() {
    ctx := context.Background()

    // 初始化 DB 和 Redis
    if err := db.Init(ctx, "postgres://postgres:password@localhost:5433/etcd_ticket?sslmode=disable"); err != nil {
        panic(err)
    }
    defer db.Close()

    if err := mq.Init(ctx, "localhost:6379"); err != nil {
        panic(err)
    }
    defer mq.Close()

    // 啟動 order worker
    service.StartOrderWorker(ctx)

    // 模擬多筆訂單同時結帳成功
    users := []service.TicketData{
        {UserName: "alice", PhoneNum: "0911111111", Area: 1},
        {UserName: "bob", PhoneNum: "0922222222", Area: 1},
        {UserName: "carol", PhoneNum: "0933333333", Area: 2},
        {UserName: "dave", PhoneNum: "0944444444", Area: 2},
        {UserName: "eve", PhoneNum: "0955555555", Area: 1},
        // 重複訂單測試：alice area1 重複，應被 ON CONFLICT 略過
        {UserName: "alice", PhoneNum: "0911111111", Area: 1},
    }

    for _, td := range users {
        if err := service.PublishOrder(ctx, td); err != nil {
            fmt.Printf("PublishOrder 失敗 [%s]: %v\n", td.UserName, err)
            continue
        }
        fmt.Printf("PublishOrder 成功: %s\n", td.UserName)
    }

    fmt.Println("等待 worker 寫入 DB...")
    time.Sleep(3 * time.Second)
    fmt.Println("完成，去 DB 確認 orders table 應有 5 筆（alice 重複那筆被略過）")
}
