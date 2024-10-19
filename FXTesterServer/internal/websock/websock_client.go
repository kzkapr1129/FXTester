package websock

import (
	"errors"
	"fxtester/internal/common"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var ErrBusy = errors.New("websocket: busy")
var ErrInvUUID = errors.New("websocket: invalid uuid")
var ErrAlreadyCompleted = errors.New("websocket: already completed")

type Message struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

type WebsockClient struct {
	channels map[string]chan any
	m        sync.Mutex
	wsu      websocket.Upgrader
}

func NewWebsockClient() *WebsockClient {
	return &WebsockClient{
		channels: map[string]chan any{},
		wsu: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 本番環境では適切にオリジンをチェックする必要があります
			},
		},
	}
}

func (p *WebsockClient) CommunicateViaWs(ctx echo.Context) error {
	// UUIDの存在チェックとチャネルの取得
	ch, err := func() (chan any, error) {
		p.m.Lock()
		defer p.m.Unlock()

		uuid := ctx.Param("uuid")
		if uuid == "" {
			return nil, ErrInvUUID
		}

		if ch, ok := p.channels[uuid]; !ok {
			return nil, ErrAlreadyCompleted
		} else {
			return ch, nil
		}
	}()
	if err != nil {
		return err
	}

	// Websocketのハンドシェイク開始
	ws, err := p.wsu.Upgrade(ctx.Response().Writer, ctx.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// チャネルをポーリングしWebsocketに書き込む
	for {
		v, ok := <-ch
		if !ok {
			return nil
		}

		if err := ws.WriteJSON(v); err != nil {
			return err
		}
	}
}

func (p *WebsockClient) NewWs() (func(action string, message any), func(), string, error) {
	p.m.Lock()
	defer p.m.Unlock()

	// 最大接続数の閾値チェック
	numConnections := len(p.channels)
	if common.GetConfig().Websocket.MaxConnections <= numConnections {
		return nil, nil, "", ErrBusy
	}

	// 新規UUIDの払い出し
	uuid := uuid.NewString()
	ch := make(chan any)
	p.channels[uuid] = ch

	// チャネル書き込みの関数
	writer := func(action string, message any) {
		msg := Message{
			Action:  action,
			Payload: message,
		}
		select {
		case ch <- msg:
		default:
		}
	}

	// チャネルと払い出したUUIDの破棄をする関数
	closer := func() {
		p.m.Lock()
		defer p.m.Unlock()
		if ch, ok := p.channels[uuid]; ok {
			close(ch)
			delete(p.channels, uuid)
		}
	}

	return writer, closer, uuid, nil
}
