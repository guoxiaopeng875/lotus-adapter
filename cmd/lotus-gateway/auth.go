package main

import (
	"crypto/rand"
	"errors"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/filecoin-project/lotus/node/repo"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/guoxiaopeng875/lotus-adapter/api/apistruct"
	"golang.org/x/xerrors"
	"io"
	"io/ioutil"
)

const JWTSecretName = "auth-jwt-private"  //nolint:gosec
const KTJwtHmacSecret = "jwt-hmac-secret" //nolint:gosec

type JwtPayload struct {
	Allow []auth.Permission
}

func APISecret(keystore types.KeyStore, lr repo.LockedRepo) (*dtypes.APIAlg, error) {
	key, err := keystore.Get(JWTSecretName)

	if errors.Is(err, types.ErrKeyInfoNotFound) {
		log.Warn("Generating new API secret")

		sk, err := ioutil.ReadAll(io.LimitReader(rand.Reader, 32))
		if err != nil {
			return nil, err
		}

		key = types.KeyInfo{
			Type:       KTJwtHmacSecret,
			PrivateKey: sk,
		}

		if err := keystore.Put(JWTSecretName, key); err != nil {
			return nil, xerrors.Errorf("writing API secret: %w", err)
		}

		// TODO: make this configurable
		p := JwtPayload{
			Allow: apistruct.AllPermissions,
		}

		cliToken, err := jwt.Sign(&p, jwt.NewHS256(key.PrivateKey))
		if err != nil {
			return nil, err
		}

		if err := lr.SetAPIToken(cliToken); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, xerrors.Errorf("could not get JWT Token: %w", err)
	}

	return (*dtypes.APIAlg)(jwt.NewHS256(key.PrivateKey)), nil
}

func AuthVerify(token string, apiSecret *dtypes.APIAlg) ([]auth.Permission, error) {
	var payload JwtPayload
	if _, err := jwt.Verify([]byte(token), (*jwt.HMACSHA)(apiSecret), &payload); err != nil {
		return nil, xerrors.Errorf("JWT Verification failed: %w", err)
	}

	return payload.Allow, nil
}

func AuthNew(perms []auth.Permission, apiSecret *dtypes.APIAlg) ([]byte, error) {
	p := JwtPayload{
		Allow: perms, // TODO: consider checking validity
	}

	return jwt.Sign(&p, (*jwt.HMACSHA)(apiSecret))
}
