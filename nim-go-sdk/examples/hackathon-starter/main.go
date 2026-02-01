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
	"path/filepath"
	"sort"
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
	srv.AddTool(createChartGeneratorTool(liminalExecutor))
	
	// ============================================================================
	// INITIALIZE LANGGRAPH ORCHESTRATOR
	// ============================================================================
	// Create the graph workflow and add it as a tool
	srv.AddTool(createGraphOrchestratorTool(liminalExecutor))
	log.Println("✅ Added custom tools with LangGraph orchestrator")

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

	// Create charts directory if it doesn't exist
	chartsDir := filepath.Join(".", "charts")
	if err := os.MkdirAll(chartsDir, 0755); err != nil {
		log.Printf("Warning: Could not create charts directory: %v", err)
	}

	// Serve static chart files
	http.HandleFunc("/charts/", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "image/svg+xml")
		
		filename := filepath.Base(r.URL.Path)
		filepath := filepath.Join(chartsDir, filename)
		
		http.ServeFile(w, r, filepath)
	})

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🚀 Hackathon Starter Server Running")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📡 WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Printf("📊 Charts: http://localhost:%s/charts/", port)
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

IMPORTANT - REQUEST ROUTING:
For EVERY user request, you MUST first call the route_request tool with their message. This orchestrator will analyze their intent and guide you on the best way to help them. The orchestrator routes requests to three specialized modes:
1. General Inquiry - standard banking queries
2. Image Payment - receipt splitting and image-based payments
3. Financial Help - financial advice, budgeting, saving guidance

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
- Route request through orchestrator (route_request) - CALL THIS FIRST!
- Analyze spending patterns (analyze_spending)
- Set weekly spending goal (spend_weekly_goal) - requires confirmation
- Check weekly spending progress (get_weekly_spending_progress)
- Quick check weekly spend status (check_weeklyspend) - use this for context
- Categorize spending by transaction notes (categorize_transactions)
- Generate balance trend chart (generate_chart) - Shows account balance over time

IMPORTANT - BALANCE TREND CHART:
When a user asks for a chart, graph, visualization, trend, or wants to see their balance over time:
1. ALWAYS call the generate_chart tool with: chart_type='line', data_type='balance_trend', days=30 (or user's requested timeframe)
2. The tool will return an 'image_url' containing a base64-encoded SVG chart
3. Display the chart directly in your response using markdown image syntax:

![Balance Trend Chart](IMAGE_URL_FROM_TOOL)

Replace IMAGE_URL_FROM_TOOL with the actual 'image_url' value from the generate_chart tool result.
4. Explain what the chart shows - their account balance trend over time based on transaction history
5. The chart image is also saved to a temp file (the tool returns 'file_path' with location)

Example response after calling generate_chart:
"Here's your account balance trend over the last 30 days:

![Balance Trend Chart](data:image/svg+xml;base64,PHN2Zy4uLg==)

The chart shows how your balance has changed over time based on your transaction history. Your current balance is $1,234.56."

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
				log.Printf("AI categorization failed, using fallback: %v", err)
				// Fallback to keyword matching
				categorized = fallbackCategorization(spendingNotes)
			}

			return &core.ToolResult{
				Success: true,
				Data:    categorized,
			}, nil
		}).
		Build()
}

// fallbackCategorization uses keyword matching when AI categorization fails
func fallbackCategorization(notes []string) map[string]interface{} {
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

	return map[string]interface{}{
		"categories":     categories,
		"total_analyzed": len(notes),
		"breakdown":      breakdown,
	}
}

// categorizeNote uses simple keyword matching to categorize transaction notes
func categorizeNote(note string) string {
	note = strings.ToLower(note)

	// Food keywords
	if strings.Contains(note, "food") || strings.Contains(note, "restaurant") ||
		strings.Contains(note, "cafe") || strings.Contains(note, "coffee") ||
		strings.Contains(note, "lunch") || strings.Contains(note, "dinner") ||
		strings.Contains(note, "grocery") || strings.Contains(note, "meal") ||
		strings.Contains(note, "hulu") {
		return "food"
	}

	// Travel keywords
	if strings.Contains(note, "uber") || strings.Contains(note, "lyft") ||
		strings.Contains(note, "flight") || strings.Contains(note, "hotel") ||
		strings.Contains(note, "gas") || strings.Contains(note, "parking") ||
		strings.Contains(note, "taxi") || strings.Contains(note, "bus") ||
		strings.Contains(note, "train") || strings.Contains(note, "ticket") {
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
		strings.Contains(note, "video") {
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

// generateBalanceTrendFromTransactions creates a line chart showing balance over time
func generateBalanceTrendFromTransactions(transactions []map[string]interface{}, currentBalance float64, days int) map[string]interface{} {
	log.Printf("📊 Generating balance trend from %d transactions, current balance: %.2f", len(transactions), currentBalance)

	if len(transactions) == 0 {
		return map[string]interface{}{
			"labels": []string{"Today"},
			"values": []float64{currentBalance},
			"title":  "Account Balance Trend",
		}
	}

	// Sort transactions by date (oldest first)
	sortedTxs := make([]map[string]interface{}, len(transactions))
	copy(sortedTxs, transactions)
	
	sort.Slice(sortedTxs, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, sortedTxs[i]["createdAt"].(string))
		timeJ, _ := time.Parse(time.RFC3339, sortedTxs[j]["createdAt"].(string))
		return timeI.Before(timeJ)
	})

	// Calculate starting balance by working backwards from current balance
	runningBalance := currentBalance
	for i := len(sortedTxs) - 1; i >= 0; i-- {
		tx := sortedTxs[i]
		direction, _ := tx["direction"].(string)
		
		var amount float64
		if amountStr, ok := tx["amount"].(string); ok {
			fmt.Sscanf(amountStr, "%f", &amount)
		} else if amountFloat, ok := tx["amount"].(float64); ok {
			amount = amountFloat
		}

		// Reverse the transaction to get starting balance
		if direction == "credit" {
			runningBalance -= amount // Was a credit, so subtract to go back
		} else if direction == "debit" {
			runningBalance += amount // Was a debit, so add to go back
		}
	}

	startingBalance := runningBalance
	log.Printf("📊 Calculated starting balance: %.2f", startingBalance)

	// Now go forward through transactions building the balance timeline
	labels := []string{}
	values := []float64{}
	
	// Add starting point
	if len(sortedTxs) > 0 {
		firstTxTime, _ := time.Parse(time.RFC3339, sortedTxs[0]["createdAt"].(string))
		labels = append(labels, firstTxTime.Format("Jan 2"))
		values = append(values, startingBalance)
	}

	runningBalance = startingBalance
	for _, tx := range sortedTxs {
		direction, _ := tx["direction"].(string)
		
		var amount float64
		if amountStr, ok := tx["amount"].(string); ok {
			fmt.Sscanf(amountStr, "%f", &amount)
		} else if amountFloat, ok := tx["amount"].(float64); ok {
			amount = amountFloat
		}

		// Apply transaction
		if direction == "credit" {
			runningBalance += amount
		} else if direction == "debit" {
			runningBalance -= amount
		}

		txTime, _ := time.Parse(time.RFC3339, tx["createdAt"].(string))
		labels = append(labels, txTime.Format("Jan 2"))
		values = append(values, runningBalance)
		
		log.Printf("📊 %s: %s $%.2f -> Balance: $%.2f", txTime.Format("Jan 2"), direction, amount, runningBalance)
	}

	log.Printf("📊 Final chart: %d data points", len(labels))

	return map[string]interface{}{
		"labels": labels,
		"values": values,
		"title":  "Account Balance Trend",
	}
}

// generateSVGChart creates an SVG image from chart data
func generateSVGChart(chartData map[string]interface{}) string {
	labels, _ := chartData["labels"].([]string)
	valuesInterface, _ := chartData["values"].([]float64)
	title, _ := chartData["title"].(string)

	if len(labels) == 0 || len(valuesInterface) == 0 {
		return `<svg width="600" height="400" xmlns="http://www.w3.org/2000/svg"><text x="300" y="200" text-anchor="middle" fill="#666">No data available</text></svg>`
	}

	// Chart dimensions
	width := 800
	height := 500
	padding := 80
	chartWidth := width - padding*2
	chartHeight := height - padding*2

	// Find min/max values
	minValue := valuesInterface[0]
	maxValue := valuesInterface[0]
	for _, v := range valuesInterface {
		if v < minValue {
			minValue = v
		}
		if v > maxValue {
			maxValue = v
		}
	}
	valueRange := maxValue - minValue
	if valueRange == 0 {
		valueRange = 1
	}

	// Build SVG
	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	
	// Background
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#ffffff"/>`, width, height))
	
	// Title
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="30" text-anchor="middle" font-size="20" font-weight="bold" fill="#333">%s</text>`, width/2, title))
	
	// Grid lines and Y-axis labels
	for i := 0; i <= 4; i++ {
		y := float64(padding) + float64(chartHeight*i)/4
		gridValue := maxValue - (float64(i)/4)*valueRange
		svg.WriteString(fmt.Sprintf(`<line x1="%d" y1="%.1f" x2="%d" y2="%.1f" stroke="#e0e0e0" stroke-width="1"/>`, padding, y, width-padding, y))
		svg.WriteString(fmt.Sprintf(`<text x="%d" y="%.1f" text-anchor="end" font-size="12" fill="#666">$%.0f</text>`, padding-10, y+4, gridValue))
	}
	
	// Build points and line path
	var points []string
	for i, value := range valuesInterface {
		x := float64(padding) + (float64(i)/float64(len(valuesInterface)-1))*float64(chartWidth)
		y := float64(padding) + float64(chartHeight) - ((value-minValue)/valueRange)*float64(chartHeight)
		points = append(points, fmt.Sprintf("%.1f,%.1f", x, y))
	}
	
	// Draw line
	svg.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="#4ECDC4" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>`, strings.Join(points, " ")))
	
	// Draw points and labels
	labelStep := 1
	if len(labels) > 15 {
		labelStep = len(labels) / 10
	}
	
	for i, value := range valuesInterface {
		x := float64(padding) + (float64(i)/float64(len(valuesInterface)-1))*float64(chartWidth)
		y := float64(padding) + float64(chartHeight) - ((value-minValue)/valueRange)*float64(chartHeight)
		
		// Draw point
		svg.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="4" fill="#4ECDC4" stroke="white" stroke-width="2"/>`, x, y))
		
		// Draw label (only for selected points to avoid crowding)
		if i%labelStep == 0 || i == len(labels)-1 {
			svg.WriteString(fmt.Sprintf(`<text x="%.1f" y="%d" text-anchor="middle" font-size="10" fill="#666" transform="rotate(-45 %.1f %d)">%s</text>`, x, height-padding+20, x, height-padding+20, labels[i]))
		}
	}
	
	svg.WriteString(`</svg>`)
	return svg.String()
}

// ============================================================================
// CUSTOM TOOL: GRAPH ORCHESTRATOR
// ============================================================================
// Routes requests through the LangGraph workflow system

func createGraphOrchestratorTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("route_request").
		Description("Analyze user's request and route to the appropriate specialized handler. Call this FIRST for any user request to determine the best way to help them. Returns routing decision and context.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"user_message": tools.StringProperty("The user's original message/request"),
		}, "user_message")).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				UserMessage string `json:"user_message"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Create graph and initialize state
			graph := CreateFinancialAgentGraph(liminalExecutor)
			graph.State = &GraphState{
				Messages:     []string{},
				UserID:       toolParams.UserID,
				Conversation: map[string]interface{}{
					"user_input": params.UserMessage,
				},
			}

			// Execute graph workflow
			if err := graph.Execute(ctx, toolParams.UserID); err != nil {
				log.Printf("Graph execution error: %v", err)
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("routing failed: %v", err),
				}, nil
			}

			// Extract routing decision
			route, _ := graph.State.Conversation["route"].(string)
			handlerType, _ := graph.State.Conversation["handler_type"].(string)

			var guidance string
			switch handlerType {
			case "image_payment":
				guidance = "User wants to split a payment or send money based on an image/receipt. You should ask for the receipt image, analyze it, calculate splits, and use send_money tool to process payments."
			case "financial_help":
				guidance = "User needs financial guidance and advice. Use check_weeklyspend, analyze_spending, and categorize_transactions to provide comprehensive financial insights. Give actionable recommendations for saving and budgeting."
			case "general":
				// Check if chart was requested
				if chartRequested, ok := graph.State.Conversation["chart_requested"].(bool); ok && chartRequested {
					guidance = "User requested a balance trend chart. You MUST call the generate_chart tool with these parameters:\n- chart_type: 'line'\n- data_type: 'balance_trend'\n- days: 30 (or ask the user for a timeframe)\n\nAfter calling generate_chart, you will receive an 'image_url' field with a base64-encoded SVG. Display it in your response using markdown:\n\n![Balance Trend Chart](image_url_here)\n\nReplace 'image_url_here' with the actual image_url from the tool result. Explain that this shows their account balance over time based on transaction history."
				} else {
					guidance = "Standard query. Use appropriate banking tools (get_balance, get_transactions, etc.) to help the user."
				}
			default:
				guidance = "Process as a general query."
			}

			result := map[string]interface{}{
				"route":        route,
				"handler_type": handlerType,
				"guidance":     guidance,
				"user_message": params.UserMessage,
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: CHART GENERATOR
// ============================================================================
// Generates charts and graphs from financial data using go-chart library
// Similar to matplotlib in Python

func createChartGeneratorTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("generate_chart").
		Description("Generate a line chart showing account balance trend over time. Calculates running balance from transaction history in chronological order.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"chart_type": tools.StringProperty("Type of chart: always 'line' for balance trend"),
			"data_type":  tools.StringProperty("What to visualize: always 'balance_trend'"),
			"days":       tools.IntegerProperty("Number of days of data to include (default: 30)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				ChartType string `json:"chart_type"`
				DataType  string `json:"data_type"`
				Days      int    `json:"days"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Set defaults
			if params.Days == 0 {
				params.Days = 30
			}

			// Fetch transaction data
			txRequest := map[string]interface{}{"limit": 200}
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

			// Get current balance
			balanceResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_balance",
				Input:     []byte("{}"),
				RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to fetch balance: %v", err),
				}, nil
			}

			// Parse current balance
			var currentBalance float64
			if balanceResponse.Success {
				var balanceData map[string]interface{}
				if err := json.Unmarshal(balanceResponse.Data, &balanceData); err == nil {
					if balanceStr, ok := balanceData["balance"].(string); ok {
						fmt.Sscanf(balanceStr, "%f", &currentBalance)
					} else if balanceFloat, ok := balanceData["balance"].(float64); ok {
						currentBalance = balanceFloat
					}
				}
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

			// Generate balance trend chart
			chartData := generateBalanceTrendFromTransactions(transactions, currentBalance, params.Days)

			// Generate SVG image
			svgContent := generateSVGChart(chartData)
			
			// Save to charts directory
			chartsDir := filepath.Join(".", "charts")
			timestamp := time.Now().Format("20060102-150405")
			filename := fmt.Sprintf("balance-trend-%s.svg", timestamp)
			filePath := filepath.Join(chartsDir, filename)
			
			if err := os.WriteFile(filePath, []byte(svgContent), 0644); err != nil {
				log.Printf("Failed to save chart to file: %v", err)
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to save chart: %v", err),
				}, nil
			}
			
			log.Printf("Chart saved to: %s", filePath)

			// Get server port from environment
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}
			
			// Create HTTP URL for the chart
			chartURL := fmt.Sprintf("http://localhost:%s/charts/%s", port, filename)

			result := map[string]interface{}{
				"chart_type":   "line",
				"data_type":    "balance_trend",
				"image_url":    chartURL,
				"file_path":    filePath,
				"total_points": len(transactions),
				"message":      fmt.Sprintf("Generated balance trend chart with %d data points. View at: %s", len(transactions), chartURL),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
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

// ============================================================================
// LANGGRAPH-STYLE WORKFLOW SYSTEM
// ============================================================================
// Stateful workflow graph for building complex agent workflows

// Node represents a single step in the agent workflow graph
type Node struct {
	Name    string
	Handler func(ctx context.Context, state *GraphState) error
}

// GraphState holds the current state as the graph executes
type GraphState struct {
	Messages     []string
	CurrentTool  string
	ToolResult   interface{}
	UserID       string
	Conversation map[string]interface{}
	Error        error
}

// Graph represents a stateful workflow for the agent
type Graph struct {
	Nodes       map[string]*Node
	Edges       map[string][]string // node -> next possible nodes
	StartNode   string
	CurrentNode string
	State       *GraphState
}

// NewGraph creates a new agent workflow graph
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]string),
		State: &GraphState{
			Messages:     []string{},
			Conversation: make(map[string]interface{}),
		},
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(name string, handler func(ctx context.Context, state *GraphState) error) {
	g.Nodes[name] = &Node{
		Name:    name,
		Handler: handler,
	}
}

// AddEdge adds a directed edge from one node to another
func (g *Graph) AddEdge(from, to string) {
	g.Edges[from] = append(g.Edges[from], to)
}

// SetStart sets the starting node
func (g *Graph) SetStart(nodeName string) {
	g.StartNode = nodeName
	g.CurrentNode = nodeName
}

// Execute runs the graph starting from the start node
func (g *Graph) Execute(ctx context.Context, userID string) error {
	g.State.UserID = userID
	g.CurrentNode = g.StartNode

	visited := make(map[string]bool)

	for g.CurrentNode != "" {
		// Prevent infinite loops
		if visited[g.CurrentNode] {
			return fmt.Errorf("cycle detected at node: %s", g.CurrentNode)
		}
		visited[g.CurrentNode] = true

		node, exists := g.Nodes[g.CurrentNode]
		if !exists {
			return fmt.Errorf("node not found: %s", g.CurrentNode)
		}

		log.Printf("Executing node: %s", node.Name)

		// Execute the node handler
		if err := node.Handler(ctx, g.State); err != nil {
			g.State.Error = err
			return err
		}

		// Move to next node
		nextNodes := g.Edges[g.CurrentNode]
		if len(nextNodes) == 0 {
			// End of graph
			break
		}

		// Simple routing: take the first edge
		// In a real implementation, you'd use conditional routing
		g.CurrentNode = nextNodes[0]
	}

	return nil
}

// ConditionalRouter determines which node to execute next based on state
func (g *Graph) ConditionalRouter(condition func(state *GraphState) string) {
	// This would be used to route to different nodes based on conditions
	// Example: if balance < 0, go to "overdraft_warning" node
	// else go to "normal_response" node
}

// CreateFinancialAgentGraph creates an example financial agent workflow graph
func CreateFinancialAgentGraph(liminalExecutor core.ToolExecutor) *Graph {
	graph := NewGraph()

	// Root Node: Orchestrator - coordinates the entire workflow
	graph.AddNode("orchestrator", func(ctx context.Context, state *GraphState) error {
		log.Println("Orchestrator: Analyzing request and routing to appropriate handler...")
		// The orchestrator analyzes the user's request and routes to the correct first-layer node
		// In a real implementation, this would use Claude to classify the request type
		
		// Example routing logic (would be replaced with actual classification)
		userInput := state.Conversation["user_input"].(string)
		
		if strings.Contains(userInput, "image") || strings.Contains(userInput, "receipt") || strings.Contains(userInput, "split") {
			state.Conversation["route"] = "image_payment"
			log.Println("💙💙💙💙💙💙💙💙💙💙💙💙💙")
		} else if strings.Contains(userInput, "broke") || strings.Contains(userInput, "help") || 
			strings.Contains(userInput, "stats") || strings.Contains(userInput, "investing") || 
			strings.Contains(userInput, "saving") {
			state.Conversation["route"] = "financial_help"
			log.Println("🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴")
			log.Println("🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴")
			log.Println("🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴🔴")
		} else {
			state.Conversation["route"] = "general_inquiry"
			log.Println("🍏🍏🍏🍏🍏🍏🍏🍏🍏🍏🍏🍏🍏")
		}
		
		return nil
	})

	// FIRST LAYER NODES - Three specialized handlers

	// Node 1: General Inquiry - handles standard queries using model's default functions
	graph.AddNode("general_inquiry", func(ctx context.Context, state *GraphState) error {
		log.Println("General Inquiry: Processing standard request...")
		// Use Claude's default capabilities for general questions
		// Check balance, view transactions, search users, etc.
		state.Conversation["handler_type"] = "general"
		state.Conversation["intent"] = "standard_query"
		
		// Check if user is asking for a chart/graph/balance trend
		userInput := state.Conversation["user_input"].(string)
		lowerInput := strings.ToLower(userInput)
		
		if strings.Contains(lowerInput, "chart") || strings.Contains(lowerInput, "graph") || 
			strings.Contains(lowerInput, "visualize") || strings.Contains(lowerInput, "plot") ||
			strings.Contains(lowerInput, "trend") || strings.Contains(lowerInput, "balance") && (strings.Contains(lowerInput, "show") || strings.Contains(lowerInput, "see")) {
			log.Println("📊 Detected balance trend chart request")
			state.Conversation["chart_requested"] = true
		}
		
		return nil
	})

	// Node 2: Image Payment - handles receipt splitting and image-based payments
	graph.AddNode("image_payment", func(ctx context.Context, state *GraphState) error {
		log.Println("Image Payment: Processing receipt/image for payment splitting...")
		// Use Claude vision to analyze receipt images
		// Extract amounts, calculate splits, identify friends to pay
		state.Conversation["handler_type"] = "image_payment"
		state.CurrentTool = "process_receipt_image"
		
		// Example: Parse receipt and determine split amounts
		state.ToolResult = map[string]interface{}{
			"total_amount": 50.00,
			"split_count":  2,
			"per_person":   25.00,
			"currency":     "USD",
			"recipients":   []string{"@alice", "@bob"},
		}
		return nil
	})

	// Node 3: Financial Help (Get Stats) - provides financial insights and recommendations
	graph.AddNode("financial_help", func(ctx context.Context, state *GraphState) error {
		log.Println("Financial Help: Analyzing financial situation and providing guidance...")
		// Analyze user's financial health
		// Provide saving/investing recommendations
		// Show spending patterns, suggest improvements
		state.Conversation["handler_type"] = "financial_help"
		state.CurrentTool = "analyze_financial_health"
		
		// Example: Generate financial insights
		state.ToolResult = map[string]interface{}{
			"spending_velocity": "high",
			"savings_rate":      "low",
			"recommendations": []string{
				"Set up automatic savings of $50/week",
				"Reduce dining out expenses by 20%",
				"Consider moving $500 to savings vault for 4% APY",
			},
		}
		return nil
	})

	// SECOND LAYER NODES - Processing nodes

	// Execute tool based on first layer routing
	graph.AddNode("execute_tool", func(ctx context.Context, state *GraphState) error {
		log.Println("Executing tool based on handler type...")
		handlerType := state.Conversation["handler_type"].(string)
		
		switch handlerType {
		case "image_payment":
			// Process payment splitting
			log.Println("Processing payment splits from receipt...")
		case "financial_help":
			// Fetch and analyze financial data
			log.Println("Fetching financial stats and generating recommendations...")
		case "general":
			// Handle standard queries
			log.Println("Processing general query...")
		}
		
		return nil
	})

	// Generate final response
	graph.AddNode("generate_response", func(ctx context.Context, state *GraphState) error {
		log.Println("Generating final response...")
		handlerType := state.Conversation["handler_type"].(string)
		
		var response string
		switch handlerType {
		case "image_payment":
			result := state.ToolResult.(map[string]interface{})
			response = fmt.Sprintf("I analyzed the receipt. Total: $%.2f. I can split this %d ways ($%.2f each) and send to %v. Ready to proceed?",
				result["total_amount"], result["split_count"], result["per_person"], result["recipients"])
		case "financial_help":
			result := state.ToolResult.(map[string]interface{})
			recommendations := result["recommendations"].([]string)
			response = fmt.Sprintf("I analyzed your finances. Here's what I found:\n- Spending: %s\n- Savings: %s\n\nRecommendations:\n",
				result["spending_velocity"], result["savings_rate"])
			for _, rec := range recommendations {
				response += fmt.Sprintf("• %s\n", rec)
			}
		case "general":
			response = "I can help you with that. Let me check your account..."
		}
		
		state.Messages = append(state.Messages, response)
		return nil
	})

	// Define the workflow edges
	// Orchestrator routes to one of three first-layer nodes
	graph.AddEdge("orchestrator", "general_inquiry")
	graph.AddEdge("orchestrator", "image_payment")
	graph.AddEdge("orchestrator", "financial_help")
	
	// All first-layer nodes converge to execute_tool
	graph.AddEdge("general_inquiry", "execute_tool")
	graph.AddEdge("image_payment", "execute_tool")
	graph.AddEdge("financial_help", "execute_tool")
	
	// Execute tool leads to response generation
	graph.AddEdge("execute_tool", "generate_response")

	// Set orchestrator as the starting point
	graph.SetStart("orchestrator")

	return graph
}

// RunGraphExample demonstrates how to use the graph workflow
func RunGraphExample(liminalExecutor core.ToolExecutor) {
	graph := CreateFinancialAgentGraph(liminalExecutor)

	ctx := context.Background()
	userID := "example-user-123"

	if err := graph.Execute(ctx, userID); err != nil {
		log.Printf("Graph execution failed: %v", err)
		return
	}

	log.Println("Graph execution completed successfully")
	log.Printf("Final state messages: %v", graph.State.Messages)
}
