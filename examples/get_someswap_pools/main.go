package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"

	"github.com/KyberNetwork/ethrpc"
	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/entity"
	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/liquidity-source/someswap"
	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/source/pool"
	poollist "github.com/KyberNetwork/kyberswap-dex-lib/pkg/source/pool/list"
	pooltrack "github.com/KyberNetwork/kyberswap-dex-lib/pkg/source/pool/tracker"
	"github.com/ethereum/go-ethereum/common"
)

// Конфигурация SomeSwap на Monad
const (
	rpcURL           = "https://rpc-mainnet.monadinfra.com/rpc/elHriSqJvel4qXhCU462GZDRcymRpGlA"
	factoryAddress   = "0x00008A3c1077325Bb19cd93e5a0f1E95144700fa"
	multicallAddress = "0xcA11bde05977b3631167028862bE2a173976CA11"
)

type feeExtra struct {
	BaseFeeBps    string `json:"baseFeeBps"`
	DynamicFeeBps string `json:"dynamicFeeBps"`
	WToken0In     string `json:"wToken0In"`
	WToken1In     string `json:"wToken1In"`
}

type feeInfo struct {
	Extra    feeExtra
	SwapFee  float64
	HasExtra bool
}

func main() {
	fmt.Printf("🔗 Подключаемся к Monad RPC\n")
	ethrpcClient := ethrpc.New(rpcURL)
	ethrpcClient.SetMulticallContract(common.HexToAddress(multicallAddress))
	fmt.Printf("📞 Multicall контракт: %s\n", multicallAddress)

	ctx := context.Background()

	// 1) Конфигурация SomeSwap
	config := &someswap.Config{
		DexID:          someswap.DexType,
		FactoryAddress: factoryAddress,
		NewPoolLimit:   10,
	}
	fmt.Printf("🏭 Factory адрес: %s\n", factoryAddress)

	// 2) PoolsListUpdater
	factoryParams := poollist.FactoryParams{
		Exchange: someswap.DexType,
		Properties: map[string]any{
			"dexID":          config.DexID,
			"factoryAddress": config.FactoryAddress,
			"newPoolLimit":   config.NewPoolLimit,
		},
		Dependencies: poollist.Dependencies{
			EthrpcClient: ethrpcClient,
		},
	}

	poolsListerFactory := poollist.Factory(someswap.DexType)
	if poolsListerFactory == nil {
		fmt.Println("❌ Ошибка: не найдена фабрика для someswap")
		return
	}

	poolsLister, err := poolsListerFactory(someswap.DexType, factoryParams)
	if err != nil {
		fmt.Printf("❌ Ошибка создания PoolsListUpdater: %v\n", err)
		return
	}

	// 3) Получаем список пулов (все)
	fmt.Println("\n📋 Получаем список пулов...")
	pools := make([]entity.Pool, 0, 1024)
	var scanMetadata []byte
	for {
		batch, newMetadata, err := poolsLister.GetNewPools(ctx, scanMetadata)
		if err != nil {
			fmt.Printf("❌ Ошибка получения пулов: %v\n", err)
			return
		}
		if len(batch) == 0 {
			break
		}
		pools = append(pools, batch...)
		scanMetadata = newMetadata
	}

	fmt.Printf("✅ Всего пулов: %d\n", len(pools))
	for i, poolEntity := range pools {
		fmt.Printf("%d) %s | %s / %s\n", i+1, poolEntity.Address, poolEntity.Tokens[0].Address, poolEntity.Tokens[1].Address)
	}
	if len(pools) == 0 {
		fmt.Println("⚠️  Пулы не найдены. Проверьте адрес Factory или RPC.")
		return
	}

	// 4) Собираем уникальные fee-конфиги по всем парам
	fmt.Println("\n🧾 Собираем уникальные fee...")
	uniqueFees := make(map[string]feeInfo)
	for _, p := range pools {
		if p.StaticExtra != "" {
			var extra feeExtra
			if err := json.Unmarshal([]byte(p.StaticExtra), &extra); err != nil {
				fmt.Printf("⚠️  Ошибка разбора fee: %v\n", err)
				continue
			}
			if _, ok := uniqueFees[p.StaticExtra]; !ok {
				uniqueFees[p.StaticExtra] = feeInfo{
					Extra:    extra,
					SwapFee:  p.SwapFee,
					HasExtra: true,
				}
			}
			continue
		}
		key := fmt.Sprintf("swapFee:%f", p.SwapFee)
		if _, ok := uniqueFees[key]; !ok {
			uniqueFees[key] = feeInfo{
				SwapFee: p.SwapFee,
			}
		}
	}

	keys := make([]string, 0, len(uniqueFees))
	for k := range uniqueFees {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("✅ Уникальных fee-конфигов: %d\n", len(uniqueFees))
	for i, key := range keys {
		info := uniqueFees[key]
		if info.HasExtra {
			fmt.Printf("   %d) base=%s, dynamic=%s, w0=%s, w1=%s, swapFee=%.6f\n",
				i+1, info.Extra.BaseFeeBps, info.Extra.DynamicFeeBps, info.Extra.WToken0In, info.Extra.WToken1In, info.SwapFee)
			continue
		}
		fmt.Printf("   %d) swapFee=%.6f\n", i+1, info.SwapFee)
	}

	// 5) PoolTracker
	trackerFactory := pooltrack.Factory(someswap.DexType)
	if trackerFactory == nil {
		fmt.Println("❌ Ошибка: не найдена фабрика трекера для someswap")
		return
	}

	trackerParams := pooltrack.FactoryParams{
		Exchange: someswap.DexType,
		Dependencies: pooltrack.Dependencies{
			EthrpcClient: ethrpcClient,
		},
	}

	poolTracker, err := trackerFactory(someswap.DexType, trackerParams)
	if err != nil {
		fmt.Printf("❌ Ошибка создания PoolTracker: %v\n", err)
		return
	}

	fmt.Printf("\n📊 Информация о всех пулах:\n")
	fmt.Println("=" + string(make([]byte, 100)) + "=")

	for i := 0; i < len(pools); i++ {
		poolEntity := pools[i]
		fmt.Printf("\n🏊 Пул #%d:\n", i+1)
		fmt.Printf("   Адрес: %s\n", poolEntity.Address)
		fmt.Printf("   Токены: %s / %s\n", poolEntity.Tokens[0].Address, poolEntity.Tokens[1].Address)

		updatedPool, err := poolTracker.GetNewPoolState(ctx, poolEntity, pool.GetNewPoolStateParams{
			Logs: nil,
		})
		if err != nil {
			fmt.Printf("   ⚠️  Ошибка получения резервов: %v\n", err)
			continue
		}

		fmt.Printf("   Резервы: %s / %s\n", updatedPool.Reserves[0], updatedPool.Reserves[1])
		printFeeConfig(updatedPool)

		poolSim, err := someswap.NewPoolSimulator(updatedPool)
		if err != nil {
			fmt.Printf("   ⚠️  Ошибка создания симулятора: %v\n", err)
			continue
		}

		token0Amount := big.NewInt(1)
		token0Amount.Mul(token0Amount, big.NewInt(1e18))

		amountIn := pool.TokenAmount{
			Token:  updatedPool.Tokens[0].Address,
			Amount: token0Amount,
		}

		result, err := poolSim.CalcAmountOut(pool.CalcAmountOutParams{
			TokenAmountIn: amountIn,
			TokenOut:      updatedPool.Tokens[1].Address,
		})
		if err != nil {
			fmt.Printf("   ⚠️  Ошибка калькуляции: %v\n", err)
			continue
		}

		amountOut := new(big.Float).SetInt(result.TokenAmountOut.Amount)
		amountOut.Quo(amountOut, big.NewFloat(1e18))

		fmt.Printf("   💱 Калькуляция свапа:\n")
		fmt.Printf("      Вход: 1.0 %s\n", shortToken(updatedPool.Tokens[0].Address))
		fmt.Printf("      Выход: %s %s\n", amountOut.Text('f', 6), shortToken(updatedPool.Tokens[1].Address))
		fmt.Printf("      Газ: %d\n", result.Gas)
	}

	fmt.Println("\n" + string(make([]byte, 100)) + "=")
	fmt.Println("✅ Готово!")
}

func shortToken(address string) string {
	if len(address) > 10 {
		return address[:10] + "..."
	}
	return address
}

func printFeeConfig(p entity.Pool) {
	if p.StaticExtra == "" {
		fmt.Printf("   FeeConfig: swapFee=%.6f\n", p.SwapFee)
		return
	}
	var extra feeExtra
	if err := json.Unmarshal([]byte(p.StaticExtra), &extra); err != nil {
		fmt.Printf("   FeeConfig: <invalid staticExtra> swapFee=%.6f\n", p.SwapFee)
		return
	}
	fmt.Printf("   FeeConfig: base=%s, dynamic=%s, w0=%s, w1=%s, swapFee=%.6f\n",
		extra.BaseFeeBps, extra.DynamicFeeBps, extra.WToken0In, extra.WToken1In, p.SwapFee)
}
