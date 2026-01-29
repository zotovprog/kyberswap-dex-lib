# Пример получения пулов Uniswap V2 и оффчейн калькуляции

Этот пример показывает, как:
1. Получить список пулов Uniswap V2 с блокчейна
2. Получить актуальные резервы для каждого пула
3. Выполнить оффчейн калькуляцию свапа

## 🇷🇺 Краткая инструкция на русском

Этот пример настроен для работы с **Monad сетью** и поможет вам:
- **Получить список пулов** Uniswap V2 с Monad mainnet
- **Узнать резервы** каждого пула (сколько токенов в пуле)
- **Рассчитать свап** оффчейн (сколько вы получите при обмене токенов)

### ⚡ Быстрая команда для запуска:

```bash
# Из корня проекта kyberswap-dex-lib
cd examples/get_uniswap_pools
FACTORY_ADDRESS="0x..." go run main.go
```

**Где взять адрес Factory?** Это адрес вашего Uniswap V2 Factory контракта на Monad. Если у вас своя биржа, это адрес контракта, который создает пары токенов.

### Быстрый старт для Monad:

#### Способ 1: Через переменные окружения (рекомендуется)

1. Откройте терминал в корне проекта `kyberswap-dex-lib`
2. Установите переменную окружения с адресом Factory:
   ```bash
   export FACTORY_ADDRESS="0x..." # Замените на адрес вашей Uniswap V2 Factory на Monad
   ```
3. Перейдите в папку с примером и запустите:
   ```bash
   cd examples/get_uniswap_pools
   go run main.go
   ```

#### Способ 2: Все в одной команде

```bash
cd examples/get_uniswap_pools
FACTORY_ADDRESS="0x..." go run main.go
```

#### Способ 3: Указать адрес в коде

1. Откройте файл `examples/get_uniswap_pools/main.go`
2. Найдите строку ~43 и раскомментируйте/измените:
   ```go
   factoryAddress = "0x..." // <-- Вставьте адрес вашей Factory здесь
   ```
3. Запустите:
   ```bash
   cd examples/get_uniswap_pools
   go run main.go
   ```

Программа автоматически:
- Подключится к Monad RPC (https://rpc-mainnet.monadinfra.com)
- Получит первые 10 пулов из вашей Uniswap V2 Factory
- Покажет резервы и рассчитает пример свапа для первых 5 пулов

### Использование другого RPC:

Если хотите использовать другой RPC:
```bash
export RPC_URL="https://your-rpc-url.com"
export FACTORY_ADDRESS="0x..."
go run main.go
```

## Требования

- Go 1.25 или выше
- Доступ к Monad RPC (по умолчанию используется https://rpc-mainnet.monadinfra.com)
- Адрес Uniswap V2 Factory контракта на Monad

## 📝 Где прописывать переменные окружения

### Вариант 1: В терминале (перед запуском)

```bash
export FACTORY_ADDRESS="0x..."
export RPC_URL="https://rpc-mainnet.monadinfra.com/rpc/elHriSqJvel4qXhCU462GZDRcymRpGlA"
cd examples/get_uniswap_pools
go run main.go
```

### Вариант 2: В одной команде

```bash
cd examples/get_uniswap_pools
FACTORY_ADDRESS="0x..." go run main.go
```

### Вариант 3: В коде (для постоянного использования)

Откройте `main.go` и найдите строку ~43, раскомментируйте и укажите адрес:
```go
factoryAddress = "0x..." // Ваш адрес Factory
```

### Вариант 4: Через .env файл (если используете библиотеку для загрузки .env)

Создайте файл `.env` в папке `examples/get_uniswap_pools/`:
```
FACTORY_ADDRESS=0x...
RPC_URL=https://rpc-mainnet.monadinfra.com/rpc/elHriSqJvel4qXhCU462GZDRcymRpGlA
```

**Примечание:** Текущий код не загружает .env автоматически, используйте варианты 1-3.

## 🚀 Как запустить

**Важно:** Запускайте из корня проекта `kyberswap-dex-lib`, так как пример использует импорты из этого модуля.

### Шаг 1: Установите переменную окружения

**ОБЯЗАТЕЛЬНО** нужно указать адрес Factory контракта:

```bash
export FACTORY_ADDRESS="0x..." # Замените на адрес вашей Factory на Monad
```

### Шаг 2: Запустите программу

```bash
# Из корня проекта
cd examples/get_uniswap_pools
go run main.go
```

**Или все в одной команде:**
```bash
cd examples/get_uniswap_pools && FACTORY_ADDRESS="0x..." go run main.go
```

### Что вы увидите:

```
🔗 Подключаемся к Monad RPC: https://rpc-mainnet.monadinfra.com/rpc/***
📞 Multicall контракт: 0xcA11bde05977b3631167028862bE2a173976CA11
🏭 Factory адрес: 0x...

📋 Получаем список пулов...
✅ Получено пулов: 10

📊 Информация о первых 5 пулах:
...
```

## Запуск (детальная инструкция)

### Вариант 1: Использовать Monad RPC (по умолчанию)

```bash
# Из корня проекта kyberswap-dex-lib

# 1. Установите переменную окружения с адресом Factory
export FACTORY_ADDRESS="0x..." # Адрес вашей Uniswap V2 Factory на Monad

# 2. Перейдите в папку с примером
cd examples/get_uniswap_pools

# 3. Запустите
go run main.go
```

**Или все в одной команде:**
```bash
cd examples/get_uniswap_pools
FACTORY_ADDRESS="0x..." go run main.go
```

### Вариант 2: Использовать другой RPC URL

```bash
export RPC_URL="https://your-rpc-url.com"
export FACTORY_ADDRESS="0x..."
cd examples/get_uniswap_pools
go run main.go
```

### Вариант 3: Указать адрес Factory в коде

Откройте `main.go` и найдите строку ~35, раскомментируйте и укажите адрес:
```go
factoryAddress = "0x..." // Ваш адрес Factory на Monad
```

### Если возникают проблемы с зависимостями

Убедитесь, что вы находитесь в корне проекта и зависимости установлены:
```bash
# Из корня kyberswap-dex-lib
go mod download
cd examples/get_uniswap_pools
go run main.go
```

## Что делает программа

1. **Подключается к RPC** - использует публичный RPC или ваш собственный
2. **Получает список пулов** - через `PoolsListUpdater` получает первые 10 пулов из Uniswap V2 Factory
3. **Получает резервы** - для каждого пула получает актуальные резервы через `PoolTracker`
4. **Выполняет калькуляцию** - для каждого пула рассчитывает, сколько токенов вы получите при свапе 1 токена первого типа на второй

## Вывод программы

Программа выводит:
- Адрес каждого пула
- Адреса токенов в пуле
- Текущие резервы
- Результат калькуляции свапа (сколько вы получите при свапе 1 токена)

## Настройка

Вы можете изменить параметры в коде:

```go
config := &uniswap.Config{
    DexID:          "uniswap",
    SwapFee:        0.003, // Комиссия пула (0.3%)
    FactoryAddress: "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f", // Адрес фабрики
    NewPoolLimit:   10,    // Количество пулов для получения
}
```

И количество пулов для отображения:
```go
numPoolsToShow := 5 // Показывать первые 5 пулов
```

## Пример вывода

```
🔗 Подключаемся к RPC: https://ethereum.kyberengineering.io

📋 Получаем список пулов...
✅ Получено пулов: 10

📊 Информация о первых 5 пулах:
====================================================================================================

🏊 Пул #1:
   Адрес: 0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852
   Токены: 0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2 / 0xdac17f958d2ee523a2206206994597c13d831ec7
   Резервы: 32981129686811504138006 / 83362838693979
   💱 Калькуляция свапа:
      Вход: 1.0 WETH
      Выход: 2527.123456 USDT
      Комиссия: 0.30%
      Газ: 60000
```

## Примечания

- Программа использует публичный RPC по умолчанию, который может быть медленным
- Для продакшена рекомендуется использовать свой RPC провайдер (Infura, Alchemy, QuickNode)
- Multicall контракт настроен для Ethereum mainnet (адрес: `0x5ba1e12693dc8f9c48aad8770482f4739beed696`)
- Для других сетей нужно изменить адрес multicall контракта
