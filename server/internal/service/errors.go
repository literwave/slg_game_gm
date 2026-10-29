package service

import (
	"errors"

	"slg_game_gm/server/internal/gameclient"
	"slg_game_gm/server/internal/repository"
	"slg_game_gm/server/pkg/errcode"
)

// repoErr 将存储层错误转换为业务错误，未知错误原样返回。
func repoErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrNotFound):
		return errcode.NotFound
	case errors.Is(err, repository.ErrDuplicate):
		return errcode.Conflict
	default:
		return err
	}
}

// gameErr 将游戏服调用错误转换为业务错误。
func gameErr(err error) error {
	var callErr *gameclient.CallError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gameclient.ErrPlayerNotFound):
		return errcode.NotFound.WithMsg("玩家不存在")
	case errors.Is(err, gameclient.ErrUnavailable):
		return errcode.GameServerUnavailable
	case errors.As(err, &callErr):
		return errcode.GameCallFailed.WithMsg("游戏服返回错误: " + callErr.Msg)
	default:
		return err
	}
}
