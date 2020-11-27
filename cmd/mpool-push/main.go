package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/lib/lotuslog"
	"github.com/filecoin-project/specs-actors/actors/builtin/multisig"
	lcli "github.com/guoxiaopeng875/lotus-adapter/cmd/cli"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"os"
)

var log = logging.Logger("main")

func main() {
	lotuslog.SetupLogLevels()

	local := []*cli.Command{
		pushCmd,
	}

	app := &cli.App{
		Name:     "mpool-push",
		Usage:    "lotus push message",
		Version:  build.UserVersion(),
		Commands: local,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				EnvVars: []string{"LOTUS_PATH"},
				Value:   "~/.lotus", // TODO: Consider XDG_DATA_HOME
			},
		},
	}
	app.Setup()

	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}

var pushCmd = &cli.Command{
	Name:  "push",
	Usage: "push message",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "msg",
			Usage: "message data",
			Value: "",
		},
		&cli.BoolFlag{
			Name:  "wait",
			Usage: "wait message on chain",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "propose",
			Usage: "check mutisig propose message",
			Value: false,
		},
		&cli.Uint64Flag{
			Name:  "confidence",
			Usage: "wait message confidence depth",
			Value: 0,
		},
	},
	Action: func(cctx *cli.Context) error {

		ctx := lcli.ReqContext(cctx)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		msgBytes, err := hex.DecodeString(cctx.String("msg"))
		if err != nil {
			return err
		}
		sm, err := types.DecodeSignedMessage(msgBytes)
		if err != nil {
			return err
		}
		cid, err := api.MpoolPush(ctx, sm)
		if err != nil {
			return err
		}
		fmt.Println(cid.String())
		if cctx.Bool("wait") {
			wait, err := api.StateWaitMsg(ctx, cid, cctx.Uint64("confidence"))
			if err != nil {
				return err
			}
			if wait.Receipt.ExitCode != 0 {
				return fmt.Errorf("proposal returned exit %d", wait.Receipt.ExitCode)
			}
			if cctx.Bool("propose") {
				var retval multisig.ProposeReturn
				if err := retval.UnmarshalCBOR(bytes.NewReader(wait.Receipt.Return)); err != nil {
					return fmt.Errorf("failed to unmarshal propose return value: %w", err)
				}

				fmt.Printf("Transaction ID: %d\n", retval.TxnID)
				if retval.Applied {
					fmt.Printf("Transaction was executed during propose\n")
					fmt.Printf("Exit Code: %d\n", retval.Code)
					fmt.Printf("Return Value: %x\n", retval.Ret)
				}
			}
		}

		return nil
	},
}
