package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"slg_game_gm/server/pkg/errcode"
	"slg_game_gm/server/pkg/response"
)

// objectIDParam 解析路径参数中的 ObjectID，失败时已写入错误响应。
func objectIDParam(c *gin.Context, name string) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param(name))
	if err != nil {
		response.Error(c, errcode.InvalidParams.WithMsg("无效的ID"))
		return id, false
	}
	return id, true
}

// playerParams 解析 /:server_id/:player_id 路径参数。
func playerParams(c *gin.Context) (int, int64, bool) {
	serverID, err1 := strconv.Atoi(c.Param("server_id"))
	playerID, err2 := strconv.ParseInt(c.Param("player_id"), 10, 64)
	if err1 != nil || err2 != nil {
		response.Error(c, errcode.InvalidParams.WithMsg("无效的区服号或玩家ID"))
		return 0, 0, false
	}
	return serverID, playerID, true
}

// bind 根据请求方法绑定 query 或 JSON body，失败时已写入错误响应。
func bind(c *gin.Context, obj any) bool {
	var err error
	if c.Request.Method == "GET" {
		err = c.ShouldBindQuery(obj)
	} else {
		err = c.ShouldBindJSON(obj)
	}
	if err != nil {
		response.BadRequest(c, err)
		return false
	}
	return true
}
