package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/KyberNetwork/ethrpc"
	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/source/pool"
	poollist "github.com/KyberNetwork/kyberswap-dex-lib/pkg/source/pool/list"
	pooltrack "github.com/KyberNetwork/kyberswap-dex-lib/pkg/source/pool/tracker"
	"github.com/KyberNetwork/kyberswap-dex-lib/pkg/source/uniswap"
	"github.com/ethereum/go-ethereum/common"
)

// Конфигурация для Monad сети
const (
	// RPC URL для Monad mainnet
	rpcURL = "https://rpc-mainnet.monadinfra.com/rpc/elHriSqJvel4qXhCU462GZDRcymRpGlA"
	
	// Адрес Uniswap V2 Factory контракта на Monad
	factoryAddress = "0x182a927119d56008d921126764bf884221b10f59"
	
	// Адрес Multicall контракта (стандартный для EVM сетей)
	multicallAddress = "0xcA11bde05977b3631167028862bE2a173976CA11"
)

func main() {
	// 1. Настройка RPC клиента для Monad
	fmt.Printf("🔗 Подключаемся к Monad RPC\n")
	ethrpcClient := ethrpc.New(rpcURL)
	ethrpcClient.SetMulticallContract(common.HexToAddress(multicallAddress))
	fmt.Printf("📞 Multicall контракт: %s\n", multicallAddress)

	ctx := context.Background()

	// 2. Создаем конфигурацию для Uniswap V2 на Monad

	config := &uniswap.Config{
		DexID:          "uniswap",
		SwapFee:        0.003, // 0.3% комиссия
		FactoryAddress: factoryAddress,
		NewPoolLimit:   10,    // Получаем первые 10 пулов для примера
	}
	
	fmt.Printf("🏭 Factory адрес: %s\n", factoryAddress)

	// 3. Создаем PoolsListUpdater через фабрику
	factoryParams := poollist.FactoryParams{
		Exchange: "uniswap",
		Properties: map[string]any{
			"dexID":          config.DexID,
			"swapFee":        config.SwapFee,
			"factoryAddress": config.FactoryAddress,
			"newPoolLimit":   config.NewPoolLimit,
		},
		Dependencies: poollist.Dependencies{
			EthrpcClient: ethrpcClient,
		},
	}

	poolsListerFactory := poollist.Factory(uniswap.DexTypeUniswap)
	if poolsListerFactory == nil {
		fmt.Println("❌ Ошибка: не найдена фабрика для uniswap")
		return
	}

	poolsLister, err := poolsListerFactory("uniswap", factoryParams)
	if err != nil {
		fmt.Printf("❌ Ошибка создания PoolsListUpdater: %v\n", err)
		return
	}

	// 4. Получаем список пулов
	fmt.Println("\n📋 Получаем список пулов...")
	pools, _, err := poolsLister.GetNewPools(ctx, nil)
	if err != nil {
		fmt.Printf("❌ Ошибка получения пулов: %v\n", err)
		return
	}

	fmt.Printf("✅ Получено пулов: %d\n", len(pools))
	if len(pools) == 0 {
		fmt.Println("⚠️  Пулы не найдены. Попробуйте другой RPC или проверьте конфигурацию.")
		return
	}

	// 5. Создаем PoolTracker для получения резервов
	trackerFactory := pooltrack.Factory(uniswap.DexTypeUniswap)
	if trackerFactory == nil {
		fmt.Println("❌ Ошибка: не найдена фабрика трекера для uniswap")
		return
	}

	trackerParams := pooltrack.FactoryParams{
		Exchange: "uniswap",
		Dependencies: pooltrack.Dependencies{
			EthrpcClient: ethrpcClient,
		},
	}

	poolTracker, err := trackerFactory("uniswap", trackerParams)
	if err != nil {
		fmt.Printf("❌ Ошибка создания PoolTracker: %v\n", err)
		return
	}

	// 6. Показываем информацию о первых N пулах и получаем их резервы
	numPoolsToShow := 5
	if len(pools) < numPoolsToShow {
		numPoolsToShow = len(pools)
	}

	fmt.Printf("\n📊 Информация о первых %d пулах:\n", numPoolsToShow)
	fmt.Println("=" + string(make([]byte, 100)) + "=")

	for i := 0; i < numPoolsToShow; i++ {
		poolEntity := pools[i]
		fmt.Printf("\n🏊 Пул #%d:\n", i+1)
		fmt.Printf("   Адрес: %s\n", poolEntity.Address)
		fmt.Printf("   Токены: %s / %s\n", poolEntity.Tokens[0].Address, poolEntity.Tokens[1].Address)

		// Получаем актуальные резервы
		updatedPool, err := poolTracker.GetNewPoolState(ctx, poolEntity, pool.GetNewPoolStateParams{
			Logs: nil, // Можно передать логи, если есть
		})
		if err != nil {
			fmt.Printf("   ⚠️  Ошибка получения резервов: %v\n", err)
			continue
		}

		fmt.Printf("   Резервы: %s / %s\n", updatedPool.Reserves[0], updatedPool.Reserves[1])

		// 7. Создаем PoolSimulator для оффчейн калькуляции
		poolSim, err := uniswap.NewPoolSimulator(updatedPool)
		if err != nil {
			fmt.Printf("   ⚠️  Ошибка создания симулятора: %v\n", err)
			continue
		}

		// Пример калькуляции: свап 1 токен первого типа на второй
		token0Amount := big.NewInt(1)
		// Умножаем на 10^18 для правильного формата (предполагаем 18 decimals)
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

		// Конвертируем результат в читаемый формат
		amountOut := new(big.Float).SetInt(result.TokenAmountOut.Amount)
		amountOut.Quo(amountOut, big.NewFloat(1e18))

		fmt.Printf("   💱 Калькуляция свапа:\n")
		fmt.Printf("      Вход: 1.0 %s\n", getTokenSymbol(updatedPool.Tokens[0].Address))
		fmt.Printf("      Выход: %s %s\n", amountOut.Text('f', 6), getTokenSymbol(updatedPool.Tokens[1].Address))
		fmt.Printf("      Комиссия: %.2f%%\n", updatedPool.SwapFee*100)
		fmt.Printf("      Газ: %d\n", result.Gas)
	}

	fmt.Println("\n" + string(make([]byte, 100)) + "=")
	fmt.Println("✅ Готово!")
}

// Простая функция для получения символа токена (можно расширить)
func getTokenSymbol(address string) string {
	// Нормализуем адрес (убираем префикс 0x если есть, приводим к нижнему регистру)
	normalized := address
	if len(normalized) > 2 && normalized[:2] == "0x" {
		normalized = normalized[2:]
	}
	normalized = "0x" + normalized

	// Несколько известных токенов для примера
	tokenMap := map[string]string{
		"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2": "WETH",
		"0xdac17f958d2ee523a2206206994597c13d831ec7": "USDT",
		"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48": "USDC",
		"0x6b175474e89094c44da98b954eedeac495271d0f": "DAI",
	}
	
	// Проверяем оба варианта (с 0x и без)
	if symbol, ok := tokenMap[normalized]; ok {
		return symbol
	}
	if symbol, ok := tokenMap[address]; ok {
		return symbol
	}
	
	// Если не нашли, показываем первые 8 символов адреса
	if len(address) > 10 {
		return address[:10] + "..."
	}
	return address
}

