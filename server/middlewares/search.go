package middlewares

import (
	"github.com/AlliotTech/openalist/internal/conf"
	"github.com/AlliotTech/openalist/internal/errs"
	"github.com/AlliotTech/openalist/internal/setting"
	"github.com/AlliotTech/openalist/server/common"
	"github.com/gin-gonic/gin"
)

func SearchIndex(c *gin.Context) {
	mode := setting.GetStr(conf.SearchIndex)
	if mode == "none" {
		common.ErrorResp(c, errs.SearchNotAvailable, 500)
		c.Abort()
	} else {
		c.Next()
	}
}
