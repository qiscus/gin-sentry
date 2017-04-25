package recovery

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	raven "github.com/qiscus/raven-go"
)

func Recovery(dsn string, onlyCrashes bool) gin.HandlerFunc {
	err := raven.SetDSN(dsn)
	if err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		defer func() {
			flags := map[string]string{
				"endpoint": c.Request.RequestURI,
			}
			if rval := recover(); rval != nil {
				debug.PrintStack()
				rvalStr := fmt.Sprint(rval)
				packet := raven.NewPacket(rvalStr, raven.NewException(errors.New(rvalStr), raven.NewStacktrace(2, 3, nil)))
				raven.Capture(packet, flags)
				c.Writer.WriteHeader(http.StatusInternalServerError)
			}
			if !onlyCrashes {
				for _, item := range c.Errors {
					packet := raven.NewPacket(item.Error(), &raven.Message{
						Message: item.Error(),
						Params:  []interface{}{item.Meta}},
					)
					raven.Capture(packet, flags)
				}
			}
		}()
		c.Next()
	}
}
