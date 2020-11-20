package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/lib/lotuslog"
	"github.com/guoxiaopeng875/lotus-adapter/apiwrapper"
	lcli "github.com/guoxiaopeng875/lotus-adapter/cmd/cli"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"gopkg.in/resty.v1"
	"net/http"
	"os"
	"time"
)

var log = logging.Logger("main")

func main() {
	lotuslog.SetupLogLevels()

	local := []*cli.Command{
		runCmd,
	}

	app := &cli.App{
		Name:     "lotus-monitor",
		Usage:    "lotus monitor",
		Version:  build.UserVersion(),
		Commands: local,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				EnvVars: []string{"LOTUS_PATH"},
				Value:   "~/.lotus", // TODO: Consider XDG_DATA_HOME
			},
			&cli.StringFlag{
				Name:    "miner-repo",
				EnvVars: []string{"LOTUS_MINER_PATH", "LOTUS_STORAGE_PATH"},
				Value:   "~/.lotusminer", // TODO: Consider XDG_DATA_HOME
				Usage:   fmt.Sprintf("Specify miner repo path.  env(LOTUS_STORAGE_PATH) are DEPRECATION, will REMOVE SOON"),
			},
		},
	}
	app.Setup()

	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "Start lotus monitor",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "proxy",
			Usage: "set monitor-center url",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "proxy-token",
			Usage: "set monitor-center token",
			Value: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiZnhpbmdnb25nIn0.vm8FX1p_I3sLlFoHtLqQrv7sITuLYOVgo-9dTbxIcUFUcwsHrYbgwU4qR1G8Sc7waoVETjkGyjf5I1buTAzvEA6ACAil4yxGMRQKmTH4IT1t2nffPHeD5qqlj2fetceue8tI2gAWB_qhrg-1khGDhU55CzsGGmrVpgRhAZfVww_AA1VvXhRxiPJdD6klFjai8eg4zWVP956lL_YmEc50mfvV-XRzAeE0rgC-gowcZH4_XjiSn3i4OaD3iPd2oEV0zS8W3vNz46sWxlefKBaGqK3mHitdKI3_yl85l29PhxdGpbIrZhhhnKX1ywu4RjBnXxHW6whodJdPNXJIb49ikw",
		},
		&cli.DurationFlag{
			Name:  "interval",
			Usage: "set monitor interval",
			Value: time.Minute,
		},
		&cli.StringFlag{
			Name:    "log-level",
			EnvVars: []string{"GOLOG_LOG_LEVEL"},
			Value:   "info",
		},
	},
	Action: func(cctx *cli.Context) error {
		log.Info("Starting lotus monitor")
		go func() {
			http.ListenAndServe(":8875", nil) //nolint:errcheck
		}()

		logLvl := cctx.String("log-level")
		if err := logging.SetLogLevel("*", logLvl); err != nil {
			return err
		}

		ctx := lcli.ReqContext(cctx)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		minerApi, mCloser, err := lcli.GetStorageMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer mCloser()

		mAddr, err := minerApi.ActorAddress(ctx)
		if err != nil {
			return err
		}
		processor := NewProcessor(map[address.Address]*apiwrapper.LotusAPIWrapper{
			mAddr: apiwrapper.NewLotusAPIWrapper(api, minerApi),
		}, resty.New(), cctx.String("proxy"), map[string]string{
			"name":  "fxinggong",
			"token": cctx.String("proxy-token"),
		})

		tick := time.Tick(cctx.Duration("interval"))
		for {
			select {
			case <-tick:
				log.Debug("push lotus miner info")
				if err := processor.PushAll(); err != nil {
					log.Errorf("push lotus miner info failed, %w", err)
				}
			case <-ctx.Done():
				log.Warn("Shutdown...")
				return nil
			}
		}

	},
}
