package watcher

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

type AvailableTicketPayload struct {
	Type      string `json:"type"`
	AreaID    int    `json:"area_id"`
	Available int64  `json:"available"`
}

func StartHTTPClient(ctx context.Context, areaIDs []int, websocketHost string, websocketPort int) (<-chan error, error) {
	websocketEndpoint := fmt.Sprintf("http://%s:%d/broadcast/available-tickets", websocketHost, websocketPort)

	errCh := make(chan error, len(areaIDs))
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, areaID := range areaIDs {
		ch, err := service.WatchAvailableTickets(ctx, areaID)
		if err != nil {
			return nil, fmt.Errorf("watch available tickets area %d: %w", areaID, err)
		}

		go func(areaID int, ch <-chan int64) {
			for {
				select {
				case available, ok := <-ch:
					if !ok {
						return
					}

					if err := sendAvailableTicket(ctx, httpClient, websocketEndpoint, areaID, available); err != nil {
						select {
						case errCh <- fmt.Errorf("send available ticket area %d: %w", areaID, err):
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

func sendAvailableTicket(ctx context.Context, httpClient *http.Client, endpoint string, areaID int, available int64) error {
	payload := AvailableTicketPayload{
		Type:      "ticket_available",
		AreaID:    areaID,
		Available: available,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("websocket server status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
