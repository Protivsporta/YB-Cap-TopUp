package main

import (
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTelegramBot(t *testing.T) (*tgbotapi.BotAPI, string) {
	// Load .env file
	err := godotenv.Load()
	require.NoError(t, err, "Failed to load .env file")

	// Get configuration from .env
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	require.NotEmpty(t, botToken, "TELEGRAM_BOT_TOKEN must be set in .env file")
	require.NotEmpty(t, chatID, "TELEGRAM_CHAT_ID must be set in .env file")

	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(botToken)
	require.NoError(t, err, "Failed to create Telegram bot")

	t.Logf("✅ Telegram bot initialized: %s", bot.Self.UserName)
	t.Logf("📱 Sending test notifications to: %s", chatID)

	return bot, chatID
}

func createTestBigInt(amount string) *big.Int {
	value := new(big.Int)
	value.SetString(amount, 10)
	return value
}

func getTestEvents() []AllocateStablecoinsEvent {
	return []AllocateStablecoinsEvent{
		{
			// Real-world event example from Etherscan
			Allocator:            common.HexToAddress("0x370a449fe8b9411c95bf897021377fe007D100c0"),
			StablecoinAllocation: createTestBigInt("200000000000000000000000000"), // 200,000,000 * 1e18
			StablecoinAllocated:  createTestBigInt("0"),                           // 0 * 1e18 (no allocation yet)
			PoolAddress:          common.HexToAddress("0x6095a220C5567360d459462A25b1AD5aEAD45204"),
			TokenName:            "WBTC",
		},
		// {
		// 	Allocator:            common.HexToAddress("0x1234567890123456789012345678901234567890"),
		// 	StablecoinAllocation: createTestBigInt("1000000000000000000000000"), // 1,000,000 * 1e18
		// 	StablecoinAllocated:  createTestBigInt("1200000000000000000000000"), // 1,200,000 * 1e18
		// 	PoolAddress:          common.HexToAddress("0x6095a220C5567360d459462A25b1AD5aEAD45204"),
		// 	TokenName:           "WBTC",
		// },
		// {
		// 	Allocator:            common.HexToAddress("0x9876543210987654321098765432109876543210"),
		// 	StablecoinAllocation: createTestBigInt("500000000000000000000000"), // 500,000 * 1e18
		// 	StablecoinAllocated:  createTestBigInt("750000000000000000000000"), // 750,000 * 1e18
		// 	PoolAddress:          common.HexToAddress("0x2B513eBe7070Cff91cf699a0BFe5075020C732FF"),
		// 	TokenName:           "TBTC",
		// },
		// {
		// 	Allocator:            common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
		// 	StablecoinAllocation: createTestBigInt("800000000000000000000000"), // 800,000 * 1e18
		// 	StablecoinAllocated:  createTestBigInt("950000000000000000000000"), // 950,000 * 1e18
		// 	PoolAddress:          common.HexToAddress("0xD6a1147666f6E4d7161caf436d9923D44d901112"),
		// 	TokenName:           "CBBTC",
		// },
	}
}

func getTestTxHashes() []string {
	return []string{
		"0x6f16a992fba7b966942bb604775e154f99de09fce962940d4e284c5c9132868", // Real-world tx hash from Etherscan
		"0xa1b2c3d4e5f6789012345678901234567890123456789012345678901234567890",
		"0xb2c3d4e5f6789012345678901234567890123456789012345678901234567890a1",
		"0xc3d4e5f6789012345678901234567890a1b2c3d4e5f6789012345678901234567890",
	}
}

func TestTelegramNotifications(t *testing.T) {
	bot, chatID := setupTelegramBot(t)
	testEvents := getTestEvents()
	testTxHashes := getTestTxHashes()

	t.Run("SendAllPoolNotifications", func(t *testing.T) {
		for i, event := range testEvents {
			t.Run(fmt.Sprintf("%sPool", event.TokenName), func(t *testing.T) {
				t.Logf("📤 Sending test notification for %s pool...", event.TokenName)

				err := sendTelegramNotification(bot, chatID, &event, testTxHashes[i])
				assert.NoError(t, err, "Failed to send %s test notification", event.TokenName)

				t.Logf("✅ %s test notification sent successfully!", event.TokenName)

				// Small delay between messages to avoid rate limiting
				if i < len(testEvents)-1 {
					time.Sleep(1 * time.Second)
				}
			})
		}
	})

	t.Logf("🎉 Test completed! Check your channel/chat: %s", chatID)
	t.Log("📋 You should see 1 test notification (real-world WBTC event)")
	t.Log("🔧 If you don't see messages, check:")
	t.Log("   - Bot is added to channel as admin")
	t.Log("   - Bot has 'Post Messages' permission")
	t.Log("   - Chat ID is correct")
}
func TestTelegramChannelVsChatID(t *testing.T) {
	bot, chatID := setupTelegramBot(t)
	event := getTestEvents()[0] // Use WBTC event for testing
	txHash := getTestTxHashes()[0]

	t.Run("ChannelIDHandling", func(t *testing.T) {
		if strings.HasPrefix(chatID, "@") {
			t.Logf("Testing channel ID format: %s", chatID)
		} else {
			t.Logf("Testing numeric chat ID: %s", chatID)
		}

		err := sendTelegramNotification(bot, chatID, &event, txHash)
		assert.NoError(t, err, "Should handle chat ID format correctly")
	})
}

func TestMessageContentValidation(t *testing.T) {
	testEvents := getTestEvents()

	for _, event := range testEvents {
		t.Run(fmt.Sprintf("Validate%sEvent", event.TokenName), func(t *testing.T) {
			t.Parallel()

			// Validate allocation amounts
			assert.NotNil(t, event.StablecoinAllocation, "StablecoinAllocation should not be nil")
			assert.NotNil(t, event.StablecoinAllocated, "StablecoinAllocated should not be nil")
			assert.True(t, event.StablecoinAllocation.Sign() > 0, "StablecoinAllocation should be positive")
			assert.True(t, event.StablecoinAllocated.Sign() >= 0, "StablecoinAllocated should be non-negative")

			// Validate addresses
			assert.NotEqual(t, common.Address{}, event.Allocator, "Allocator address should not be zero")
			assert.NotEqual(t, common.Address{}, event.PoolAddress, "Pool address should not be zero")

			// Validate token name
			assert.NotEmpty(t, event.TokenName, "Token name should not be empty")
			assert.Contains(t, []string{"WBTC", "TBTC", "CBBTC"}, event.TokenName, "Token name should be valid")
		})
	}
}

func TestConfigurationLoading(t *testing.T) {
	t.Run("LoadConfigFromEnv", func(t *testing.T) {
		config, err := loadConfig()

		if err != nil {
			t.Skipf("Skipping config test due to missing environment: %v", err)
			return
		}

		assert.NotEmpty(t, config.TelegramToken, "Telegram token should be loaded")
		assert.NotEmpty(t, config.TelegramChatID, "Telegram chat ID should be loaded")
		assert.NotEmpty(t, config.Pools, "Pools should be configured")

		// Validate pool configuration
		for _, pool := range config.Pools {
			assert.NotEqual(t, common.Address{}, pool.Address, "Pool address should not be zero")
			assert.NotEmpty(t, pool.TokenName, "Pool token name should not be empty")
		}
	})
}

func TestABILoading(t *testing.T) {
	t.Run("LoadContractABI", func(t *testing.T) {
		contractABI, err := loadABI()

		if err != nil {
			t.Skipf("Skipping ABI test due to missing file: %v", err)
			return
		}

		// Check if the AllocateStablecoins event exists in the ABI
		event, exists := contractABI.Events["AllocateStablecoins"]
		assert.True(t, exists, "AllocateStablecoins event should exist in ABI")
		assert.NotNil(t, event, "Event should not be nil")
	})
}
