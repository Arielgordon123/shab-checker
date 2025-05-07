# SHAB Checker

A Go-based service that monitors and synchronizes data between Google Sheets, detects changes, and sends notifications via a Telegram service.

## Overview

SHAB Checker is designed to:
1. Compare data between two Google Spreadsheets
2. Detect and track changes in predefined cell ranges
3. Send notifications about changes to a configured service endpoint
4. Synchronize data between the spreadsheets

The application runs as a standalone service that can be deployed as a Docker container.

## Features

- **Spreadsheet Comparison**: Automatically detects differences between two Google Sheets
- **Selective Sheet Processing**: Configurable filtering to process only relevant sheets
- **Change Detection**: Identifies changes in predefined cell ranges
- **Change Notification**: Sends detected changes to a configured service endpoint
- **Spreadsheet Synchronization**: Keeps two spreadsheets in sync after processing


## Prerequisites

- Go 1.16 or later
- Google Sheets API credentials
- Access to the target Google Spreadsheets

## Configuration

The application uses a JSON configuration file with the following structure:

```json
{
  "spreadsheetIDs": {
    "sheet1": "YOUR_PRIMARY_SPREADSHEET_ID",
    "sheet2": "YOUR_SECONDARY_SPREADSHEET_ID"
  },
  "credentialsFile": "/path/to/your/credentials.json",
  "preDefinedCells": [
    {
      "cellRange": "A1:Z100",
      "titleRange": "A1",
      "timeRange": "B1"
    }
  ],
  "tgServiceURL": "http://your-telegram-service/endpoint"
}
```

### Configuration Options

- **spreadsheetIDs**: IDs of the primary and secondary Google Spreadsheets to compare and sync
- **credentialsFile**: Path to the Google API credentials JSON file
- **preDefinedCells**: Array of cell configurations to monitor for changes
  - **cellRange**: Range of cells to monitor (e.g., "A1:Z100")
  - **titleRange**: Cell containing the title for the range
  - **timeRange**: Cell containing the timestamp for the range
- **tgServiceURL**: URL endpoint of the notification service

## Installation

### Using Docker

1. Clone the repository
2. Place your Google API credentials in the appropriate location
3. Configure the application (see Configuration section)
4. Build and run the Docker container:

```bash
docker build -t shab-checker .
docker run -v /path/to/config:/root/config -v /path/to/logs:/root/logs shab-checker
```

### Manual Installation

1. Clone the repository
2. Install dependencies:

```bash
go mod download
```

3. Build the application:

```bash
go build -o shab-checker
```

4. Run the application:

```bash
./shab-checker
```

## Environment Variables

- **CONFIG_PATH**: Path to the configuration file (default: `/root/config/config.json`)

## Project Structure

- **main.go**: Entry point and main application logic
- **config/**: Configuration handling
- **pkg/client/**: HTTP client for sending notifications
- **pkg/diff/**: Spreadsheet comparison logic
- **pkg/sheets/**: Google Sheets API integration

## Logging

The application logs to stdout and to log files which are rotated according to the configuration in `logrotate.conf`.

## License

[Specify your license here]

## Contributing

[Add contribution guidelines if applicable]