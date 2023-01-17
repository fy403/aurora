package api

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// ctx.request.body格式解析至参数, 并调用对应函数句柄
func (api *Api) WarpHandle(fn interface{}) gin.HandlerFunc {
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func {
		panic("value not func")
	}
	typ := val.Type()
	switch typ.NumIn() {
	case 1: // func(done *WaitConn)
		return func(ctx *gin.Context) {
			api.transBegin(ctx, fn)
		}
	case 2: // func(done *WaitConn, req struct)
		return func(ctx *gin.Context) {
			inValue, err := api.getJsonBody(ctx, typ.In(1))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": "Json unmarshal failed!",
					"error":   err.Error(),
				})
				return
			}

			api.transBegin(ctx, fn, inValue)
		}
	default:
		panic("func symbol error")
	}
}

func (api *Api) transBegin(ctx *gin.Context, fn interface{}, args ...reflect.Value) {
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func {
		panic("value not func")
	}
	typ := val.Type()
	if typ.NumIn() != len(args)+1 {
		panic("func argument error")
	}

	route := api.getCurrentRoute(ctx)
	wait := NewWaitConn(ctx, route)
	if err := api.taskQueue.SubmitTask(webTask(func() {
		ok := api.checkToken(ctx, route)
		if !ok {
			wait.SetResult("Token验证失败", nil)
			wait.Done(401)
			return
		}
		val.Call(append([]reflect.Value{reflect.ValueOf(wait)}, args...))
	})); err != nil {
		wait.SetResult("访问人数过多", nil)
		wait.Done()
	}
	wait.Wait()

	ctx.JSON(wait.code, wait.result)
}

func (api *Api) getCurrentRoute(ctx *gin.Context) string {
	return ctx.FullPath()
}

func (api *Api) getJsonBody(ctx *gin.Context, inType reflect.Type) (inValue reflect.Value, err error) {
	if inType.Kind() == reflect.Ptr {
		inValue = reflect.New(inType.Elem())
	} else {
		inValue = reflect.New(inType)
	}
	// 常用解析JSON数据的方法
	// 解析ctx.request.body到inValue
	if err = ctx.ShouldBindJSON(inValue.Interface()); err != nil {
		return
	}
	if inType.Kind() != reflect.Ptr {
		inValue = inValue.Elem()
	}
	return
}
