package output

import (
	"context"
	"crypto/ecdsa"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/machinefi/sprout/task"
)

var (
	testMethodName = "testMethod"
	//go:embed testdata/testABI_proofInputOnlyMethod.json
	testABIProofInputOnlyMethod string
	//go:embed testdata/testABI_projectInputOnlyMethod.json
	testABIProjectInputOnlyMethod string
	//go:embed testdata/testABI_dataSnarkInputOnlyMethod.json
	testABIDataSnarkInputOnlyMethod string
	//go:embed testdata/testABI_receiverInputOnlyMethod.json
	testABIReceiverInputOnlyMethod string
	//go:embed testdata/testABI_otherInputOnlyMethod.json
	testABIOtherInputOnlyMethod string
	//go:embed testdata/testABI_otherInputOnlyMethod_address.json
	testABIOtherInputOnlyMethod_address string
	//go:embed testdata/testABI_otherInputOnlyMethod_uint256.json
	testABIOtherInputOnlyMethod_uint256 string

	conf = &Config{
		Type: EthereumContract,
		Ethereum: EthereumConfig{
			ContractMethod: testMethodName,
		},
	}
)

func patchEthereumContractSendTX(p *Patches, txhash string, err error) *Patches {
	return p.ApplyPrivateMethod(&ethereumContract{}, "sendTX",
		func(contract *ethereumContract, ctx context.Context, data []byte) (string, error) {
			return txhash, err
		},
	)
}

func Test_ethereumContract_Output(t *testing.T) {
	r := require.New(t)
	txHashRet := "anyTxHash"

	t.Run("BuildParameters", func(t *testing.T) {
		t.Run("Proof", func(t *testing.T) {
			p := NewPatches()
			defer p.Reset()

			patchEthereumContractSendTX(p, txHashRet, nil)
			p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
			p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", nil, nil)

			conf.Ethereum.ContractAbiJSON = testABIProofInputOnlyMethod
			o, err := New(conf, "1", "", "")
			r.NoError(err)
			_, ok := o.(*ethereumContract)
			r.True(ok)

			txHash, err := o.Output(uint64(0), &task.Task{}, []byte("any proof data"))
			r.Equal(txHash, txHashRet)
			r.NoError(err)
		})

		t.Run("ProjectID", func(t *testing.T) {
			p := NewPatches()
			defer p.Reset()

			patchEthereumContractSendTX(p, txHashRet, nil)
			p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
			p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", nil, nil)

			conf.Ethereum.ContractAbiJSON = testABIProjectInputOnlyMethod
			o, err := New(conf, "1", "", "")
			r.NoError(err)
			_, ok := o.(*ethereumContract)
			r.True(ok)

			txHash, err := o.Output(uint64(0), &task.Task{}, []byte("any proof data"))
			r.Equal(txHash, txHashRet)
			r.NoError(err)
		})

		t.Run("Receiver", func(t *testing.T) {
			p := NewPatches()
			defer p.Reset()

			patchEthereumContractSendTX(p, txHashRet, nil)
			p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
			p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", nil, nil)

			conf.Ethereum.ContractAbiJSON = testABIReceiverInputOnlyMethod
			o, err := New(conf, "1", "", "")
			r.NoError(err)
			_, ok := o.(*ethereumContract)
			r.True(ok)

			t.Run("ReceiverAddressNotInConfig", func(t *testing.T) {
				txHash, err := o.Output(uint64(0), &task.Task{}, []byte("any proof data"))
				r.Equal(txHash, "")
				r.Equal(err, errMissingReceiverParam)
			})

			conf.Ethereum.ReceiverAddress = "0x"
			o, err = New(conf, "1", "", "")
			r.NoError(err)
			_, ok = o.(*ethereumContract)
			r.True(ok)

			t.Run("Success", func(t *testing.T) {
				txHash, err := o.Output(uint64(0), &task.Task{}, []byte("any proof data"))
				r.Equal(txHash, txHashRet)
				r.NoError(err)
			})
		})

		t.Run("DataSnark", func(t *testing.T) {
			p := NewPatches()
			defer p.Reset()

			patchEthereumContractSendTX(p, txHashRet, nil)
			p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
			p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", nil, nil)

			conf.Ethereum.ContractAbiJSON = testABIDataSnarkInputOnlyMethod
			o, err := New(conf, "1", "", "")
			r.NoError(err)
			_, ok := o.(*ethereumContract)
			r.True(ok)

			t.Run("FailedToDecodeProof", func(t *testing.T) {
				txHash, err := o.Output(uint64(0), &task.Task{}, []byte("INVALID_HEX_DECODE"))
				r.Equal(txHash, "")
				r.Error(err)
			})
			proof := &struct {
				Snark map[string]string `json:"Snark"`
			}{
				Snark: make(map[string]string),
			}
			t.Run("MissingSnarkField", func(t *testing.T) {
				data, err := json.Marshal(proof)
				r.NoError(err)
				hexdata := hex.EncodeToString(data)

				txHash, err := o.Output(uint64(0), &task.Task{}, []byte(hexdata))
				r.Equal(txHash, "")
				r.Equal(err, errSnarkProofDataMissingFieldSnark)
			})
		})
		t.Run("Default", func(t *testing.T) {
			p := NewPatches()
			defer p.Reset()

			patchEthereumContractSendTX(p, txHashRet, nil)
			p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
			p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", nil, nil)

			t.Run("MissingMethodNameParam", func(t *testing.T) {
				conf.Ethereum.ContractAbiJSON = testABIOtherInputOnlyMethod
				o, err := New(conf, "1", "", "")
				r.NoError(err)
				_, ok := o.(*ethereumContract)
				r.True(ok)

				txHash, err := o.Output(uint64(0), &task.Task{
					ProjectID: 1,
					Data:      [][]byte{[]byte(`{"other":""}`)},
				}, nil)
				r.Equal(txHash, "")
				r.Error(err)
			})

			t.Run("BuildParamsByType", func(t *testing.T) {
				t.Run("Address", func(t *testing.T) {
					conf.Ethereum.ContractAbiJSON = testABIOtherInputOnlyMethod_address
					o, err := New(conf, "1", "", "")
					r.NoError(err)
					_, ok := o.(*ethereumContract)
					r.True(ok)

					txHash, err := o.Output(uint64(0), &task.Task{
						ProjectID: 1,
						Data:      [][]byte{[]byte(`{"other":"any"}`)},
					}, nil)
					r.Equal(txHash, txHashRet)
					r.NoError(err)
				})
				t.Run("Uint256", func(t *testing.T) {
					conf.Ethereum.ContractAbiJSON = testABIOtherInputOnlyMethod_uint256
					o, err := New(conf, "1", "", "")
					r.NoError(err)
					_, ok := o.(*ethereumContract)
					r.True(ok)

					txHash, err := o.Output(uint64(0), &task.Task{
						ProjectID: 1,
						Data:      [][]byte{[]byte(`{"other":"any"}`)},
					}, nil)
					r.Equal(txHash, txHashRet)
					r.NoError(err)
				})
				t.Run("Other", func(t *testing.T) {
					conf.Ethereum.ContractAbiJSON = testABIOtherInputOnlyMethod
					o, err := New(conf, "1", "", "")
					r.NoError(err)
					_, ok := o.(*ethereumContract)
					r.True(ok)

					txHash, err := o.Output(uint64(0), &task.Task{
						ProjectID: 1,
						Data:      [][]byte{[]byte(`{"other":"any"}`)},
					}, nil)
					r.NoError(err)
					r.Equal(txHash, txHashRet)
				})
			})
		})
	})
	t.Run("PackTxData", func(t *testing.T) {
		t.Run("FailedToPack", func(t *testing.T) {
			p := NewPatches()
			defer p.Reset()

			patchEthereumContractSendTX(p, txHashRet, nil)
			p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
			p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", nil, nil)
			p.ApplyMethodReturn(abi.ABI{}, "Pack", nil, errors.New(t.Name()))

			conf.Ethereum.ContractAbiJSON = testABIProofInputOnlyMethod
			o, err := New(conf, "1", "", "")
			r.NoError(err)

			txHash, err := o.Output(uint64(0), &task.Task{
				ProjectID: 1,
				Data:      [][]byte{[]byte(`{"other":""}`)},
			}, nil)

			r.Equal(txHash, "")
			r.ErrorContains(err, t.Name())
		})
	})

	t.Run("SendTxData", func(t *testing.T) {
		t.Run("FailedToSendTx", func(t *testing.T) {
			p := NewPatches()
			defer p.Reset()

			patchEthereumContractSendTX(p, "", errors.New(t.Name()))
			p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
			p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", nil, nil)
			p.ApplyMethodReturn(abi.ABI{}, "Pack", nil, errors.New(t.Name()))

			o, err := New(conf, "1", "", "")
			r.NoError(err)

			txHash, err := o.Output(uint64(0), &task.Task{
				ProjectID: 1,
				Data:      [][]byte{[]byte(`{"other":""}`)},
			}, nil)

			r.Equal(txHash, "")
			r.ErrorContains(err, t.Name())
		})
	})

	t.Run("Success", func(t *testing.T) {
		p := NewPatches()
		defer p.Reset()

		patchEthereumContractSendTX(p, txHashRet, nil)
		p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", nil, nil)
		p.ApplyMethodReturn(abi.ABI{}, "Pack", nil, nil)

		o, err := New(conf, "1", "", "")
		r.NoError(err)

		txHash, err := o.Output(uint64(0), &task.Task{
			ProjectID: 1,
			Data:      [][]byte{[]byte(`{"other":""}`)},
		}, nil)
		r.NoError(err)
		r.Equal(txHash, txHashRet)
	})
}

func Test_ethereumContract_sendTX(t *testing.T) {
	r := require.New(t)

	conf.Ethereum.ContractAbiJSON = testABIOtherInputOnlyMethod
	ctx := context.Background()

	t.Run("SuggestGasFailed", func(t *testing.T) {
		p := NewPatches()
		defer p.Reset()

		p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", nil, nil)
		p.ApplyFuncReturn(crypto.ToECDSAUnsafe, &ecdsa.PrivateKey{})
		p.ApplyFuncReturn(crypto.PubkeyToAddress, common.Address{})
		p.ApplyFuncReturn(common.HexToAddress, common.Address{})
		p.ApplyMethodReturn(&ethclient.Client{}, "SuggestGasPrice", nil, errors.New(t.Name()))

		o, err := New(conf, "1", "", "")
		r.NoError(err)
		contract, ok := o.(*ethereumContract)
		r.True(ok)
		_, err = contract.sendTX(ctx, nil)
		r.ErrorContains(err, t.Name())
	})

	t.Run("GetNonceFailed", func(t *testing.T) {
		p := NewPatches()
		defer p.Reset()

		p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
		p.ApplyFuncReturn(crypto.ToECDSAUnsafe, &ecdsa.PrivateKey{})
		p.ApplyFuncReturn(crypto.PubkeyToAddress, common.Address{})
		p.ApplyFuncReturn(common.HexToAddress, common.Address{})
		p.ApplyMethodReturn(&ethclient.Client{}, "SuggestGasPrice", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "EstimateGas", uint64(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "PendingNonceAt", nil, errors.New(t.Name()))

		o, err := New(conf, "1", "", "")
		r.NoError(err)
		contract, ok := o.(*ethereumContract)
		r.True(ok)
		_, err = contract.sendTX(ctx, nil)
		r.ErrorContains(err, t.Name())
	})

	t.Run("EstimateGasFailed", func(t *testing.T) {
		p := NewPatches()
		defer p.Reset()

		p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
		p.ApplyFuncReturn(crypto.ToECDSAUnsafe, &ecdsa.PrivateKey{})
		p.ApplyFuncReturn(crypto.PubkeyToAddress, common.Address{})
		p.ApplyFuncReturn(common.HexToAddress, common.Address{})
		p.ApplyMethodReturn(&ethclient.Client{}, "SuggestGasPrice", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "PendingNonceAt", uint64(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "EstimateGas", nil, errors.New(t.Name()))

		o, err := New(conf, "1", "", "")
		r.NoError(err)
		contract, ok := o.(*ethereumContract)
		r.True(ok)
		_, err = contract.sendTX(ctx, nil)
		r.ErrorContains(err, t.Name())
	})

	t.Run("SignTxFailed", func(t *testing.T) {
		p := NewPatches()
		defer p.Reset()

		p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
		p.ApplyFuncReturn(crypto.ToECDSAUnsafe, &ecdsa.PrivateKey{})
		p.ApplyFuncReturn(crypto.PubkeyToAddress, common.Address{})
		p.ApplyFuncReturn(common.HexToAddress, common.Address{})
		p.ApplyMethodReturn(&ethclient.Client{}, "SuggestGasPrice", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "PendingNonceAt", uint64(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "EstimateGas", uint64(1), nil)
		p.ApplyFuncReturn(ethtypes.SignTx, nil, errors.New(t.Name()))

		o, err := New(conf, "1", "", "")
		r.NoError(err)
		contract, ok := o.(*ethereumContract)
		r.True(ok)
		_, err = contract.sendTX(ctx, nil)
		r.ErrorContains(err, t.Name())
	})

	t.Run("TransactionFailed", func(t *testing.T) {
		p := NewPatches()
		defer p.Reset()

		p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
		p.ApplyFuncReturn(crypto.ToECDSAUnsafe, &ecdsa.PrivateKey{})
		p.ApplyFuncReturn(crypto.PubkeyToAddress, common.Address{})
		p.ApplyFuncReturn(common.HexToAddress, common.Address{})
		p.ApplyMethodReturn(&ethclient.Client{}, "SuggestGasPrice", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "PendingNonceAt", uint64(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "EstimateGas", uint64(1), nil)
		p.ApplyFuncReturn(ethtypes.SignTx, nil, nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "SendTransaction", errors.New(t.Name()))

		o, err := New(conf, "1", "", "")
		r.NoError(err)
		contract, ok := o.(*ethereumContract)
		r.True(ok)
		_, err = contract.sendTX(ctx, nil)
		r.ErrorContains(err, t.Name())
	})

	t.Run("TransactionSuccess", func(t *testing.T) {
		p := NewPatches()
		defer p.Reset()

		p.ApplyFuncReturn(ethclient.Dial, &ethclient.Client{}, nil)
		p.ApplyFuncReturn(crypto.ToECDSAUnsafe, &ecdsa.PrivateKey{})
		p.ApplyFuncReturn(crypto.PubkeyToAddress, common.Address{})
		p.ApplyFuncReturn(common.HexToAddress, common.Address{})
		p.ApplyMethodReturn(&ethclient.Client{}, "SuggestGasPrice", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "ChainID", big.NewInt(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "PendingNonceAt", uint64(1), nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "EstimateGas", uint64(1), nil)
		p.ApplyFuncReturn(ethtypes.SignTx, nil, nil)
		p.ApplyMethodReturn(&ethclient.Client{}, "SendTransaction", nil)
		p.ApplyMethodReturn(&ethtypes.Transaction{}, "Hash", common.Hash{})

		o, err := New(conf, "1", "", "")
		r.NoError(err)
		contract, ok := o.(*ethereumContract)
		r.True(ok)
		tx, err := contract.sendTX(ctx, nil)
		r.NoError(err)
		r.Equal(tx, "0x0000000000000000000000000000000000000000000000000000000000000000")
	})
}
