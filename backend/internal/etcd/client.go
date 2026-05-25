package etcd

// 這個檔案提供給需要建立 client 並對etcd server進行操作的檔案一個入口
// 外部檔案使用client就能對server下指令進行操作

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"sigs.k8s.io/yaml"
)

// 對應 config.yaml 最外層 key "etcd:"
type Config struct {
	Etcd EtcdConfig `yaml:"etcd"`
}

type EtcdConfig struct {
	Endpoints   []string `yaml:"endpoints"`
	DialTimeout int      `yaml:"dial_timeout"`
}

// 讀取config.yaml並解析
func loadConfig(path string) (EtcdConfig, error) {
	rawbyte, err := os.ReadFile(path)
	if err != nil {
		return EtcdConfig{}, err
	}

	cfg := Config{}
	if err := yaml.Unmarshal(rawbyte, &cfg); err != nil {
		return EtcdConfig{}, err
	}

	return cfg.Etcd, nil
}

// 用設定建立一個新的 etcd client
func newClient(cfg EtcdConfig) (*clientv3.Client, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Singleton：整個後端只建立一個 client 實例
var (
	once      sync.Once
	singleton *clientv3.Client
)

// GetClient 供其他套件（Txn、Watch）取得 etcd client
func getClient() *clientv3.Client {
	once.Do(func() {
		cfg, err := loadConfig("./configs/config.yaml")
		if err != nil {
			panic(fmt.Sprintf("etcd config 讀取失敗: %v", err))
		}

		c, err := newClient(cfg)
		if err != nil {
			panic(fmt.Sprintf("etcd 連線失敗: %v", err))
		}

		singleton = c
		fmt.Println("etcd connected:", cfg.Endpoints)
	})
	return singleton
}

// Close 在程式結束時釋放連線，於 main.go 呼叫
func Close() {
	if singleton != nil {
		singleton.Close()
	}
}

// Client 封裝 etcd 的 CRUD 與 Watch 操作
type Client struct {
	Cli       *clientv3.Client
	opTimeout time.Duration
}

func New() *Client {
	return &Client{
		Cli:       getClient(),
		opTimeout: 3 * time.Second,
	}
}

func (c *Client) timeoutCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.opTimeout)
}

func (c *Client) Put(ctx context.Context, key, value string) error {
	ctx, cancel := c.timeoutCtx(ctx)
	defer cancel()

	_, err := c.Cli.Put(ctx, key, value)
	return err
}

func (c *Client) Get(ctx context.Context, key string) (string, bool, error) {
	ctx, cancel := c.timeoutCtx(ctx)
	defer cancel()

	resp, err := c.Cli.Get(ctx, key)
	if err != nil {
		return "", false, err
	}
	if len(resp.Kvs) == 0 {
		return "", false, nil
	}

	return string(resp.Kvs[0].Value), true, nil
}

func (c *Client) GetPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	ctx, cancel := c.timeoutCtx(ctx)
	defer cancel()

	resp, err := c.Cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}
	return result, nil
}

// 只有回傳數量，不用每次都map，省時間、記憶體
func (c *Client) CountPrefix(ctx context.Context, prefix string) (int64, error) {
	ctx, cancel := c.timeoutCtx(ctx)
	defer cancel()

	resp, err := c.Cli.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	ctx, cancel := c.timeoutCtx(ctx)
	defer cancel()

	_, err := c.Cli.Delete(ctx, key)
	return err
}

func (c *Client) DeletePrefix(ctx context.Context, prefix string) error {
	ctx, cancel := c.timeoutCtx(ctx)
	defer cancel()

	_, err := c.Cli.Delete(ctx, prefix, clientv3.WithPrefix())
	return err
}

func (c *Client) WatchPrefix(ctx context.Context, prefix string) clientv3.WatchChan {
	return c.Cli.Watch(ctx, prefix, clientv3.WithPrefix())
}
