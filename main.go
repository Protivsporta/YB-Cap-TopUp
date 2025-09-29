package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type AllocateStablecoinsEvent struct {
	Allocator            common.Address
	StablecoinAllocation *big.Int
	StablecoinAllocated  *big.Int
	PoolAddress          common.Address // Added to identify which pool
	TokenName            string         // Added to identify token name
	YBURL                string         // Added to include YieldBasis interface URL
}

type ApprovalEvent struct {
	Owner       common.Address
	Spender     common.Address
	Value       *big.Int
	PoolAddress common.Address // Added to identify which pool
	TokenName   string         // Added to identify token name
}

type PoolInfo struct {
	Address   common.Address
	TokenName string
	YBURL     string
}

type Config struct {
	InfuraWsURL    string
	TelegramToken  string
	TelegramChatID string
	Pools          []PoolInfo
}

func loadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		InfuraWsURL:    os.Getenv("INFURA_WS_URL"),
		TelegramToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID: os.Getenv("TELEGRAM_CHAT_ID"),
	}

	// Validate required env vars
	if config.InfuraWsURL == "" {
		return nil, fmt.Errorf("INFURA_WS_URL is required")
	}
	if config.TelegramToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if config.TelegramChatID == "" {
		return nil, fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}

	// Parse pool addresses
	poolAddresses := os.Getenv("POOL_ADDRESSES")
	if poolAddresses == "" {
		// Default pools if not specified
		poolAddresses = "0x6095a220C5567360d459462A25b1AD5aEAD45204,0x2B513eBe7070Cff91cf699a0BFe5075020C732FF,0xD6a1147666f6E4d7161caf436d9923D44d901112"
	}

	// Define pool to token mapping (all addresses in lowercase for consistent lookup)
	poolTokenMap := map[string]string{
		"0x6095a220c5567360d459462a25b1ad5aead45204": "WBTC",
		"0x2b513ebe7070cff91cf699a0bfe5075020c732ff": "TBTC",
		"0xd6a1147666f6e4d7161caf436d9923d44d901112": "CBBTC",
	}

	// Define YieldBasis interface URLs for each pool
	poolYBURLMap := map[string]string{
		"0xa5bfb61af14afe7b81cac7fa4f7c4483dedc36df": "https://yieldbasis.com/market/0x6095a220C5567360d459462A25b1AD5aEAD45204",
		"0x2b513ebe7070cff91cf699a0bfe5075020c732ff": "https://yieldbasis.com/market/0x2B513eBe7070Cff91cf699a0BFe5075020C732FF",
		"0xd6a1147666f6e4d7161caf436d9923d44d901112": "https://yieldbasis.com/market/0xD6a1147666f6E4d7161caf436d9923D44d901112",
	}

	// Parse addresses and create pool info
	addressList := strings.Split(poolAddresses, ",")
	for _, addr := range addressList {
		addr = strings.TrimSpace(addr)
		if !common.IsHexAddress(addr) {
			return nil, fmt.Errorf("invalid pool address: %s", addr)
		}

		tokenName := poolTokenMap[strings.ToLower(addr)]
		if tokenName == "" {
			tokenName = "UNKNOWN"
		}

		ybURL := poolYBURLMap[strings.ToLower(addr)]
		if ybURL == "" {
			ybURL = "https://yieldbasis.com"
		}

		config.Pools = append(config.Pools, PoolInfo{
			Address:   common.HexToAddress(addr),
			TokenName: tokenName,
			YBURL:     ybURL,
		})
	}

	if len(config.Pools) == 0 {
		return nil, fmt.Errorf("no valid pool addresses found")
	}

	return config, nil
}

func loadABI() (abi.ABI, error) {
	abiBytes, err := os.ReadFile("abi.json")
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to read ABI file: %v", err)
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse ABI: %v", err)
	}

	return contractABI, nil
}

func sendTelegramNotification(bot *tgbotapi.BotAPI, chatID string, event *AllocateStablecoinsEvent, txHash string) error {
	// Format amounts in a readable way (assuming 18 decimals for stablecoin)
	allocation := new(big.Float).Quo(new(big.Float).SetInt(event.StablecoinAllocation), big.NewFloat(1e18))
	allocated := new(big.Float).Quo(new(big.Float).SetInt(event.StablecoinAllocated), big.NewFloat(1e18))

	difference := new(big.Int).Sub(event.StablecoinAllocated, event.StablecoinAllocation)
	diffFloat := new(big.Float).Quo(new(big.Float).SetInt(difference.Abs(difference)), big.NewFloat(1e18))

	var changeText string
	if difference.Sign() > 0 {
		changeText = fmt.Sprintf("+%.2f increase", diffFloat)
	} else if difference.Sign() < 0 {
		changeText = fmt.Sprintf("-%.2f decrease", diffFloat)
	} else {
		changeText = "No change"
	}

	message := fmt.Sprintf(`🚀 *YieldBasis %s Pool Cap Update*
	
*YieldBasis Interface*: [View %s Pool](%s)

*Pool*: %s Pool
*Event*: AllocateStablecoins

*Allocation*: %.2f stablecoins
*Allocated*: %.2f stablecoins
*Change*: %s

*Transaction*: [View on Etherscan](https://etherscan.io/tx/%s)

*New %s deposit capacity available!*`,
		event.TokenName,
		event.TokenName, event.YBURL,
		event.TokenName,
		allocation, allocated, changeText,
		txHash,
		event.TokenName)

	msg := tgbotapi.NewMessageToChannel("@"+chatID, message)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	// If chatID is not a channel, treat it as a regular chat ID
	if !strings.HasPrefix(chatID, "@") {
		// Parse chat ID as int64
		if chatIDInt, err := strconv.ParseInt(chatID, 10, 64); err == nil {
			msg.ChatID = chatIDInt
		}
	}

	_, err := bot.Send(msg)
	return err
}

func sendUnparsedEventNotification(bot *tgbotapi.BotAPI, chatID string, eventType string, tokenName string, poolAddress common.Address, txHash string, rawData []byte) error {
	message := fmt.Sprintf(`🚀 *YieldBasis %s Pool Event Detected*

*Pool*: %s Pool
*Event*: %s (parsing failed)
*Address*: %s

*Raw Event Data*: %s

*Transaction*: [View on Etherscan](https://etherscan.io/tx/%s)

*Event detected but could not be parsed - please check transaction for details*`,
		tokenName,
		tokenName, eventType, poolAddress.Hex()[:10]+"...",
		fmt.Sprintf("0x%x", rawData),
		txHash)

	msg := tgbotapi.NewMessageToChannel("@"+chatID, message)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	// If chatID is not a channel, treat it as a regular chat ID
	if !strings.HasPrefix(chatID, "@") {
		// Parse chat ID as int64
		if chatIDInt, err := strconv.ParseInt(chatID, 10, 64); err == nil {
			msg.ChatID = chatIDInt
		}
	}

	_, err := bot.Send(msg)
	return err
}

func monitorEvents(config *Config) error {
	// Connect to Ethereum via WebSocket
	client, err := ethclient.Dial(config.InfuraWsURL)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum client: %v", err)
	}
	defer client.Close()

	// Load contract ABI
	contractABI, err := loadABI()
	if err != nil {
		return fmt.Errorf("failed to load ABI: %v", err)
	}

	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		return fmt.Errorf("failed to create Telegram bot: %v", err)
	}
	bot.Debug = false

	log.Printf("Telegram bot initialized: %s", bot.Self.UserName)

	// Create address to token and URL mapping for quick lookup
	addressToToken := make(map[common.Address]string)
	addressToYBURL := make(map[common.Address]string)
	var contractAddresses []common.Address

	for _, pool := range config.Pools {
		contractAddresses = append(contractAddresses, pool.Address)
		addressToToken[pool.Address] = pool.TokenName
		addressToYBURL[pool.Address] = pool.YBURL
		log.Printf("🔍 Monitoring %s pool: %s", pool.TokenName, pool.Address.Hex())
	}

	log.Println("This version of Cap-Monitor monitors only AllocateStablecoins event")

	// Get the event signature hashes
	allocateSignature := []byte("AllocateStablecoins(address,uint256,uint256)")
	allocateHash := crypto.Keccak256Hash(allocateSignature)

	// Create filter for both AllocateStablecoins and Approval events on all pools
	query := ethereum.FilterQuery{
		Addresses: contractAddresses,
		Topics:    [][]common.Hash{{allocateHash}},
	}

	// Subscribe to logs
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %v", err)
	}
	defer sub.Unsubscribe()

	log.Printf("📱 Telegram notifications will be sent to: %s", config.TelegramChatID)
	log.Printf("🎯 Monitoring %d pools for AllocateStablecoins and Approval events...", len(config.Pools))

	// Monitor for events
	for {
		select {
		case err := <-sub.Err():
			return fmt.Errorf("subscription error: %v", err)
		case vLog := <-logs:
			// Identify which pool this event came from
			tokenName := addressToToken[vLog.Address]
			ybURL := addressToYBURL[vLog.Address]
			eventType := ""

			// Determine event type by topic hash
			if len(vLog.Topics) > 0 {
				if vLog.Topics[0] == allocateHash {
					eventType = "AllocateStablecoins"
				}
			}

			log.Printf("📊 New %s event detected from %s pool! TxHash: %s", eventType, tokenName, vLog.TxHash.Hex())

			var err error
			var notificationSent bool = false

			switch eventType {
			case "AllocateStablecoins":
				// Parse AllocateStablecoins event
				var event AllocateStablecoinsEvent
				parseErr := contractABI.UnpackIntoInterface(&event, "AllocateStablecoins", vLog.Data)
				if parseErr != nil {
					log.Printf("❌ Failed to unpack AllocateStablecoins event: %v", parseErr)
					// Send unparsed notification
					err = sendUnparsedEventNotification(bot, config.TelegramChatID, eventType, tokenName, vLog.Address, vLog.TxHash.Hex(), vLog.Data)
					notificationSent = true
				} else {
					// The allocator address is in the indexed topics
					if len(vLog.Topics) > 1 {
						event.Allocator = common.HexToAddress(vLog.Topics[1].Hex())
					}

					// Add pool identification info
					event.PoolAddress = vLog.Address
					event.TokenName = tokenName
					event.YBURL = ybURL

					// Send Telegram notification
					err = sendTelegramNotification(bot, config.TelegramChatID, &event, vLog.TxHash.Hex())
					notificationSent = true
				}

			}

			// Always report the result
			if notificationSent {
				if err != nil {
					log.Printf("❌ Failed to send Telegram notification: %v", err)
				} else {
					log.Printf("✅ %s notification sent successfully for %s pool", eventType, tokenName)
				}
			}
		}
	}
}

func main() {
	log.Println("🚀 Starting YieldBasis Pool Cap Monitor...")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Monitor events with retry logic
	for {
		err := monitorEvents(config)
		log.Printf("❌ Monitoring stopped: %v", err)
		log.Println("🔄 Retrying in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}
