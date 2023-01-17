package api

import (
	"aurora/internal/auth"
	"aurora/internal/log"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

type authHandler struct {
}

func (*authHandler) login(wait *WaitConn, req struct {
	Username string `json:"username"`
	Password string `json:"password"`
}) {
	log.Runtime().Infof("%s %v", wait.GetRoute(), req)
	defer func() { wait.Done() }()

	users := auth.DefaultUsers()
	if password, ok := users[req.Username]; !ok || req.Password != password {
		wait.SetCode(http.StatusUnauthorized)
		wait.SetResult("用户名或密码错误", "")
		return
	}

	// generate session
	if len(wait.ctx.Request.Header["Authorization"]) > 0 {
		wait.ctx.Request.Header.Add("Cookie", wait.ctx.Request.Header["Authorization"][0])
	}

	session, err := auth.DefaultStore().Get(wait.ctx.Request, "aurora_session")
	if err != nil {
		wait.SetCode(http.StatusUnauthorized)
		wait.SetResult(fmt.Sprintf("Failed to execute auth.DefaultStore().Get: %v", err), "")
		return
	}

	session.Options = auth.DefaultSessionOption()
	session.Values["UUID"] = uuid.New().String()
	session.Values["User"] = req.Username

	keyPairs := auth.DefaultKeyPairs()
	codes := securecookie.CodecsFromPairs(keyPairs...)
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values,
		codes...)

	if err != nil {
		wait.SetCode(http.StatusUnauthorized)
		wait.SetResult(fmt.Sprintf("Failed to execute securecookie.EncodeMulti: %v", err), "")
		return
	}

	wait.SetResult("", struct {
		User  UserMeta `json:"user"`
		Token string   `json:"token"`
	}{
		User: UserMeta{
			Roles: []string{
				"admin",
			},
			User: UserDetail{
				CreateTime: time.Now().UTC().GoString(),
				Enabled:    true,
				NickName:   "管理员",
				UserName:   session.Values["User"].(string),
			},
		},
		Token: sessions.NewCookie(session.Name(), encoded, session.Options).String(),
	})
}

func (*authHandler) info(wait *WaitConn) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	if len(wait.ctx.Request.Header["Authorization"]) > 0 {
		wait.ctx.Request.Header.Add("Cookie", wait.ctx.Request.Header["Authorization"][0])
	}
	session, err := auth.DefaultStore().Get(wait.ctx.Request, "aurora_session")

	if session.IsNew || err != nil {
		wait.SetCode(http.StatusUnauthorized)
		wait.SetResult(fmt.Sprintf("No permission err is %s", err.Error()), "")
		return
	}
	wait.SetResult("", UserMeta{
		Roles: []string{
			"admin",
		},
		User: UserDetail{
			CreateTime: time.Now().UTC().GoString(),
			Enabled:    true,
			NickName:   "管理员",
			UserName:   session.Values["User"].(string),
		},
	})
}

func (*authHandler) logout(wait *WaitConn) {
	log.Runtime().Infof("%s", wait.GetRoute())
	defer func() { wait.Done() }()
	wait.SetResult("登出成功", "")
}

type UserMeta struct {
	Roles []string   `json:"roles"`
	User  UserDetail `json:"user"`
}

type UserDetail struct {
	CreateTime string `json:"createTime"`
	Enabled    bool   `json:"enabled"`
	NickName   string `json:"nickName"`
	UserName   string `json:"username"`
}
