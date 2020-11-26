package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/actors"
	multisig0 "github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/multisig"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/wallet"
	"github.com/filecoin-project/lotus/lib/lotuslog"
	"github.com/filecoin-project/lotus/node/repo"
)

var log = logging.Logger("main")

var (
	wlAPI api.WalletAPI
	once  sync.Once
)

const FlagWalletRepo = "wallet-repo"

func main() {
	lotuslog.SetupLogLevels()

	local := []*cli.Command{
		walletNew,
		walletList,
		walletDelete,
		walletImport,
		walletExport,
		walletSign,
		msigProposeCmd,
		msigApproveCmd,
	}

	app := &cli.App{
		Name:    "lotus-wallet",
		Usage:   "Basic external wallet",
		Version: build.UserVersion(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    FlagWalletRepo,
				EnvVars: []string{"WALLET_PATH"},
				Value:   "~/.lotuswallet", // TODO: Consider XDG_DATA_HOME
			},
		},

		Commands: local,
	}
	app.Setup()

	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}

func NewWalletAPI(cctx *cli.Context) error {
	var err error
	once.Do(func() {
		wlAPI, err = newWalletAPI(cctx)
	})
	return err
}

func newWalletAPI(cctx *cli.Context) (api.WalletAPI, error) {
	repoPath := cctx.String(FlagWalletRepo)
	r, err := repo.NewFS(repoPath)
	if err != nil {
		return nil, err
	}

	ok, err := r.Exists()
	if err != nil {
		return nil, err
	}
	if !ok {
		if err := r.Init(repo.Wallet); err != nil {
			return nil, err
		}
	}

	lr, err := r.Lock(repo.Wallet)
	if err != nil {
		return nil, err
	}

	ks, err := lr.KeyStore()
	if err != nil {
		return nil, err
	}

	return wallet.NewWallet(ks)
}

var walletNew = &cli.Command{
	Name:      "new",
	Usage:     "Generate a new key of the given type",
	ArgsUsage: "[bls|secp256k1 (default secp256k1)]",
	Action: func(cctx *cli.Context) error {
		api, err := newWalletAPI(cctx)
		if err != nil {
			return err
		}

		t := cctx.Args().First()
		if t == "" {
			t = "secp256k1"
		}

		nk, err := api.WalletNew(cctx.Context, types.KeyType(t))
		if err != nil {
			return err
		}

		fmt.Println(nk.String())

		return nil
	},
}

var walletList = &cli.Command{
	Name:  "list",
	Usage: "List wallet address",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "addr-only",
			Usage:   "Only print addresses",
			Aliases: []string{"a"},
		},
		&cli.BoolFlag{
			Name:    "id",
			Usage:   "Output ID addresses",
			Aliases: []string{"i"},
		},
		&cli.BoolFlag{
			Name:    "market",
			Usage:   "Output market balances",
			Aliases: []string{"m"},
		},
	},
	Action: func(cctx *cli.Context) error {
		api, err := newWalletAPI(cctx)
		if err != nil {
			return err
		}

		addrs, err := api.WalletList(cctx.Context)
		if err != nil {
			return err
		}
		for _, addr := range addrs {
			fmt.Println(addr)
		}
		return nil
	},
}

var walletExport = &cli.Command{
	Name:      "export",
	Usage:     "export keys",
	ArgsUsage: "[address]",
	Action: func(cctx *cli.Context) error {
		api, err := newWalletAPI(cctx)
		if err != nil {
			return err
		}

		if !cctx.Args().Present() {
			return fmt.Errorf("must specify key to export")
		}

		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		ki, err := api.WalletExport(cctx.Context, addr)
		if err != nil {
			return err
		}

		b, err := json.Marshal(ki)
		if err != nil {
			return err
		}

		fmt.Println(hex.EncodeToString(b))
		return nil
	},
}

var walletImport = &cli.Command{
	Name:      "import",
	Usage:     "import keys",
	ArgsUsage: "[<path> (optional, will read from stdin if omitted)]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "format",
			Usage: "specify input format for key",
			Value: "hex-lotus",
		},
	},
	Action: func(cctx *cli.Context) error {
		api, err := newWalletAPI(cctx)
		if err != nil {
			return err
		}

		var inpdata []byte
		if !cctx.Args().Present() || cctx.Args().First() == "-" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter private key: ")
			indata, err := reader.ReadBytes('\n')
			if err != nil {
				return err
			}
			inpdata = indata

		} else {
			fdata, err := ioutil.ReadFile(cctx.Args().First())
			if err != nil {
				return err
			}
			inpdata = fdata
		}

		var ki types.KeyInfo
		switch cctx.String("format") {
		case "hex-lotus":
			data, err := hex.DecodeString(strings.TrimSpace(string(inpdata)))
			if err != nil {
				return err
			}

			if err := json.Unmarshal(data, &ki); err != nil {
				return err
			}
		case "json-lotus":
			if err := json.Unmarshal(inpdata, &ki); err != nil {
				return err
			}
		case "gfc-json":
			var f struct {
				KeyInfo []struct {
					PrivateKey []byte
					SigType    int
				}
			}
			if err := json.Unmarshal(inpdata, &f); err != nil {
				return xerrors.Errorf("failed to parse go-filecoin key: %s", err)
			}

			gk := f.KeyInfo[0]
			ki.PrivateKey = gk.PrivateKey
			switch gk.SigType {
			case 1:
				ki.Type = types.KTSecp256k1
			case 2:
				ki.Type = types.KTBLS
			default:
				return fmt.Errorf("unrecognized key type: %d", gk.SigType)
			}
		default:
			return fmt.Errorf("unrecognized format: %s", cctx.String("format"))
		}

		addr, err := api.WalletImport(cctx.Context, &ki)
		if err != nil {
			return err
		}

		fmt.Printf("imported key %s successfully!\n", addr)
		return nil
	},
}

var walletSign = &cli.Command{
	Name:      "sign",
	Usage:     "sign a message",
	ArgsUsage: "<signing address> <hexMessage>",
	Action: func(cctx *cli.Context) error {
		walletAPI, err := newWalletAPI(cctx)
		if err != nil {
			return err
		}

		if !cctx.Args().Present() || cctx.NArg() != 2 {
			return fmt.Errorf("must specify signing address and message to sign")
		}

		addr, err := address.NewFromString(cctx.Args().First())

		if err != nil {
			return err
		}

		msg, err := hex.DecodeString(cctx.Args().Get(1))

		if err != nil {
			return err
		}

		sig, err := walletAPI.WalletSign(cctx.Context, addr, msg, api.MsgMeta{
			Type: api.MTUnknown,
		})

		if err != nil {
			return err
		}

		sigBytes := append([]byte{byte(sig.Type)}, sig.Data...)

		fmt.Println(hex.EncodeToString(sigBytes))
		return nil
	},
}

var walletDelete = &cli.Command{
	Name:      "delete",
	Usage:     "Delete an account from the wallet",
	ArgsUsage: "<address> ",
	Action: func(cctx *cli.Context) error {
		api, err := newWalletAPI(cctx)
		if err != nil {
			return err
		}

		if !cctx.Args().Present() || cctx.NArg() != 1 {
			return fmt.Errorf("must specify address to delete")
		}

		addr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		return api.WalletDelete(cctx.Context, addr)
	},
}

var msigProposeCmd = &cli.Command{
	Name:      "propose",
	Usage:     "Propose a multisig transaction",
	ArgsUsage: "[multisigAddress destinationAddress value <methodId methodParams> (optional)]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "from",
			Usage: "account to send the propose message from",
		},
		&cli.StringFlag{
			Name:  "gas-premium",
			Usage: "specify gas price to use in AttoFIL",
			Value: "0",
		},
		&cli.StringFlag{
			Name:  "gas-feecap",
			Usage: "specify gas fee cap to use in AttoFIL",
			Value: "0",
		},
		&cli.Int64Flag{
			Name:  "gas-limit",
			Usage: "specify gas limit",
			Value: 0,
		},
		&cli.Int64Flag{
			Name:  "nonce",
			Usage: "specify the nonce to use",
			Value: 0,
		},
		&cli.Uint64Flag{
			Name:  "method",
			Usage: "specify method to invoke",
			Value: 0,
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := newWalletAPI(cctx)
		if err != nil {
			return err
		}

		if cctx.Args().Len() < 3 {
			return ShowHelp(cctx, fmt.Errorf("must pass at least multisig address, destination, and value"))
		}

		if cctx.Args().Len() > 3 && cctx.Args().Len() != 5 {
			return ShowHelp(cctx, fmt.Errorf("must either pass three or five arguments"))
		}

		if !cctx.IsSet("from") {
			return ShowHelp(cctx, xerrors.New("must specify --from"))
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		dest, err := address.NewFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		value, err := types.ParseFIL(cctx.Args().Get(2))
		if err != nil {
			return err
		}

		var method uint64
		var params []byte
		if cctx.Args().Len() == 5 {
			m, err := strconv.ParseUint(cctx.Args().Get(3), 10, 64)
			if err != nil {
				return err
			}
			method = m

			p, err := hex.DecodeString(cctx.Args().Get(4))
			if err != nil {
				return err
			}
			params = p
		}

		from, err := address.NewFromString(cctx.String("from"))
		if err != nil {
			return err
		}

		m := abi.MethodNum(method)

		msgBuilder := multisig0.Message(actors.Version0, from)

		msg, err := msgBuilder.Propose(msig, dest, types.BigInt(value), m, params)

		if err != nil {
			return xerrors.Errorf("buildProposeMessage: %w", err)
		}
		if err := setGasParamsFromCCtx(cctx, msg); err != nil {
			return err
		}
		msgStr, err := signMsg(cctx.Context, walletAPI, msg)
		if err != nil {
			return err
		}
		fmt.Println(msgStr)
		return nil
	},
}

func signMsg(ctx context.Context, walletAPI api.WalletAPI, msg *types.Message) (string, error) {
	mb, err := msg.ToStorageBlock()
	if err != nil {
		return "", xerrors.Errorf("serializing message: %w", err)
	}
	sig, err := walletAPI.WalletSign(ctx, msg.From, mb.Cid().Bytes(), api.MsgMeta{
		Type:  api.MTChainMsg,
		Extra: mb.RawData(),
	})
	if err != nil {
		return "", xerrors.Errorf("failed to sign message: %w", err)
	}
	sm := &types.SignedMessage{
		Message:   *msg,
		Signature: *sig,
	}
	smBytes, err := sm.Serialize()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(smBytes), nil
}

func parseProposeMessageFromContext(cctx *cli.Context) (*types.Message, error) {
	msig, err := address.NewFromString(cctx.Args().Get(0))
	if err != nil {
		return nil, err
	}

	dest, err := address.NewFromString(cctx.Args().Get(1))
	if err != nil {
		return nil, err
	}

	value, err := types.ParseFIL(cctx.Args().Get(2))
	if err != nil {
		return nil, err
	}

	var method uint64
	var params []byte
	if cctx.Args().Len() == 5 {
		m, err := strconv.ParseUint(cctx.Args().Get(3), 10, 64)
		if err != nil {
			return nil, err
		}
		method = m

		p, err := hex.DecodeString(cctx.Args().Get(4))
		if err != nil {
			return nil, err
		}
		params = p
	}

	from, err := address.NewFromString(cctx.String("from"))
	if err != nil {
		return nil, err
	}

	m := abi.MethodNum(method)

	enc, actErr := actors.SerializeParams(&multisig.ProposeParams{
		To:     dest,
		Value:  types.BigInt(value),
		Method: m,
		Params: params,
	})
	if actErr != nil {
		return nil, xerrors.Errorf("failed to serialize parameters: %w", actErr)
	}

	msg := &types.Message{
		From:   from,
		To:     msig,
		Value:  abi.NewTokenAmount(0),
		Method: builtin.MethodsMultisig.Propose,
		Params: enc,
	}
	if err := setGasParamsFromCCtx(cctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func setGasParamsFromCCtx(cctx *cli.Context, msg *types.Message) error {
	gp, err := types.BigFromString(cctx.String("gas-premium"))
	if err != nil {
		return fmt.Errorf("failed to parse gas-premium: %w", err)
	}
	gfc, err := types.BigFromString(cctx.String("gas-feecap"))
	if err != nil {
		return fmt.Errorf("failed to parse gas-feecap: %w", err)
	}
	msg.GasPremium = gp
	msg.GasFeeCap = gfc
	msg.GasLimit = cctx.Int64("gas-limit")
	msg.Nonce = uint64(cctx.Int64("nonce"))
	return nil
}

var msigApproveCmd = &cli.Command{
	Name:      "approve",
	Usage:     "Approve a multisig message",
	ArgsUsage: "<multisigAddress messageId> [proposerAddress destination value [methodId methodParams]]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "from",
			Usage: "account to send the approve message from",
		},
		&cli.StringFlag{
			Name:  "gas-premium",
			Usage: "specify gas price to use in AttoFIL",
			Value: "0",
		},
		&cli.StringFlag{
			Name:  "gas-feecap",
			Usage: "specify gas fee cap to use in AttoFIL",
			Value: "0",
		},
		&cli.Int64Flag{
			Name:  "gas-limit",
			Usage: "specify gas limit",
			Value: 0,
		},
		&cli.Int64Flag{
			Name:  "nonce",
			Usage: "specify the nonce to use",
			Value: 0,
		},
		&cli.Uint64Flag{
			Name:  "method",
			Usage: "specify method to invoke",
			Value: 0,
		},
	},
	Action: func(cctx *cli.Context) error {
		walletAPI, err := newWalletAPI(cctx)
		if err != nil {
			return err
		}
		if cctx.Args().Len() < 2 {
			return ShowHelp(cctx, fmt.Errorf("must pass at least multisig address and message ID"))
		}

		if cctx.Args().Len() > 5 && cctx.Args().Len() != 7 {
			return ShowHelp(cctx, fmt.Errorf("usage: msig approve <msig addr> <message ID> <proposer address> <desination> <value> [ <method> <params> ]"))
		}

		if cctx.Args().Len() > 2 && cctx.Args().Len() != 5 {
			return ShowHelp(cctx, fmt.Errorf("usage: msig approve <msig addr> <message ID> <proposer address> <desination> <value>"))
		}

		msig, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		txid, err := strconv.ParseUint(cctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}

		from, err := address.NewFromString(cctx.String("from"))
		if err != nil {
			return err
		}

		msgBuilder := multisig0.Message(actors.Version0, from)

		msg, err := msgBuilder.Approve(msig, txid, nil)
		if err != nil {
			return err
		}
		if err := setGasParamsFromCCtx(cctx, msg); err != nil {
			return err
		}
		msgStr, err := signMsg(cctx.Context, walletAPI, msg)
		if err != nil {
			return err
		}
		fmt.Println(msgStr)
		return nil
	},
}
