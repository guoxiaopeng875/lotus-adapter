package main

import (
	"fmt"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/lotus/chain/wallet"
	"github.com/filecoin-project/lotus/node/repo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuth(t *testing.T) {
	lr := repo.NewMemory(nil)
	r, err := lr.Lock(repo.FullNode)
	require.NoError(t, err)
	secret, err := APISecret(wallet.NewMemKeyStore(), r)
	require.NoError(t, err)
	data, err := AuthNew([]auth.Permission{
		"admin",
	}, secret)
	require.NoError(t, err)
	token := string(data)
	fmt.Println(token)
	pms, err := AuthVerify(token, secret)
	require.NoError(t, err)
	fmt.Println(pms)
}
