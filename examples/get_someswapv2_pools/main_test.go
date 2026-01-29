package main

import (
	"context"
	"math/big"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/liquidity-source/someswapv2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

// Hardcoded config (as requested).
const (
	rpcURL = "https://rpc-mainnet.monadinfra.com/rpc/elHriSqJvel4qXhCU462GZDRcymRpGlA"

	factoryAddress = "0xF4B30295EA24938d9705E30F88e144140422BAa3"
)

func TestGetSomeSwapV2PairsOnMonad(t *testing.T) {
	// This is an integration test (it uses a real RPC endpoint).
	// Run it with: go test -v ./examples/get_someswapv2_pools -run TestGetSomeSwapV2PairsOnMonad

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)

	client, err := ethclient.DialContext(ctx, rpcURL)
	require.NoError(t, err)

	factory := common.HexToAddress(factoryAddress)
	abi := someswapv2.FactoryABI()

	// 1) Call allPairsLength() and pairCount(), compare them.
	var allPairsLength *big.Int
	var pairCount *big.Int
	{
		data, err := abi.Pack("allPairsLength")
		require.NoError(t, err)

		out, err := client.CallContract(ctx, ethereum.CallMsg{
			To:   &factory,
			Data: data,
		}, nil)
		require.NoError(t, err)

		vals, err := abi.Unpack("allPairsLength", out)
		require.NoError(t, err)
		require.Len(t, vals, 1)

		n, ok := vals[0].(*big.Int)
		require.True(t, ok)
		allPairsLength = n
	}

	{
		data, err := abi.Pack("pairCount")
		require.NoError(t, err)

		out, err := client.CallContract(ctx, ethereum.CallMsg{
			To:   &factory,
			Data: data,
		}, nil)
		require.NoError(t, err)

		vals, err := abi.Unpack("pairCount", out)
		require.NoError(t, err)
		require.Len(t, vals, 1)

		n, ok := vals[0].(*big.Int)
		require.True(t, ok)
		pairCount = n
	}

	t.Logf("factory=%s allPairsLength=%s pairCount=%s", factory.Hex(), allPairsLength.String(), pairCount.String())
	require.Equal(t, 0, allPairsLength.Cmp(pairCount), "pairCount() should match allPairsLength()")

	// 2) Fetch recent PairCreated logs and decode them
	event := abi.Events["PairCreated"]
	t.Logf("PairCreated topic0=%s", event.ID.Hex())

	latest, err := client.BlockNumber(ctx)
	require.NoError(t, err)

	// Reasonable default for a test; adjust locally if needed.
	// Note: Monad RPC may limit eth_getLogs ranges, so keep this modest.
	const lookbackBlocks = uint64(20_000)
	var from uint64
	if latest > lookbackBlocks {
		from = latest - lookbackBlocks
	}

	// Monad RPC may reject large block ranges ("block range too large"), so query logs in chunks.
	// Also: stop once we have enough logs to print.
	const (
		initialChunkSize = uint64(1_000) // inclusive range ~= chunkSize+1 blocks
		printLimit = 25
	)
	logs := make([]types.Log, 0, printLimit)

	// Scan backwards to get most recent events quickly.
	for end := latest; end >= from; {
		chunkSize := initialChunkSize

		var (
			start uint64
			batch []types.Log
		)
		for {
			start = from
			if end > chunkSize {
				start = end - chunkSize
			}
			if start < from {
				start = from
			}

			q := ethereum.FilterQuery{
				FromBlock: big.NewInt(int64(start)),
				ToBlock:   big.NewInt(int64(end)),
				Addresses: []common.Address{factory},
				Topics:    [][]common.Hash{{event.ID}},
			}

			batch, err = client.FilterLogs(ctx, q)
			if err == nil {
				break
			}

			// Auto-shrink range on RPC limitation.
			if strings.Contains(err.Error(), "block range too large") && chunkSize > 1 {
				chunkSize /= 2
				continue
			}

			require.NoError(t, err)
		}

		logs = append(logs, batch...)

		if len(logs) >= printLimit || start == from {
			break
		}

		// next window (end is exclusive of previous start)
		if start == 0 {
			break
		}
		end = start - 1
	}

	// Ensure deterministic order (oldest -> newest), like a single FilterLogs would return.
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].BlockNumber != logs[j].BlockNumber {
			return logs[i].BlockNumber < logs[j].BlockNumber
		}
		return logs[i].Index < logs[j].Index
	})
	t.Logf("PairCreated logs: count=%d blocks=[%d..%d]", len(logs), from, latest)

	type baseFeeConfig struct {
		BaseFee uint32 `abi:"baseFee"`
		WToken0 uint32 `abi:"wToken0"`
		WToken1 uint32 `abi:"wToken1"`
	}
	type pairCreatedNonIndexed struct {
		Pair          common.Address `abi:"pair"`
		PairCount     *big.Int       `abi:"pairCount"`
		ModuleMask    uint8          `abi:"moduleMask"`
		BaseFeeConfig baseFeeConfig  `abi:"baseFeeConfig"`
	}

	decodedPairs := make([]common.Address, 0, len(logs))

	// Keep output readable.
	for i := 0; i < len(logs) && i < printLimit; i++ {
		lg := logs[i]
		require.GreaterOrEqual(t, len(lg.Topics), 4, "PairCreated should have 4 topics (topic0 + 3 indexed addresses)")

		token0 := common.BytesToAddress(lg.Topics[1].Bytes())
		token1 := common.BytesToAddress(lg.Topics[2].Bytes())
		module := common.BytesToAddress(lg.Topics[3].Bytes())

		var decoded pairCreatedNonIndexed
		require.NoError(t, abi.UnpackIntoInterface(&decoded, "PairCreated", lg.Data))

		decodedPairs = append(decodedPairs, decoded.Pair)

		t.Logf(
			"[%d] block=%d tx=%s pair=%s token0=%s token1=%s module=%s pairCount=%s moduleMask=%d baseFee={baseFee=%d w0=%d w1=%d} data=%s",
			i,
			lg.BlockNumber,
			lg.TxHash.Hex(),
			decoded.Pair.Hex(),
			token0.Hex(),
			token1.Hex(),
			module.Hex(),
			func() string {
				if decoded.PairCount == nil {
					return "<nil>"
				}
				return decoded.PairCount.String()
			}(),
			decoded.ModuleMask,
			decoded.BaseFeeConfig.BaseFee,
			decoded.BaseFeeConfig.WToken0,
			decoded.BaseFeeConfig.WToken1,
			hexutil.Encode(lg.Data),
		)
	}

	// Print a sorted unique list of decoded pair addresses.
	if len(decodedPairs) > 0 {
		uniq := map[common.Address]struct{}{}
		for _, p := range decodedPairs {
			uniq[p] = struct{}{}
		}
		uniquePairs := make([]string, 0, len(uniq))
		for p := range uniq {
			uniquePairs = append(uniquePairs, p.Hex())
		}
		sort.Strings(uniquePairs)
		t.Logf("unique pairs in printed sample (%d):", len(uniquePairs))
		for _, p := range uniquePairs {
			t.Logf(" - %s", p)
		}
	}
}

