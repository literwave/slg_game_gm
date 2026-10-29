package gameclient

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 游戏服约定的 "玩家不存在" 错误码。
const codePlayerNotFound = 404

// httpClient 通过 HTTP 调用游戏服 GM 接口：POST {gm_addr}/gm/{cmd}，
// 请求头携带 X-GM-Timestamp 与 X-GM-Sign = hex(HMAC-SHA256(secret, timestamp + body))，
// 响应格式为 {"code":0,"msg":"","data":{}}。
type httpClient struct {
	hc     *http.Client
	secret []byte
}

func NewHTTP(secret string, timeout time.Duration) Client {
	return &httpClient{hc: &http.Client{Timeout: timeout}, secret: []byte(secret)}
}

func (c *httpClient) call(ctx context.Context, t Target, cmd string, req, out any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(ts))
	mac.Write(body)

	url := strings.TrimRight(t.Addr, "/") + "/gm/" + cmd
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-GM-Timestamp", ts)
	httpReq.Header.Set("X-GM-Sign", hex.EncodeToString(mac.Sum(nil)))

	resp, err := c.hc.Do(httpReq)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: http status %d", ErrUnavailable, resp.StatusCode)
	}

	var r struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("%w: decode response: %v", ErrUnavailable, err)
	}
	if r.Code == codePlayerNotFound {
		return ErrPlayerNotFound
	}
	if r.Code != 0 {
		return &CallError{Code: r.Code, Msg: r.Msg}
	}
	if out != nil && len(r.Data) > 0 {
		return json.Unmarshal(r.Data, out)
	}
	return nil
}

func (c *httpClient) Status(ctx context.Context, t Target) (*ServerStatus, error) {
	var s ServerStatus
	return &s, c.call(ctx, t, "status", struct{}{}, &s)
}

func (c *httpClient) SearchPlayers(ctx context.Context, t Target, keyword string) ([]Player, error) {
	var list []Player
	return list, c.call(ctx, t, "player/search", map[string]string{"keyword": keyword}, &list)
}

func (c *httpClient) GetPlayer(ctx context.Context, t Target, playerID int64) (*Player, error) {
	var p Player
	if err := c.call(ctx, t, "player/get", map[string]int64{"player_id": playerID}, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *httpClient) BanPlayer(ctx context.Context, t Target, req BanRequest) error {
	return c.call(ctx, t, "player/ban", req, nil)
}

func (c *httpClient) UnbanPlayer(ctx context.Context, t Target, playerID int64) error {
	return c.call(ctx, t, "player/unban", map[string]int64{"player_id": playerID}, nil)
}

func (c *httpClient) KickPlayer(ctx context.Context, t Target, playerID int64) error {
	return c.call(ctx, t, "player/kick", map[string]int64{"player_id": playerID}, nil)
}

func (c *httpClient) ModifyResources(ctx context.Context, t Target, playerID int64, changes map[string]int64) error {
	return c.call(ctx, t, "player/resources", map[string]any{"player_id": playerID, "changes": changes}, nil)
}

func (c *httpClient) SendMail(ctx context.Context, t Target, req MailRequest) error {
	return c.call(ctx, t, "mail/send", req, nil)
}

func (c *httpClient) PushAnnouncement(ctx context.Context, t Target, req AnnouncementRequest) error {
	return c.call(ctx, t, "announcement/push", req, nil)
}
