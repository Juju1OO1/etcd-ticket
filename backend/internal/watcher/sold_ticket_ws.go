package watcher

// 傳交易成功的紀錄但是前端未必用到
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"etcd-ticket/internal/service"
)

type SoldTicketPayload struct {
	Type     string `json:"type"`
	UserName string `json:"user_name"`
	Phone    string `json:"phone"`
	AreaID   int    `json:"area_id"`
}

// StartSoldTicketHTTPClient
//
// 監聽 sold ticket event
// 並轉送到 websocket server
func StartSoldTicketHTTPClient(
	ctx context.Context,
	areaIDs []int,
	websocketHost string,
	websocketPort int,
) (<-chan error, error) {

	websocketEndpoint := fmt.Sprintf(
		"http://%s:%d/broadcast/sold-tickets",
		websocketHost,
		websocketPort,
	)

	errCh := make(chan error, len(areaIDs))

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, areaID := range areaIDs {

		ch, err := service.WatchSoldTickets(ctx, areaID)
		if err != nil {
			return nil, fmt.Errorf(
				"watch sold tickets area %d: %w",
				areaID,
				err,
			)
		}

		go func(
			areaID int,
			ch <-chan service.SoldTicketEvent,
		) {

			for {
				select {

				case soldEvent, ok := <-ch:
					if !ok {
						return
					}

					if err := sendSoldTicket(
						ctx,
						httpClient,
						websocketEndpoint,
						soldEvent,
					); err != nil {

						select {
						case errCh <- fmt.Errorf(
							"send sold ticket area %d: %w",
							areaID,
							err,
						):

						case <-ctx.Done():
							return
						}
					}

				case <-ctx.Done():
					return
				}
			}

		}(areaID, ch)
	}

	return errCh, nil
}

func sendSoldTicket(

	ctx context.Context,
	httpClient *http.Client,
	endpoint string,
	event service.SoldTicketEvent,
) error {

	payload := SoldTicketPayload{
		Type:     "ticket_sold",
		UserName: event.UserName,
		Phone:    event.Phone,
		AreaID:   event.AreaID,
	}

	fmt.Printf(
		"[sendSoldTicket] user=%s area=%d\n",
		event.UserName,
		event.AreaID,
	)

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("[BRIDGE] POST sold user=%s area=%d -> wsserver\n", event.UserName, event.AreaID)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {

		respBody, _ := io.ReadAll(
			io.LimitReader(resp.Body, 1024),
		)

		return fmt.Errorf(
			"websocket server status %d: %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	return nil
}
