package output

import (
	"context"
	"crypto/ed25519"
	"log/slog"

	"github.com/blocto/solana-go-sdk/client"
	solcommon "github.com/blocto/solana-go-sdk/common"
	soltypes "github.com/blocto/solana-go-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/machinefi/sprout/task"
)

type solanaProgram struct {
	endpoint       string
	programID      string
	secretKey      string
	stateAccountPK string
}

func (e *solanaProgram) Output(proverID uint64, task *task.Task, proof []byte) (string, error) {
	slog.Debug("outputing to solana program", "chain endpoint", e.endpoint)
	ins := e.packInstructions(proof)
	txHash, err := e.sendTX(ins)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func (e *solanaProgram) sendTX(ins []soltypes.Instruction) (string, error) {
	cli := client.NewClient(e.endpoint)
	b := common.FromHex(e.secretKey)
	pk := ed25519.PrivateKey(b)
	account := soltypes.Account{
		PublicKey:  solcommon.PublicKeyFromBytes(pk.Public().(ed25519.PublicKey)),
		PrivateKey: pk,
	}
	if len(ins) == 0 {
		return "", errors.New("missing instruction data")
	}

	resp, err := cli.GetLatestBlockhash(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "failed to get solana latest block hash")
	}
	tx, err := soltypes.NewTransaction(soltypes.NewTransactionParam{
		Message: soltypes.NewMessage(soltypes.NewMessageParam{
			FeePayer:        account.PublicKey,
			RecentBlockhash: resp.Blockhash,
			Instructions:    ins,
		}),
		Signers: []soltypes.Account{account},
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to build solana raw tx")
	}

	hash, err := cli.SendTransaction(context.Background(), tx)
	if err != nil {
		return "", errors.Wrap(err, "failed to send solana tx")
	}
	return hash, nil
}

// encodeData encodes the proof into the data field of the instruction
// the first byte is the instruction, which is 0 for now;
// the rest is the proof data.
// e.g. assume proof is [0x01, 0x02, 0x03], then the encoded data is [0x00, 0x01, 0x02, 0x03]
func (e *solanaProgram) encodeData(proof []byte) []byte {
	data := []byte{}
	data = append(data, byte(0)) // 0 means submit proof
	data = append(data, proof...)
	return data
}

func (e *solanaProgram) packInstructions(proof []byte) []soltypes.Instruction {
	accounts := []soltypes.AccountMeta{}
	if e.stateAccountPK != "" {
		accounts = append(accounts, soltypes.AccountMeta{
			PubKey:     solcommon.PublicKeyFromString(e.stateAccountPK),
			IsSigner:   false,
			IsWritable: true,
		})
	}

	return []soltypes.Instruction{
		{
			ProgramID: solcommon.PublicKeyFromString(e.programID),
			Accounts:  accounts,
			Data:      e.encodeData(proof),
		},
	}
}

func newSolanaProgram(conf SolanaConfig, secretKey string) (*solanaProgram, error) {
	if secretKey == "" {
		return nil, errors.New("secret key is empty")
	}
	return &solanaProgram{
		endpoint:       conf.ChainEndpoint,
		programID:      conf.ProgramID,
		secretKey:      secretKey,
		stateAccountPK: conf.StateAccountPK,
	}, nil
}
