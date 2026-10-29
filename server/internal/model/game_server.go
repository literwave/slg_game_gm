package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServerStatus int

const (
	ServerMaintain ServerStatus = 0 // 维护
	ServerNormal   ServerStatus = 1 // 正常
	ServerBusy     ServerStatus = 2 // 火爆
	ServerClosed   ServerStatus = 3 // 关闭（合服/下线）
)

// GameServer 游戏区服。ServerID 为业务区服号（如 1001），GMAddr 为游戏服 GM 接口地址。
type GameServer struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServerID  int                `bson:"server_id" json:"server_id"`
	Name      string             `bson:"name" json:"name"`
	GMAddr    string             `bson:"gm_addr" json:"gm_addr"`
	Status    ServerStatus       `bson:"status" json:"status"`
	OpenAt    time.Time          `bson:"open_at" json:"open_at"`
	Remark    string             `bson:"remark" json:"remark"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
