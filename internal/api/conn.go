package api

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// 应答结构
type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type WaitConn struct {
	code     int
	ctx      *gin.Context
	route    string
	result   Result
	done     chan struct{}
	doneOnce sync.Once
}

func NewWaitConn(ctx *gin.Context, route string) *WaitConn {
	return &WaitConn{
		ctx:   ctx,
		code:  http.StatusOK,
		route: route,
		done:  make(chan struct{}),
	}
}

func (this *WaitConn) Done(code ...int) {
	this.doneOnce.Do(func() {
		if len(code) > 0 {
			this.code = code[0]
		}
		if this.code == 0 {
			this.code = 200
		}
		// 注入结果
		this.result.Code = this.code
		// 此处无需手动注入
		// this.ctx.JSON(this.code, this.result)
		close(this.done)
	})
}

func (this *WaitConn) GetRoute() string {
	return this.route
}

func (this *WaitConn) SetCode(code int) {
	this.code = code
}

func (this *WaitConn) Context() *gin.Context {
	return this.ctx
}

func (this *WaitConn) SetResult(message string, data interface{}) {
	this.result.Message = message
	this.result.Data = data
}

func (this *WaitConn) Wait() {
	<-this.done
}

type webTask func()

func (t webTask) Do() {
	t()
}
