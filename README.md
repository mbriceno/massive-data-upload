# Massive Data Upload Tool

## Overview
This project is a high-performance CLI utility written in Go, designed to import large datasets from Excel files into a PostgreSQL database. It is built to handle "massive" files by using a streaming approach that avoids loading entire spreadsheets into memory, coupled with a concurrent worker pool for efficient database persistence.

## Architecture
The application follows a **Producer-Consumer** pattern using Go's concurrency primitives:

1.  **Reader Engine (Producer)**: 
    - Uses the `excelize` library to stream rows from the Excel file one by one.
    - For each row, it identifies the corresponding `TabProcessor` based on the sheet name.
    - The processor validates and transforms the raw string data into GORM domain models.
    - Rows are grouped into batches (default: 500) and sent to a buffered channel.

2.  **Worker Pool (Consumers)**:
    - A configurable number of worker goroutines listen to the batch channel.
    - Each worker takes a batch and uses the appropriate processor to perform a bulk `INSERT` via GORM.
    - This decouples the file parsing (I/O and CPU bound) from the database insertion (Network and DB bound).

3.  **Processors**:
    - Implements the `TabProcessor` interface.
    - Allows for custom validation logic and data mapping per sheet.
    - Processors includes a local cache to resolve administrative entity IDs, significantly reducing redundant database queries.

## Key Features
- **Streaming Parsing**: Processes files of virtually any size without RAM exhaustion.
- **Concurrent Insertion**: Parallelized database writes through a worker pool.
- **Batching**: Optimized GORM bulk inserts to minimize database roundtrips.
- **Foreign Key Resolution**: Smart caching during the import process to handle relational data efficiently.
- **Graceful Shutdown**: Context-aware design to handle cancellations.

## Project Structure
- `cmd/processor-cli/`: CLI entry point and orchestration.
- `internal/config/`: Environment-based configuration management.
- `internal/domain/`: GORM models and schema definitions.
- `internal/excel/`: Logic for streaming rows from Excel.
- `internal/importer/`: Core worker pool logic and processing interfaces.
- `internal/database/`: PostgreSQL connection pooling and client setup.

## Getting Started

### Prerequisites
- Go 1.21 or higher.
- A running PostgreSQL instance.

### Configuration
The application reads configuration from environment variables or a `.env` file in the root directory.

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `secret` |
| `DB_NAME` | Database name | `mi_empresa` |
| `DB_PORT` | Database port | `5432` |
| `NUM_WORKERS` | Number of concurrent DB writers | `4` |

### Installation
```bash
go mod tidy
```

### Building

```bash
go build -o bin/massive-data-upload cmd/processor-cli/main.go
```

### Usage
Run the CLI by pointing to an Excel file using the `--file` flag:

```bash
# Run directly with go
go run cmd/processor-cli/main.go --file=your_data_file.xlsx

# Run builded binary
./bin/massive-data-upload --file your_data_file.xlsx
```
