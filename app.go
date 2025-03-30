package main

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/rs/zerolog"
	"github.com/tarantool/go-tarantool/v2"
)


type application struct {
	config	 	Config
	logger	 	zerolog.Logger
	mmClient	*model.Client4
	mmWebSocket *model.WebSocketClient
	mmUser 		*model.User
	mmChannel 	*model.Channel
	mmTeam 		*model.Team
	mmTarantool *tarantool.Connection
}