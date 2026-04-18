package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// 受信イベントをbindし、usecaseに変換して渡し、応答をWS向けのJSONになおす。
type WsCtl struct {
	flowUC    usecase.SearchFlowUC
	sessionUC usecase.SessionUC
	upgrader  websocket.Upgrader
}

// 接続開始時だけ使い、業務処理そのものは持たない。
func NewWsCtl(
	flowUC usecase.SearchFlowUC,
	sessionUC usecase.SessionUC,
) *WsCtl {
	return &WsCtl{
		flowUC:    flowUC,
		sessionUC: sessionUC,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

type wsClientEvent struct {
	Type      string          `json:"type"`
	SessionID uint            `json:"session_id"`
	Body      string          `json:"body,omitempty"`
	Diff      json.RawMessage `json:"diff,omitempty"`
}

type wsPatchPrefReq struct {
	Flavor     *int             `json:"flavor"`
	Acidity    *int             `json:"acidity"`
	Bitterness *int             `json:"bitterness"`
	Body       *int             `json:"body"`
	Aroma      *int             `json:"aroma"`
	Mood       *entity.Mood     `json:"mood"`
	Method     *entity.Method   `json:"method"`
	Scene      *entity.Scene    `json:"scene"`
	TempPref   *entity.TempPref `json:"temp_pref"`
	Excludes   []string         `json:"excludes"`
	Note       *string          `json:"note"`
}

type wsServerEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
}

// 接続を開始し、受信イベントをloopで処理する。
// controllerの責務は「bind → DTO変換 → usecase 呼び出し → WS応答整形」
func (ctl *WsCtl) Connect(c echo.Context) error {
	conn, err := ctl.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return ErrInvalidRequest
	}
	defer conn.Close()

	actor := actorFromCtx(c)
	sessionKey := sessionKeyFromWs(c)

	for {
		var ev wsClientEvent
		if err := conn.ReadJSON(&ev); err != nil {
			// クライアント切断や不正frameではloopを抜けて終了。
			return nil
		}

		switch strings.TrimSpace(ev.Type) {
		case "turn.add":
			if err := ctl.onAddTurn(conn, actor, sessionKey, ev); err != nil {
				_ = writeWsError(conn, err)
				continue
			}

		case "pref.patch":
			if err := ctl.onPatchPref(conn, actor, sessionKey, ev); err != nil {
				_ = writeWsError(conn, err)
				continue
			}

		case "session.close":
			if err := ctl.onCloseSession(conn, actor, sessionKey, ev); err != nil {
				_ = writeWsError(conn, err)
				continue
			}
			return nil

		case "ping":
			if err := conn.WriteJSON(wsServerEvent{
				Type: "done",
				Payload: map[string]string{
					"message": "pong",
				},
			}); err != nil {
				return nil
			}

		default:
			_ = writeWsCodeError(conn, "invalid_request", "unknown event type")
		}
	}
}

// turn.addイベントをAddTurnInに変換し、再検索結果を返す。
func (ctl *WsCtl) onAddTurn(
	conn *websocket.Conn,
	actor *entity.Actor,
	sessionKey string,
	ev wsClientEvent,
) error {
	out, err := ctl.flowUC.AddTurn(usecase.AddTurnIn{
		Actor:      actor,
		SessionID:  ev.SessionID,
		SessionKey: sessionKey,
		Body:       strings.TrimSpace(ev.Body),
	})
	if err != nil {
		return err
	}

	// まず追加されたturn自体を返す。
	if err := conn.WriteJSON(wsServerEvent{
		Type: "turn.added",
		Payload: map[string]interface{}{
			"turn": out.Turn,
		},
	}); err != nil {
		return err
	}

	// 次に候補更新結果を返す。
	if err := conn.WriteJSON(wsServerEvent{
		Type: "candidate.update",
		Payload: map[string]interface{}{
			"result": toSearchResultRes(out.Result),
		},
	}); err != nil {
		return err
	}

	// 最後に1イベント分の処理完了を通知。
	return conn.WriteJSON(wsServerEvent{
		Type: "done",
		Payload: map[string]uint{
			"session_id": ev.SessionID,
		},
	})
}

// pref.patch イベントをPatchPrefInに変換し、再検索結果を返す。
func (ctl *WsCtl) onPatchPref(
	conn *websocket.Conn,
	actor *entity.Actor,
	sessionKey string,
	ev wsClientEvent,
) error {
	var diff wsPatchPrefReq
	if err := json.Unmarshal(ev.Diff, &diff); err != nil {
		return ErrInvalidRequest
	}

	out, err := ctl.flowUC.PatchPref(usecase.PatchPrefIn{
		Actor:      actor,
		SessionID:  ev.SessionID,
		SessionKey: sessionKey,
		Flavor:     diff.Flavor,
		Acidity:    diff.Acidity,
		Bitterness: diff.Bitterness,
		Body:       diff.Body,
		Aroma:      diff.Aroma,
		Mood:       diff.Mood,
		Method:     diff.Method,
		Scene:      diff.Scene,
		TempPref:   diff.TempPref,
		Excludes:   diff.Excludes,
		Note:       diff.Note,
	})
	if err != nil {
		return err
	}

	if err := conn.WriteJSON(wsServerEvent{
		Type: "candidate.update",
		Payload: map[string]interface{}{
			"result": toSearchResultRes(out.Result),
		},
	}); err != nil {
		return err
	}

	return conn.WriteJSON(wsServerEvent{
		Type: "done",
		Payload: map[string]uint{
			"session_id": ev.SessionID,
		},
	})
}

// session.closeイベントを受けてsessionを終了。
// private / guest の最終判定自体はusecase側。
func (ctl *WsCtl) onCloseSession(
	conn *websocket.Conn,
	actor *entity.Actor,
	sessionKey string,
	ev wsClientEvent,
) error {
	err := ctl.sessionUC.CloseSession(usecase.CloseSessionIn{
		Actor:      actor,
		SessionID:  ev.SessionID,
		SessionKey: sessionKey,
	})
	if err != nil {
		return err
	}

	return conn.WriteJSON(wsServerEvent{
		Type: "done",
		Payload: map[string]interface{}{
			"session_id": ev.SessionID,
			"message":    "session closed",
		},
	})
}

// guest再開用のsession keyをquery/headerから取得。
func sessionKeyFromWs(c echo.Context) string {
	if v := strings.TrimSpace(c.QueryParam("session_key")); v != "" {
		return v
	}
	return strings.TrimSpace(c.Request().Header.Get(HeaderSessionKey))
}

// writeWsError は usecase / repository エラーを WS 用 error event に変換して返す。
func writeWsError(conn *websocket.Conn, err error) error {
	_, code, msg := mapError(err)
	return conn.WriteJSON(wsServerEvent{
		Type:    "error",
		Code:    code,
		Message: msg,
	})
}

// controller側で即時に返したいエラーコードをそのままWSへ返す。
func writeWsCodeError(conn *websocket.Conn, code string, msg string) error {
	return conn.WriteJSON(wsServerEvent{
		Type:    "error",
		Code:    code,
		Message: msg,
	})
}
