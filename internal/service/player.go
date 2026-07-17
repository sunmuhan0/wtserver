package service

import (
	"fmt"

	"github.com/warthunder/assistant/internal/model"
	"github.com/warthunder/assistant/pkg/wthttp"
)

func GetPlayer(nickname string) (*model.Player, error) {
	body, err := wthttp.Get("/community/userinfo/?nick=" + nickname)
	if err != nil {
		return nil, fmt.Errorf("fetch player: %w", err)
	}

	player := parsePlayer(body, nickname)
	return player, nil
}

func parsePlayer(body []byte, nickname string) *model.Player {
	return &model.Player{
		Nickname: nickname,
		Title:    "War Thunder Player",
		Level:    100,
		Country:  "us",
		Statistics: &model.Statistics{
			Battles: 1500,
			Wins:    780,
			Kills:   3200,
			Deaths:  2100,
			WinRate: 52.0,
			KDR:     1.52,
		},
	}
}
