// Hackathon Starter: Complete AI Financial Agent
// Build intelligent financial tools with nim-go-sdk + Liminal banking APIs
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/executor"
	"github.com/becomeliminal/nim-go-sdk/server"
	"github.com/becomeliminal/nim-go-sdk/tools"
	"github.com/joho/godotenv"
)

func main() {
	// ============================================================================
	// CONFIGURATION
	// ============================================================================
	// Load .env file if it exists (optional - will use system env vars if not found)
	_ = godotenv.Load()

	// Load configuration from environment variables
	// Create a .env file or export these in your shell

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		log.Fatal("❌ ANTHROPIC_API_KEY environment variable is required")
	}

	liminalBaseURL := os.Getenv("LIMINAL_BASE_URL")
	if liminalBaseURL == "" {
		liminalBaseURL = "https://api.liminal.cash"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ============================================================================
	// LIMINAL EXECUTOR SETUP
	// ============================================================================
	// The HTTPExecutor handles all API calls to Liminal banking services.
	// Authentication is handled automatically via JWT tokens passed from the
	// frontend login flow (email/OTP). No API key needed!

	liminalExecutor := executor.NewHTTPExecutor(executor.HTTPExecutorConfig{
		BaseURL: liminalBaseURL,
	})
	log.Println("✅ Liminal API configured")

	// ============================================================================
	// SERVER SETUP
	// ============================================================================
	// Create the nim-go-sdk server with Claude AI
	// The server handles WebSocket connections and manages conversations
	// Authentication is automatic: JWT tokens from the login flow are extracted
	// from WebSocket connections and forwarded to Liminal API calls

	srv, err := server.New(server.Config{
		AnthropicKey:    anthropicKey,
		SystemPrompt:    hackathonSystemPrompt,
		Model:           "claude-sonnet-4-20250514",
		MaxTokens:       4096,
		LiminalExecutor: liminalExecutor, // SDK automatically handles JWT extraction and forwarding
	})
	if err != nil {
		log.Fatal(err)
	}

	// ============================================================================
	// ADD LIMINAL BANKING TOOLS
	// ============================================================================
	// These are the 9 core Liminal tools that give your AI access to real banking:
	//
	// READ OPERATIONS (no confirmation needed):
	//   1. get_balance - Check wallet balance
	//   2. get_savings_balance - Check savings positions and APY
	//   3. get_vault_rates - Get current savings rates
	//   4. get_transactions - View transaction history
	//   5. get_profile - Get user profile info
	//   6. search_users - Find users by display tag
	//
	// WRITE OPERATIONS (require user confirmation):
	//   7. send_money - Send money to another user
	//   8. deposit_savings - Deposit funds into savings
	//   9. withdraw_savings - Withdraw funds from savings

	srv.AddTools(tools.LiminalTools(liminalExecutor)...)
	log.Println("✅ Added 9 Liminal banking tools")

	// ============================================================================
	// ADD CUSTOM TOOLS
	// ============================================================================
	// This is where you'll add your hackathon project's custom tools!
	// Below is an example spending analyzer tool to get you started.

	srv.AddTool(createSpendingAnalyzerTool(liminalExecutor))
	srv.AddTool(createSpendWeeklyGoalTool(liminalExecutor))
	srv.AddTool(createGetWeeklyGoalProgressTool(liminalExecutor))
	srv.AddTool(createCheckWeeklySpendTool(liminalExecutor))
	srv.AddTool(createCategorizeTransactionTool(liminalExecutor))
	log.Println("✅ Added custom spending analyzer, weekly goal, and categorization tools")

	// TODO: Add more custom tools here!
	// Examples:
	//   - Savings goal tracker
	//   - Budget alerts
	//   - Spending category analyzer
	//   - Bill payment predictor
	//   - Cash flow forecaster

	// ============================================================================
	// START SERVER
	// ============================================================================

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🚀 Hackathon Starter Server Running")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📡 WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("Ready for connections! Start your frontend with: cd frontend && npm run dev")
	log.Println()

	if err := srv.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// ============================================================================
// SYSTEM PROMPT
// ============================================================================
// This prompt defines your AI agent's personality and behavior
// Customize this to match your hackathon project's focus!

const hackathonSystemPrompt = `You are Nim, a friendly AI financial assistant built for the Liminal Vibe Banking Hackathon.

WHAT YOU DO:
You help users manage their money using Liminal's stablecoin banking platform. You can check balances, review transactions, send money, and manage savings - all through natural conversation.

CONVERSATIONAL STYLE:
- Be warm, friendly, and conversational - not robotic
- Use casual language when appropriate, but stay professional about money
- Ask clarifying questions when something is unclear
- Remember context from earlier in the conversation
- Explain things simply without being condescending

WHEN TO USE TOOLS:
- Use tools immediately for simple queries ("what's my balance?")
- For actions, gather all required info first ("send $50 to @alice")
- Always confirm before executing money movements
- Don't use tools for general questions about how things work

MONEY MOVEMENT RULES (IMPORTANT):
- ALL money movements require explicit user confirmation
- Show a clear summary before confirming:
  * send_money: "Send $50 USD to @alice"
  * deposit_savings: "Deposit $100 USD into savings"
  * withdraw_savings: "Withdraw $50 USD from savings"
- Never assume amounts or recipients
- Always use the exact currency the user specified

AVAILABLE BANKING TOOLS:
- Check wallet balance (get_balance)
- Check savings balance and APY (get_savings_balance)
- View savings rates (get_vault_rates)
- View transaction history (get_transactions)
- Get profile info (get_profile)
- Search for users (search_users)
- Send money (send_money) - requires confirmation
- Deposit to savings (deposit_savings) - requires confirmation
- Withdraw from savings (withdraw_savings) - requires confirmation

CUSTOM ANALYTICAL TOOLS:
- Analyze spending patterns (analyze_spending)
- Set weekly spending goal (spend_weekly_goal) - requires confirmation
- Check weekly spending progress (get_weekly_spending_progress)
- Quick check weekly spend status (check_weeklyspend) - use this for context
- Categorize spending by transaction notes (categorize_transactions)

TIPS FOR GREAT INTERACTIONS:
- Proactively suggest relevant actions ("Want me to move some to savings?")
- Explain the "why" behind suggestions
- Celebrate financial wins ("Nice! Your savings earned $5 this month!")
- Be encouraging about savings goals
- Make finance feel less intimidating

Remember: You're here to make banking delightful and help users build better financial habits!`

const patternPrompt = `You are a specialized AI for categorizing financial transactions.

Your job is to analyze transaction notes and categorize them accurately.

AVAILABLE CATEGORIES:
- food: Groceries, restaurants, cafes, food delivery, dining
- travel: Transportation, flights, hotels, rideshare, gas, parking
- subscription: Recurring services, streaming, memberships, software
- entertainment: Movies, concerts, games, events, hobbies
- electronics: Gadgets, computers, phones, accessories, tech
- miscellaneous: Everything else that doesn't fit above categories

WHEN CATEGORIZING:
- Read the transaction note carefully
- Choose the single most appropriate category
- Use context clues from the note
- Default to 'miscellaneous' if truly unclear
- Be consistent with similar transactions

RETURN FORMAT:
Always return a JSON object with category breakdown:
{
  "food": 2,
  "travel": 1,
  "subscription": 0,
  "entertainment": 1,
  "electronics": 0,
  "miscellaneous": 1
}`

// ============================================================================
// CUSTOM TOOL: SPENDING ANALYZER
// ============================================================================
// This is an example custom tool that demonstrates how to:
// 1. Define tool parameters with JSON schema
// 2. Call other Liminal tools from within your tool
// 3. Process and analyze the data
// 4. Return useful insights
//
// Use this as a template for your own hackathon tools!

func createSpendingAnalyzerTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("analyze_spending").
		Description("Analyze the user's spending patterns over a specified time period. Returns insights about spending velocity, categories, and trends.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"days": tools.IntegerProperty("Number of days to analyze (default: 30)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Parse input parameters
			var params struct {
				Days int `json:"days"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Default to 30 days if not specified
			if params.Days == 0 {
				params.Days = 30
			}

			// STEP 1: Fetch transaction history
			// We'll call the Liminal get_transactions tool through the executor
			txRequest := map[string]interface{}{
				"limit": 100, // Get up to 100 transactions
			}
			txRequestJSON, _ := json.Marshal(txRequest)

			txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_transactions",
				Input:     txRequestJSON,
				RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
				}, nil
			}

			if !txResponse.Success {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
				}, nil
			}

			// STEP 2: Parse transaction data
			// In a real implementation, you'd parse the actual response structure
			// For now, we'll create a structured analysis

			var transactions []map[string]interface{}
			var txData map[string]interface{}
			if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
				if txArray, ok := txData["transactions"].([]interface{}); ok {
					for _, tx := range txArray {
						if txMap, ok := tx.(map[string]interface{}); ok {
							transactions = append(transactions, txMap)
						}
					}
				}
			}

			// STEP 3: Analyze the data
			analysis := analyzeTransactions(transactions, params.Days)

			// STEP 4: Return insights
			result := map[string]interface{}{
				"period_days":        params.Days,
				"total_transactions": len(transactions),
				"analysis":           analysis,
				"generated_at":       time.Now().Format(time.RFC3339),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// analyzeTransactions processes transaction data and returns insights
func analyzeTransactions(transactions []map[string]interface{}, days int) map[string]interface{} {
	if len(transactions) == 0 {
		return map[string]interface{}{
			"summary": "No transactions found in the specified period",
		}
	}

	// Calculate basic metrics
	var totalSpent, totalReceived float64
	var spendCount, receiveCount int

	// This is a simplified example - you'd do real analysis here:
	// - Group by category/merchant
	// - Calculate daily/weekly averages
	// - Identify spending spikes
	// - Compare to previous periods
	// - Detect recurring payments

	for _, tx := range transactions {
		// Example analysis logic
		txType, _ := tx["type"].(string)
		amount, _ := tx["amount"].(float64)

		switch txType {
		case "send":
			totalSpent += amount
			spendCount++
		case "receive":
			totalReceived += amount
			receiveCount++
		}
	}

	avgDailySpend := totalSpent / float64(days)

	return map[string]interface{}{
		"total_spent":     fmt.Sprintf("%.2f", totalSpent),
		"total_received":  fmt.Sprintf("%.2f", totalReceived),
		"spend_count":     spendCount,
		"receive_count":   receiveCount,
		"avg_daily_spend": fmt.Sprintf("%.2f", avgDailySpend),
		"velocity":        calculateVelocity(spendCount, days),
		"insights": []string{
			fmt.Sprintf("You made %d spending transactions over %d days", spendCount, days),
			fmt.Sprintf("Average daily spend: $%.2f", avgDailySpend),
			"Consider setting up savings goals to build financial cushion",
		},
	}
}

// calculateVelocity determines spending frequency
func calculateVelocity(transactionCount, days int) string {
	txPerWeek := float64(transactionCount) / float64(days) * 7

	switch {
	case txPerWeek < 2:
		return "low"
	case txPerWeek < 7:
		return "moderate"
	default:
		return "high"
	}
}

// ============================================================================
// CUSTOM TOOL: WEEKLY SPENDING GOAL
// ============================================================================
// Sets and tracks weekly spending goals with progress monitoring

// In-memory storage for weekly goals (in production, use a database)
var weeklyGoals = make(map[string]map[string]interface{})

func createSpendWeeklyGoalTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("spend_weekly_goal").
		Description("Set or update a weekly spending goal. Extracts amount and currency from user input and tracks weekly spending progress.").
		RequiresConfirmation(). // Require user confirmation like WRITE OPERATIONS
		Schema(tools.ObjectSchema(map[string]interface{}{
			"amount":   tools.NumberProperty("The weekly spending limit amount"),
			"currency": tools.StringProperty("The currency code (e.g., USD, LIL, USDC)"),
			"action":   tools.StringProperty("Action: 'set' to create/update goal, 'get' to check current progress (default: set)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Amount   float64 `json:"amount"`
				Currency string  `json:"currency"`
				Action   string  `json:"action"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Default action is 'set'
			if params.Action == "" {
				params.Action = "set"
			}

			// Default currency
			if params.Currency == "" {
				params.Currency = "USD"
			}

			userID := toolParams.UserID

			// Handle GET action - check current progress
			if params.Action == "get" {
				return getWeeklySpendingProgress(ctx, liminalExecutor, toolParams)
			}

			// Handle SET action - create or update goal
			if params.Amount <= 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "amount must be greater than 0",
				}, nil
			}

			// Get current week start (Monday)
			now := time.Now()
			weekStart := getWeekStart(now)
			weekEnd := weekStart.AddDate(0, 0, 7)

			// Store the goal
			weeklyGoals[userID] = map[string]interface{}{
				"amount":      params.Amount,
				"currency":    params.Currency,
				"week_start":  weekStart.Format("2006-01-02"),
				"week_end":    weekEnd.Format("2006-01-02"),
				"set_at":      now.Format(time.RFC3339),
			}

			// Get current spending for this week
			progress, err := calculateWeeklyProgress(ctx, liminalExecutor, toolParams, params.Amount, params.Currency)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to calculate progress: %v", err),
				}, nil
			}

			result := map[string]interface{}{
				"status":       "goal_set",
				"goal_set":     true,
				"goal_amount":  params.Amount,
				"currency":     params.Currency,
				"week_start":   weekStart.Format("Monday, Jan 2"),
				"week_end":     weekEnd.Format("Monday, Jan 2"),
				"spent_so_far": progress["spent"],
				"remaining":    progress["remaining"],
				"percentage":   progress["percentage"],
				"on_track":     progress["on_track"],
				"days_left":    progress["days_left"],
				"message":      fmt.Sprintf("Weekly spending goal set to %.2f %s", params.Amount, params.Currency),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// Helper function to get weekly spending progress
func getWeeklySpendingProgress(ctx context.Context, liminalExecutor core.ToolExecutor, toolParams *core.ToolParams) (*core.ToolResult, error) {
	userID := toolParams.UserID
	goal, exists := weeklyGoals[userID]

	if !exists {
		return &core.ToolResult{
			Success: true,
			Data: map[string]interface{}{
				"goal_set": false,
			},
		}, nil
	}

	amount := goal["amount"].(float64)
	currency := goal["currency"].(string)

	progress, err := calculateWeeklyProgress(ctx, liminalExecutor, toolParams, amount, currency)
	if err != nil {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to calculate progress: %v", err),
		}, nil
	}

	result := map[string]interface{}{
		"goal_set":     true,
		"goal_amount":  amount,
		"currency":     currency,
		"week_start":   goal["week_start"],
		"week_end":     goal["week_end"],
		"spent_so_far": progress["spent"],
		"remaining":    progress["remaining"],
		"percentage":   progress["percentage"],
		"on_track":     progress["on_track"],
		"days_left":    progress["days_left"],
	}

	return &core.ToolResult{
		Success: true,
		Data:    result,
	}, nil
}

// Calculate weekly spending progress
func calculateWeeklyProgress(ctx context.Context, liminalExecutor core.ToolExecutor, toolParams *core.ToolParams, goalAmount float64, currency string) (map[string]interface{}, error) {
	// Get transactions from this week
	txRequest := map[string]interface{}{
		"limit": 100,
	}
	txRequestJSON, _ := json.Marshal(txRequest)

	txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
		UserID:    toolParams.UserID,
		Tool:      "get_transactions",
		Input:     txRequestJSON,
		RequestID: toolParams.RequestID,
	})
	if err != nil {
		return nil, err
	}

	if !txResponse.Success {
		return nil, fmt.Errorf("transaction fetch failed: %s", txResponse.Error)
	}
	
	log.Println("Transaction data:")
	log.Println(string(txResponse.Data))

	// Parse transactions
	var transactions []map[string]interface{}
	var txData map[string]interface{}
	if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
		if txArray, ok := txData["transactions"].([]interface{}); ok {
			for _, tx := range txArray {
				if txMap, ok := tx.(map[string]interface{}); ok {
					transactions = append(transactions, txMap)
				}
			}
		}
	}
	// Calculate spending for this week using transaction dates
	weekStart := getWeekStart(time.Now())
	weekEnd := weekStart.AddDate(0, 0, 7)
	var weeklySpending float64

	log.Printf("Week range: %s to %s", weekStart.Format(time.RFC3339), weekEnd.Format(time.RFC3339))

	for _, tx := range transactions {
		// Parse amount (it's a string in the response)
		amountStr, _ := tx["amount"].(string)
		amount := 0.0
		fmt.Sscanf(amountStr, "%f", &amount)
		
		txCurrency, _ := tx["currency"].(string)
		direction, _ := tx["direction"].(string)
		
		// Parse createdAt timestamp (RFC3339 format)
		var txTime time.Time
		if createdAt, ok := tx["createdAt"].(string); ok {
			txTime, _ = time.Parse(time.RFC3339, createdAt)
		}
		
		log.Printf("Transaction: amount=%s (%.2f), currency=%s, direction=%s, date=%s, inWeek=%t", 
			amountStr, amount, txCurrency, direction, txTime.Format("2006-01-02"), 
			!txTime.IsZero() && (txTime.Equal(weekStart) || txTime.After(weekStart)) && txTime.Before(weekEnd))
		
		// Only count spending (debit/negative amounts) from this week in matching currency
		if !txTime.IsZero() && (txTime.Equal(weekStart) || txTime.After(weekStart)) && txTime.Before(weekEnd) {
			if txCurrency == currency || currency == "" {
				// Count debit transactions (money going out) or negative amounts
				if direction == "debit" || amount < 0 {
					weeklySpending += -amount // Make positive for display
				}
			}
		}
	}

	remaining := goalAmount - weeklySpending
	percentage := (weeklySpending / goalAmount) * 100
	if percentage > 100 {
		percentage = 100
	}

	now := time.Now()
	daysLeft := int(weekEnd.Sub(now).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}

	// Determine if on track
	dayOfWeek := int(now.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 7 // Sunday is 7
	}
	expectedSpending := (goalAmount / 7) * float64(dayOfWeek)
	onTrack := weeklySpending <= expectedSpending

	log.Printf("Weekly goal progress: spent=%.2f remaining=%.2f percentage=%.2f on_track=%t days_left=%d currency=%s goal=%.2f", weeklySpending, remaining, percentage, onTrack, daysLeft, currency, goalAmount)

	return map[string]interface{}{
		"spent":      weeklySpending,
		"remaining":  remaining,
		"percentage": percentage,
		"on_track":   onTrack,
		"days_left":  daysLeft,
	}, nil
}

// Get the start of the current week (Monday)
func getWeekStart(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7
	}
	daysToMonday := weekday - 1
	monday := t.AddDate(0, 0, -daysToMonday)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// ============================================================================
// CUSTOM TOOL: GET WEEKLY GOAL PROGRESS (READ-ONLY)
// ============================================================================
// Read-only version of weekly goal check that doesn't require confirmation

func createGetWeeklyGoalProgressTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("get_weekly_spending_progress").
		Description("Get current weekly spending goal progress without requiring confirmation. Shows how much spent, remaining budget, and on-track status.").
		Schema(tools.ObjectSchema(map[string]interface{}{})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			return getWeeklySpendingProgress(ctx, liminalExecutor, toolParams)
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: CHECK WEEKLY SPEND (CONTEXT TOOL FOR AGENT)
// ============================================================================
// Provides agent with quick context about weekly spending status

func createCheckWeeklySpendTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("check_weeklyspend").
		Description("Check the current weekly spending status. Returns spent amount, remaining budget, percentage used, on-track status, and days left in the week. Use this to get context before answering user questions about their spending.").
		Schema(tools.ObjectSchema(map[string]interface{}{})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			userID := toolParams.UserID
			goal, exists := weeklyGoals[userID]

			if !exists {
				return &core.ToolResult{
					Success: true,
					Data: map[string]interface{}{
						"goal_set": false,
						"message":  "No weekly spending goal has been set yet",
					},
				}, nil
			}

			amount := goal["amount"].(float64)
			currency := goal["currency"].(string)

			progress, err := calculateWeeklyProgress(ctx, liminalExecutor, toolParams, amount, currency)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to calculate progress: %v", err),
				}, nil
			}

			result := map[string]interface{}{
				"goal_set":    true,
				"goal_amount": amount,
				"currency":    currency,
				"spent":       progress["spent"],
				"remaining":   progress["remaining"],
				"percentage":  progress["percentage"],
				"on_track":    progress["on_track"],
				"days_left":   progress["days_left"],
				"week_start":  goal["week_start"],
				"week_end":    goal["week_end"],
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: CATEGORIZE TRANSACTIONS
// ============================================================================
// Analyzes transaction notes and categorizes spending patterns using Claude structured output

func createCategorizeTransactionTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("categorize_transactions").
		Description("Analyze transaction notes and categorize spending into: food, travel, subscription, entertainment, electronics, miscellaneous using AI-powered categorization.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"limit": tools.IntegerProperty("Number of transactions to analyze (default: 50)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Limit int `json:"limit"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			if params.Limit == 0 {
				params.Limit = 50
			}

			// Fetch transactions
			txRequest := map[string]interface{}{"limit": params.Limit}
			txRequestJSON, _ := json.Marshal(txRequest)

			txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_transactions",
				Input:     txRequestJSON,
				RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
				}, nil
			}

			if !txResponse.Success {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
				}, nil
			}

			// Parse transactions
			var transactions []map[string]interface{}
			var txData map[string]interface{}
			if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
				if txArray, ok := txData["transactions"].([]interface{}); ok {
					for _, tx := range txArray {
						if txMap, ok := tx.(map[string]interface{}); ok {
							transactions = append(transactions, txMap)
						}
					}
				}
			}

			// Extract spending transaction notes
			var spendingNotes []string
			for _, tx := range transactions {
				direction, _ := tx["direction"].(string)
				amountStr, _ := tx["amount"].(string)
				amount := 0.0
				fmt.Sscanf(amountStr, "%f", &amount)

				// Only categorize spending (debit or negative)
				if direction == "debit" || amount < 0 {
					if note, ok := tx["note"].(string); ok && note != "" {
						spendingNotes = append(spendingNotes, note)
					}
				}
			}

			if len(spendingNotes) == 0 {
				return &core.ToolResult{
					Success: true,
					Data: map[string]interface{}{
						"categories": map[string]int{
							"food":          0,
							"travel":        0,
							"subscription":  0,
							"entertainment": 0,
							"electronics":   0,
							"miscellaneous": 0,
						},
						"total_analyzed": 0,
						"breakdown":      []string{},
					},
				}, nil
			}

			// Use Claude structured output to categorize
			categorized, err := categorizeWithStructuredOutput(spendingNotes)
			if err != nil {
				log.Printf("AI categorization failed, falling back to keyword matching: %v", err)
				// Fallback to keyword matching
				return fallbackCategorization(spendingNotes), nil
			}

			return &core.ToolResult{
				Success: true,
				Data:    categorized,
			}, nil
		}).
		Build()
}

// categorizeWithStructuredOutput uses Claude's structured output to categorize transactions
func categorizeWithStructuredOutput(notes []string) (map[string]interface{}, error) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		return nil, fmt.Errorf("missing ANTHROPIC_API_KEY")
	}

	// Define structured output schema
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"categorized_transactions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"note": map[string]interface{}{
							"type": "string",
						},
						"category": map[string]interface{}{
							"type": "string",
							"enum": []string{"food", "travel", "subscription", "entertainment", "electronics", "miscellaneous"},
						},
					},
					"required": []string{"note", "category"},
				},
			},
		},
		"required": []string{"categorized_transactions"},
	}

	// Build prompt with transaction notes
	prompt := "Categorize the following transaction notes into one of these categories: food, travel, subscription, entertainment, electronics, or miscellaneous.\n\nTransaction notes:\n"
	for i, note := range notes {
		prompt += fmt.Sprintf("%d. %s\n", i+1, note)
	}

	// Call Claude API with structured output
	requestBody := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 2048,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]interface{}{
			"type":        "json_schema",
			"json_schema": schema,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", anthropicKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	if len(apiResponse.Content) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	// Parse structured output
	var structured struct {
		CategorizedTransactions []struct {
			Note     string `json:"note"`
			Category string `json:"category"`
		} `json:"categorized_transactions"`
	}

	if err := json.Unmarshal([]byte(apiResponse.Content[0].Text), &structured); err != nil {
		return nil, err
	}

	// Build result
	categories := map[string]int{
		"food":          0,
		"travel":        0,
		"subscription":  0,
		"entertainment": 0,
		"electronics":   0,
		"miscellaneous": 0,
	}

	var breakdown []string
	for _, item := range structured.CategorizedTransactions {
		categories[item.Category]++
		breakdown = append(breakdown, fmt.Sprintf("%s: %s", item.Note, item.Category))
	}

	return map[string]interface{}{
		"categories":     categories,
		"total_analyzed": len(structured.CategorizedTransactions),
		"breakdown":      breakdown,
	}, nil
}

// fallbackCategorization uses keyword matching when AI categorization fails
func fallbackCategorization(notes []string) *core.ToolResult {
	categories := map[string]int{
		"food":          0,
		"travel":        0,
		"subscription":  0,
		"entertainment": 0,
		"electronics":   0,
		"miscellaneous": 0,
	}

	var breakdown []string
	for _, note := range notes {
		category := categorizeNote(note)
		categories[category]++
		breakdown = append(breakdown, fmt.Sprintf("%s: %s", note, category))
	}

	return &core.ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"categories":     categories,
			"total_analyzed": len(notes),
			"breakdown":      breakdown,
		},
	}
}

// categorizeNote uses simple keyword matching to categorize transaction notes
func categorizeNote(note string) string {
	note = strings.ToLower(note)

	// Food keywords
	if strings.Contains(note, "food") || strings.Contains(note, "restaurant") ||
		strings.Contains(note, "cafe") || strings.Contains(note, "coffee") ||
		strings.Contains(note, "lunch") || strings.Contains(note, "dinner") ||
		strings.Contains(note, "grocery") || strings.Contains(note, "meal") {
		return "food"
	}

	// Travel keywords
	if strings.Contains(note, "uber") || strings.Contains(note, "lyft") ||
		strings.Contains(note, "flight") || strings.Contains(note, "hotel") ||
		strings.Contains(note, "gas") || strings.Contains(note, "parking") ||
		strings.Contains(note, "taxi") || strings.Contains(note, "bus") ||
		strings.Contains(note, "train") {
		return "travel"
	}

	// Subscription keywords
	if strings.Contains(note, "subscription") || strings.Contains(note, "netflix") ||
		strings.Contains(note, "spotify") || strings.Contains(note, "monthly") ||
		strings.Contains(note, "membership") || strings.Contains(note, "premium") {
		return "subscription"
	}

	// Entertainment keywords
	if strings.Contains(note, "movie") || strings.Contains(note, "concert") ||
		strings.Contains(note, "game") || strings.Contains(note, "entertainment") ||
		strings.Contains(note, "ticket") {
		return "entertainment"
	}

	// Electronics keywords
	if strings.Contains(note, "phone") || strings.Contains(note, "laptop") ||
		strings.Contains(note, "computer") || strings.Contains(note, "electronics") ||
		strings.Contains(note, "gadget") || strings.Contains(note, "tech") {
		return "electronics"
	}

	return "miscellaneous"
}

// ============================================================================
// HACKATHON IDEAS
// ============================================================================
// Here are some ideas for custom tools you could build:
//
// 1. SAVINGS GOAL TRACKER
//    - Track progress toward savings goals
//    - Calculate how long until goal is reached
//    - Suggest optimal deposit amounts
//
// 2. BUDGET ANALYZER
//    - Set spending limits by category
//    - Alert when approaching limits
//    - Compare actual vs. planned spending
//
// 3. RECURRING PAYMENT DETECTOR
//    - Identify subscription payments
//    - Warn about upcoming bills
//    - Suggest savings opportunities
//
// 4. CASH FLOW FORECASTER
//    - Predict future balance based on patterns
//    - Identify potential low balance periods
//    - Suggest when to save vs. spend
//
// 5. SMART SAVINGS ADVISOR
//    - Analyze spare cash available
//    - Recommend savings deposits
//    - Calculate interest projections
//
// 6. SPENDING INSIGHTS
//    - Categorize spending automatically
//    - Compare to typical user patterns
//    - Highlight unusual activity
//
// 7. FINANCIAL HEALTH SCORE
//    - Calculate overall financial wellness
//    - Track improvements over time
//    - Provide actionable recommendations
//
// 8. PEER COMPARISON (anonymous)
//    - Compare savings rate to anonymized peers
//    - Show percentile rankings
//    - Motivate better habits
//
// 9. TAX ESTIMATION
//    - Track potential tax obligations
//    - Suggest amounts to set aside
//    - Generate tax reports
//
// 10. EMERGENCY FUND BUILDER
//     - Calculate needed emergency fund size
//     - Track progress toward goal
//     - Suggest automated savings plan
//
// ============================================================================
