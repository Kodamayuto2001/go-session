package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
)

//	セッションマネージャ構造体
type Manager struct {
	//	キーが文字列、要素がinterface型（C言語のvoid*型）のマップ
	database map[string]interface{}
}

var mg Manager

func NewManager() *Manager {
	return &mg
}

//	セッションIDの発行
func (m *Manager) NewSessionID() string {
	//	cryptoパッケージの乱数生成機能を使用して、64バイト長のランダム文字列を生成することでユニークであることを保証しようとしている
	b := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	//	base64にエンコーディング
	return base64.URLEncoding.EncodeToString(b)
}

//	新規セッションの生成
//	セッションの実体化自体は、session.goファイル側に記述するべき内容
//	セッションマネージャ側で行う新規セッション生成処理は、クライアントのCookieをチェックしてすでにセッションが開始済みでないことを確認し、session.goファイルの処理を呼び出す制御処理となる
func (m *Manager) New(r *http.Request, cookieName string) (*Session, error) {
	//	(*Request) Cookie関数は、リクエストに含まれる名前付きCookieのうち、引数で指定された名前のCookieを返す。
	//	これにより、リクエストのCookieをチェックし、すでに該当のクライアントのセッションが開始されていないかをチェックしている。
	cookie, err := r.Cookie(cookieName)
	if err == nil && m.Exists(cookie.Value) {
		return nil, errors.New("sessionIDはすでに発行されています")
	}

	session := NewSession(m, cookieName)
	//	NewSession関数で新しいセッション構造体を生成し、構造体に必要な情報をセットしている。
	//	構造体の定義及びNewSession関数の処理内容は、session.goファイル側で記述
	session.ID = m.NewSessionID()
	session.request = r 

	return session, nil
}

//	セッション情報の保存
func (m *Manager) Save(r *http.Request, w http.ResponseWriter, session *Session) error {
	//	セッションマネージャ構造体のマップに対し、セッションIDとキーとして、セッション構造体を要素に代入している。
	//	サーバーサイドへのセッション情報保存がこれにあたる
	m.database[session.ID] = session

	//	同一のセッションIDを保持するCookieを生成する。
	c := &http.Cookie{
		Name:	session.Name(),
		Value:	session.ID,
		Path:	"/",
	}

	//	クライアントのCookieに保存している
	http.SetCookie(session.writer, c)
	return nil
}

//	既存セッションの存在チェック
//	引数で渡されたセッションIDが存在するかどうかをチェックしている
//	クライアントリクエストに含まれるセッションIDの存在のチェックのほか、新規IDを発行したときの重複チェックにも用いることができる。
func (m *Manager) Exists(sessionID string) bool {
	_, r := m.database[sessionID]
	return r 
}

//	既存セッションの取得
//	サーバサイドに保存されている情報をクライアントのリクエストと紐づけて取得している処理
//	セッション管理によってステートフルなレスポンスを実現している処理がここ
func (m *Manager) Get(r *http.Request, cookieName string) (*Session, error) {
	//	新規セッションの生成時と同様に、Cookie情報を取得している。
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err 
	}

	sessionID := cookie.Value

	//	cookie情報からセッション情報を取得
	buffer, exists := m.database[sessionID]
	if !exists {
		return nil, errors.New("無効なセッションIDです")
	}

	session := buffer.(*Session)
	session.request = r
	return session, nil 
}

//	セッションの破棄
func (m *Manager) Destroy(sessionID string) {
	//	セッションマネージャ構造体のマップから、セッションIDをキーにdelete関数で要素を削除している
	delete(m.database, sessionID)
}