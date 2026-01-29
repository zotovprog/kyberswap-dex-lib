# Пример получения пулов SomeSwap и оффчейн калькуляции

Этот пример показывает, как:
1. Получить список пулов SomeSwap с блокчейна
2. Получить актуальные резервы для каждого пула
3. Выполнить оффчейн калькуляцию свапа с учетом кастомного fee

## Быстрый запуск

```bash
cd examples/get_someswap_pools
go run main.go
```

## Что делает программа

- Подключается к Monad RPC
- Берет первые 10 пулов из SomeSwap Factory
- Читает `token0`, `token1`, `fee()` для каждого пула
- Обновляет резервы через `getReserves()` или `Sync`
- Считает amountOut с учетом входного/выходного fee

## Важно

- Константы в `main.go` можно менять:
  - `rpcURL`
  - `factoryAddress`
  - `multicallAddress`
- Если токены не 18 decimals, расчет amountOut будет приблизительным (мы используем 1e18 как вход)

## Проверка ABI

ABI для Factory/Pair лежат в:
- `pkg/liquidity-source/someswap/abis/SomeFactory.json`
- `pkg/liquidity-source/someswap/abis/SomePair.json`

Если ваши контракты отличаются по сигнатурам — пришлите ABI, я подгоню.
