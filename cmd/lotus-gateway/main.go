package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/repo"
	"github.com/gorilla/mux"
	lcli "github.com/guoxiaopeng875/lotus-adapter/cmd/cli"
	logging "github.com/ipfs/go-log/v2"
	"github.com/patrickmn/go-cache"
	"github.com/urfave/cli/v2"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/lib/lotuslog"
)

var log = logging.Logger("main")

func main() {
	lotuslog.SetupLogLevels()

	local := []*cli.Command{
		runCmd,
	}

	app := &cli.App{
		Name:     "lotus-cached-gateway",
		Usage:    "lotus cached api",
		Version:  build.UserVersion(),
		Commands: local,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "gw-repo",
				EnvVars: []string{"LOTUS_GW_PATH"},
				Value:   "~/.lotusgw",
			},
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
	Usage: "Start lotus cached api",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "listen",
			Usage: "host address and port the miner api will listen on",
			Value: "0.0.0.0:9988",
		},
		&cli.DurationFlag{
			Name:  "expiration",
			Usage: "set cache expiration",
			Value: 10 * time.Second,
		},
		&cli.DurationFlag{
			Name:  "interval",
			Usage: "set cache cleanup interval",
			Value: time.Minute,
		},
	},
	Action: func(cctx *cli.Context) error {
		log.Info("Starting lotus gateway")
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

		address := cctx.String("listen")
		mux := mux.NewRouter()

		log.Info("Setting up API endpoint at " + address)

		rpcServer := jsonrpc.NewServer()
		c := cache.New(cctx.Duration("expiration"), cctx.Duration("interval"))

		fs, err := repo.NewFS(cctx.String("gw-repo"))
		if err != nil {
			return err
		}
		ks, err := fs.Lock(repo.FullNode)
		if err != nil {
			return err
		}
		secret, err := APISecret(ks.(types.KeyStore), ks)
		if err != nil {
			return err
		}
		gwAPI := NewCachedFullNode(api, minerApi, c, secret)
		rpcServer.Register("Filecoin", gwAPI)

		mux.Handle("/rpc/v0", rpcServer)
		mux.PathPrefix("/").Handler(http.DefaultServeMux)

		ah := &auth.Handler{
			Verify: gwAPI.AuthVerify,
			Next:   mux.ServeHTTP,
		}

		srv := &http.Server{
			Handler: ah,
		}

		go func() {
			<-ctx.Done()
			log.Warn("Shutting down...")
			if err := srv.Shutdown(context.TODO()); err != nil {
				log.Errorf("shutting down RPC server failed: %s", err)
			}
			log.Warn("Graceful shutdown successful")
		}()

		nl, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}

		return srv.Serve(nl)
	},
}
