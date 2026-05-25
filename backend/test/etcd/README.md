## 測試
這個檔案測試etcd的功能。
1. docker-compose，建立三個節點的叢集
2. client封裝etcd server，用來與etcd通訊(CRUD、Watch)
3. ticket_service負責搶票的邏輯