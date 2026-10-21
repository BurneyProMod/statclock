// Package display sends values to the ESPHome LED matrix through its
// built-in web server (see web_server and the "elo" number in esp32.yaml).
package display

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

// SetElo shows elo on the device at deviceURL, e.g. "http://faceit-clock.local".
func SetElo(ctx context.Context, deviceURL string, elo int) error {
	u := strings.TrimRight(deviceURL, "/") + "/number/elo/set?value=" + strconv.Itoa(elo)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send elo to device: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("send elo to device: http %d", res.StatusCode)
	}
	return nil
}
