package sessions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/context"
)

const (
	DefaultSessionName = "default-session"
	DefaultCookieName = "default-cookie"
)

//	セッション構造体
type Session struct {
	cookieName	string
	ID			string
	manager		*Manager
	request		*http.Request
	writer		http.ResponseWriter
	Values		map[string]interface{}
}

//	新規セッション生成
func NewSession(manager *Manager, cookieName string) *Session {
	return &Session{
		cookieName:	cookieName,
		manager:	manager,
		Values:		map[string]interface{}{}
	}
}

//	セッションの開始
func StartSession(sessionName, cookieName string, manager *Manager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var session *Session
		var err error
		session, err = manager.Get(ctx.Request, cookieName)
		if err != nil {
			session, err = manager.New(ctx.Request, cookieName)
			if err != nil {
				println(err.Error())
				ctx.About()
			}
		}
		session.writer = ctx.Writer 
		ctx.Set(sessionName, session)
		defer context.Clear(ctx.Request)
		ctx.Next()
	}
}

//	デフォルトセッションの開始
func StartDefaultSession(manager *Manager) gin.HandlerFunc {
	return StartSession(DefaultSessionName, DefaultCookieName, manager)
}

//	セッションの取得
func GetSession(c *gin.Context, sessionName string) *Session {
	return c.MustGet(sessionName).(*Session)
}

//	デフォルトセッションの取得
func GetDefaultSession(c *gin.Context) *Session {
	return GetSession(c, DefaultSessionName)
}

//	セッションの保存
func (s *Session) Save() error {
	return s.manager.Save(s.request, s.writer, s)
}

//	セッション名の取得
func (s *Session) Name() string {
    return s.cookieName
}

//	セッション変数値の取得
func (s *Session) Get(key string) (interface{}, bool) {
	ret, exists := s.Values[key]
	return ret, exists
}

//	セッション変数値のセット
func (s *Session) Set(key string, val interface{}) {
	s.Values[key] = val 
}

//	セッション変数の削除
func (s *Session) Delete(key string) {
	delete(s.Values, key)
}

//	セッションの削除
func (s *Session) Terminate() {
	s.manager.Destroy(s.ID)
}