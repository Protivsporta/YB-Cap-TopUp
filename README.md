# YieldBasis Multi-Pool Cap Monitor

This Go application monitors multiple YieldBasis pool capacity changes by listening to `AllocateStablecoins` events and sends real-time Telegram notifications with token identification when new deposit capacity becomes available.

## Features

- 🔍 Real-time monitoring of `AllocateStablecoins` events across multiple pools
- 🎯 **Multi-token support**: WBTC, TBTC, and CBBTC pools simultaneously
- 📱 Instant Telegram notifications with token-specific formatting
- 🏷️ **Token identification**: Clear labeling of which pool's cap was updated
- 🔄 Automatic reconnection on connection failures
- 💰 Formatted amounts for easy reading
- 🔗 Direct links to Etherscan transaction details
- 🎨 Token-specific emojis and formatting

## Setup

### 1. Install Dependencies

```bash
cd cap-monitor
go mod tidy
```

### 2. Create Telegram Bot

1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Create a new bot with `/newbot`
3. Save your bot token

### 3. Get Telegram Chat ID

Option A - Personal Chat:
1. Message your bot
2. Visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
3. Look for your `chat.id` in the response

Option B - Channel/Group:
1. Add your bot to the channel/group as admin
2. Use the channel username (e.g., `@your_channel`) as chat ID

### 4. Configure Environment

Create `.env` file:

```bash
cp .env.example .env
```

Edit `.env`:

```env
# Your Infura WebSocket URL (API key is already set as "AllocateStablecoins")
INFURA_WS_URL=wss://mainnet.infura.io/ws/v3/AllocateStablecoins

# Your Telegram bot token from BotFather
TELEGRAM_BOT_TOKEN=your_bot_token_here

# Your Telegram chat ID or @channel_name
TELEGRAM_CHAT_ID=your_chat_id_or_@channel

# YieldBasis Pool Addresses (comma-separated, all three pools are monitored by default)
# WBTC, TBTC, CBBTC pools
POOL_ADDRESSES=0x6095a220C5567360d459462A25b1AD5aEAD45204,0x2B513eBe7070Cff91cf699a0BFe5075020C732FF,0xD6a1147666f6E4d7161caf436d9923D44d901112
```

## Running

### Development
```bash
go run main.go
```

### Production Build
```bash
go build -o cap-monitor
./cap-monitor
```

### Background Service
```bash
nohup ./cap-monitor > monitor.log 2>&1 &
```

## How It Works

1. **Event Detection**: Listens to Ethereum mainnet via Infura WebSocket for `AllocateStablecoins` events
2. **Cap Calculation**: When stablecoins are allocated to the AMM, the deposit cap increases
3. **Notification**: Sends formatted Telegram message with:
   - New allocation amount
   - Change from previous allocation
   - Allocator address
   - Transaction hash with Etherscan link

## Notification Format

Each pool gets token-specific formatting:

### WBTC Pool
```
🚀 YieldBasis WBTC Pool Cap Update

₿ Pool: WBTC Pool
📊 Event: AllocateStablecoins
📍 Address: 0x6095a220...

📈 Allocation: 1000000.00 stablecoins
💯 Allocated: 1200000.00 stablecoins
🔄 Change: 📈 +200000.00 increase

👤 Allocator: 0x1234...5678
🔗 Transaction: View on Etherscan

New WBTC deposit capacity available! 🚀
```

### TBTC Pool
```
🚀 YieldBasis TBTC Pool Cap Update

🌟 Pool: TBTC Pool
📊 Event: AllocateStablecoins
📍 Address: 0x2B513eBe70...

📈 Allocation: 500000.00 stablecoins
💯 Allocated: 600000.00 stablecoins
🔄 Change: 📈 +100000.00 increase

👤 Allocator: 0x1234...5678
🔗 Transaction: View on Etherscan

New TBTC deposit capacity available! 🚀
```

### CBBTC Pool
```
🚀 YieldBasis CBBTC Pool Cap Update

🔷 Pool: CBBTC Pool
📊 Event: AllocateStablecoins
📍 Address: 0xD6a1147666...

📈 Allocation: 750000.00 stablecoins
💯 Allocated: 900000.00 stablecoins
🔄 Change: 📈 +150000.00 increase

👤 Allocator: 0x1234...5678
🔗 Transaction: View on Etherscan

New CBBTC deposit capacity available! 🚀
```

## Event Details

The `AllocateStablecoins` event signals that:
- New stablecoins have been allocated to the pool's AMM
- Deposit capacity has increased
- `max_debt()` has been updated in the AMM contract
- New WBTC deposits are now possible up to the new limit

## Troubleshooting

### Connection Issues
- Verify Infura WebSocket URL and API key
- Check internet connection
- The script will auto-retry every 10 seconds

### Telegram Issues
- Verify bot token is correct
- Ensure bot can send messages to your chat/channel
- For channels: bot must be added as admin

### No Events
- Verify pool address is correct
- Check if pool is active
- Events only occur when admins allocate new stablecoins

## Pool Information

The monitor automatically tracks these YieldBasis pools:

| Token | Pool Address | Emoji |
|-------|-------------|-------|
| WBTC  | `0x6095a220C5567360d459462A25b1AD5aEAD45204` | ₿ |
| TBTC  | `0x2B513eBe7070Cff91cf699a0BFe5075020C732FF` | 🌟 |
| CBBTC | `0xD6a1147666f6E4d7161caf436d9923D44d901112` | 🔷 |

## Customizing Monitored Pools

To monitor only specific pools, modify the `POOL_ADDRESSES` in your `.env`:

```env
# Monitor only WBTC and TBTC pools
POOL_ADDRESSES=0x6095a220C5567360d459462A25b1AD5aEAD45204,0x2B513eBe7070Cff91cf699a0BFe5075020C732FF
```

## Security Notes

- Keep your `.env` file secure and never commit it
- The Infura API key "AllocateStablecoins" is used as provided
- Telegram bot token should be kept private