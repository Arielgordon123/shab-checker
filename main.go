package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"shab-checker/config"
	"shab-checker/pkg/client"
	"shab-checker/pkg/diff"

	"shab-checker/pkg/sheets"

	sheetsClient "google.golang.org/api/sheets/v4"
)

// Application represents the main application structure
type Application struct {
	ctx           context.Context
	config        config.Config
	sheetsService *sheetsClient.Service
	client        *client.Client
	logger        *log.Logger
}

// NewApplication creates and initializes a new Application
func NewApplication(ctx context.Context, configPath string) (*Application, error) {
	logger := log.New(os.Stdout, "[SHAB-CHECKER] ", log.LstdFlags|log.Lshortfile)

	// Load configuration from file
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	credentials, err := config.LoadCredentials(cfg.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	// Initialize the Google Sheets API client
	sheetsService, err := sheets.NewSheetsClient(ctx, credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets client: %w", err)
	}

	// Initialize HTTP client for sending changes
	apiClient := client.NewClient(cfg.TgServiceURL)

	return &Application{
		ctx:           ctx,
		config:        cfg,
		sheetsService: sheetsService,
		client:        apiClient,
		logger:        logger,
	}, nil
}

// Run executes the main application logic
func (app *Application) Run() error {
	app.logger.Println("Starting SHAB Checker service")

	// Spreadsheet IDs for the two spreadsheets to compare
	spreadsheetID1 := app.config.SpreadsheetIDs.Sheet1
	spreadsheetID2 := app.config.SpreadsheetIDs.Sheet2

	// Get sheet names from the first spreadsheet
	sheetNames, err := sheets.GetSheetNames(app.ctx, app.sheetsService, spreadsheetID1)
	if err != nil {
		return fmt.Errorf("failed to get sheet names: %w", err)
	}

	for i, sheetName := range sheetNames {
		// Skip irrelevant sheets based on filtering criteria
		if !app.shouldProcessSheet(sheetName, i) {
			app.logger.Printf("Skipping sheet: %s", sheetName)
			continue
		}

		app.logger.Printf("Processing sheet: %s", sheetName)
		if err := app.processSheet(spreadsheetID1, spreadsheetID2, sheetName); err != nil {
			app.logger.Printf("Error processing sheet %s: %v", sheetName, err)
			// Continue with next sheet even if there was an error with this one
		}
	}

	app.logger.Println("SHAB Checker service completed successfully")
	return nil
}

// shouldProcessSheet determines if a sheet should be processed based on criteria
func (app *Application) shouldProcessSheet(sheetName string, index int) bool {
	// Skip sheets that match certain criteria
	return sheetName != "מספרים אישיים" &&
		!strings.Contains(sheetName, "עותק") &&
		!strings.Contains(sheetName, "שלד") &&
		index <= 4
}

// processSheet handles the comparison and synchronization of a single sheet
func (app *Application) processSheet(spreadsheetID1, spreadsheetID2, sheetName string) error {
	// Read data from the first sheet
	sheet1, err := sheets.ReadSpreadsheet(app.sheetsService, spreadsheetID1, sheetName)
	if err != nil {
		return fmt.Errorf("failed to read spreadsheet 1: %w", err)
	}

	// Try to read data from the second sheet
	sheet2, err := sheets.ReadSpreadsheet(app.sheetsService, spreadsheetID2, sheetName)
	if err != nil {
		app.logger.Printf("Sheet %s not found in second spreadsheet, creating it", sheetName)

		// Create a new sheet with the same name in the second spreadsheet
		_, err = sheets.AddSheet(app.sheetsService, spreadsheetID2, sheetName, true)
		if err != nil {
			return fmt.Errorf("failed to add a new sheet: %w", err)
		}

		// Read the newly created sheet
		sheet2, err = sheets.ReadSpreadsheet(app.sheetsService, spreadsheetID2, sheetName)
		if err != nil {
			return fmt.Errorf("failed to read the newly created sheet: %w", err)
		}
	}

	// Compare the spreadsheets and find differences
	changedCells := diff.CompareSpreadsheetsAndGetDiff(sheet1, sheet2, sheetName, app.config.PreDefinedCells)

	// Handle changes if any were found
	if len(changedCells) > 0 {
		if err := app.handleChanges(sheetName, changedCells); err != nil {
			app.logger.Printf("Warning: Error handling changes: %v", err)
		}
	} else {
		app.logger.Printf("No changes detected in sheet: %s", sheetName)
	}

	// Sync data from first spreadsheet to second
	if err := app.syncSheets(spreadsheetID2, sheetName, sheet1); err != nil {
		return err
	}

	return nil
}

// handleChanges processes and reports detected changes
func (app *Application) handleChanges(sheetName string, changedCells []diff.ChangedCell) error {
	app.logger.Printf("Found %d changes in sheet: %s", len(changedCells), sheetName)

	// Send changes to the notification service
	err := app.client.SendChanges(sheetName, changedCells)
	if err != nil {
		return fmt.Errorf("failed to send changes: %w", err)
	}

	app.logger.Println("Changes sent successfully")

	// Log detailed changes for debugging
	for _, cell := range changedCells {
		app.logger.Printf("Change - title: %v, time: %v, Value: %v, Old value: %v",
			cell.Title, cell.Time, cell.Value, cell.OldValue)
	}

	return nil
}

// syncSheets synchronizes data from first spreadsheet to second
func (app *Application) syncSheets(spreadsheetID2, sheetName string, sheet1 [][]interface{}) error {
	// Clear the existing data in the second spreadsheet
	if err := sheets.ClearSheet(app.sheetsService, spreadsheetID2, sheetName); err != nil {
		return fmt.Errorf("failed to clear the sheet: %w", err)
	}

	// Write the new data to the second spreadsheet
	if err := sheets.WriteToSheet(app.sheetsService, spreadsheetID2, sheetName, sheet1); err != nil {
		return fmt.Errorf("failed to write to the sheet: %w", err)
	}

	app.logger.Printf("Successfully synchronized sheet: %s", sheetName)
	return nil
}

func main() {
	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Printf("Received signal: %v\n", sig)
		cancel()
	}()

	// Use environment-specific config path or default
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/root/config/config.json"
	}

	// Initialize the application
	app, err := NewApplication(ctx, configPath)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
