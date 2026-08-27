package center

import (
	"github.com/bychannel/cherry-server/internal/data"
	"github.com/bychannel/cherry-server/nodes/center/db"
	"github.com/bychannel/cherry-server/nodes/center/module/account"
	"github.com/bychannel/cherry-server/nodes/center/module/ops"
	"github.com/cherry-game/cherry"
	cherryCron "github.com/cherry-game/components/cron"
)

func Run(profileFilePath, nodeID string) {
	app := cherry.Configure(
		profileFilePath,
		nodeID,
		false,
		cherry.Cluster,
	)

	app.Register(cherryCron.New())
	app.Register(data.New())
	app.Register(db.New())

	app.AddActors(
		&account.ActorAccount{},
		&ops.ActorOps{},
	)

	app.Startup()
}
