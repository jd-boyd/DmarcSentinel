# DMARC Report Viewer - Task Breakdown for AI Agent

This document breaks down the DMARC Report Viewer project into discrete, sequential tasks suitable for a non-thinking AI agent. Each task includes what should be implemented, what files to create/modify, and specific testable outcomes.

---

## TASK 1: Project Initialization and Configuration Module ✅ COMPLETED

**Completion Date:** 2025-11-25
**Status:** All objectives met, tests passing, configuration priority working correctly

### Objectives
- Set up Go module structure
- Implement configuration loading (YAML, environment variables, CLI flags)
- Create basic project skeleton

### Files to Create
1. `go.mod` - Initialize Go module named `dmarc-viewer`
2. `config.yaml.example` - Example configuration file with all options documented
3. `internal/config/config.go` - Configuration struct and loading logic
4. `internal/config/config_test.go` - Unit tests for configuration loading
5. `cmd/dmarc-viewer/main.go` - Basic main function that loads config and prints it

### Implementation Requirements

**Configuration Struct** (`internal/config/config.go`):
```go
type Config struct {
    IMAP     IMAPConfig     `yaml:"imap"`
    Database DatabaseConfig `yaml:"database"`
    Web      WebConfig      `yaml:"web"`
    Sync     SyncConfig     `yaml:"sync"`
    Logging  LogConfig      `yaml:"logging"`
}

type IMAPConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    Folder   string `yaml:"folder"`
    UseTLS   bool   `yaml:"use_tls"`
}

type DatabaseConfig struct {
    Path string `yaml:"path"`
}

type WebConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

type SyncConfig struct {
    Interval  string `yaml:"interval"` // e.g., "15m"
    OnStartup bool   `yaml:"on_startup"`
}

type LogConfig struct {
    Level  string `yaml:"level"`  // debug, info, warn, error
    Format string `yaml:"format"` // json, text
}
```

**Configuration Priority**:
1. CLI flags (highest)
2. Environment variables (prefix: `DMARC_`)
3. YAML file (default: `config.yaml`)

**Dependencies**:
- `github.com/spf13/viper`
- `github.com/spf13/pflag`
- `gopkg.in/yaml.v3`

**Unit Tests** (`internal/config/config_test.go`):
- Test YAML file loading
- Test environment variable override
- Test CLI flag override (highest priority)
- Test default values when no config provided
- Test missing required fields (should return error)
- Test invalid YAML format
- Table-driven tests for all priority combinations

### What You Should See and Test

**After completing this task:**

1. Run `go mod init dmarc-viewer` and verify `go.mod` exists
2. Run `go mod tidy` to download dependencies
3. Run unit tests: `go test ./internal/config/...` - all should pass
4. Create a test config.yaml file:
   ```yaml
   imap:
     host: imap.example.com
     port: 993
     username: test@example.com
     password: testpass
     folder: INBOX
   database:
     path: ./test.db
   web:
     port: 8080
   ```
5. Run the application: `go run cmd/dmarc-viewer/main.go --config config.yaml`
6. Verify it prints the loaded configuration
7. Test environment variable override:
   ```bash
   DMARC_IMAP_HOST=override.example.com go run cmd/dmarc-viewer/main.go --config config.yaml
   ```
   Verify the IMAP host shows "override.example.com"
8. Test CLI flag override (highest priority):
   ```bash
   DMARC_IMAP_HOST=env.example.com go run cmd/dmarc-viewer/main.go \
     --config config.yaml --imap-host=cli.example.com
   ```
   Verify the IMAP host shows "cli.example.com"

**Success Criteria**:
- All unit tests pass with >90% coverage
- Configuration loads from YAML file
- Environment variables override YAML values
- CLI flags override both environment and YAML
- Application prints loaded config and exits cleanly

---

## TASK 2: Database Schema and Basic Operations

### Objectives
- Set up SQLite database with pure Go driver
- Create schema with migrations
- Implement basic CRUD operations
- Create models for reports and records

### Files to Create
1. `internal/database/db.go` - Database connection and initialization
2. `internal/database/migrations.go` - Schema creation and migrations
3. `internal/database/models.go` - Go structs for database models
4. `internal/database/db_test.go` - Unit tests using in-memory SQLite

### Implementation Requirements

**Database Schema** (`internal/database/migrations.go`):
```sql
CREATE TABLE IF NOT EXISTS reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_uid TEXT NOT NULL,
    report_type TEXT NOT NULL, -- 'rua' or 'ruf'
    org_name TEXT,
    report_id TEXT,
    date_begin INTEGER NOT NULL,
    date_end INTEGER NOT NULL,
    email TEXT,
    domain TEXT NOT NULL,
    raw_xml TEXT,
    created_at INTEGER NOT NULL,
    UNIQUE(message_uid, report_id)
);

CREATE INDEX IF NOT EXISTS idx_reports_domain ON reports(domain);
CREATE INDEX IF NOT EXISTS idx_reports_date_begin ON reports(date_begin);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at);

CREATE TABLE IF NOT EXISTS report_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_id INTEGER NOT NULL,
    source_ip TEXT NOT NULL,
    count INTEGER NOT NULL,
    disposition TEXT,
    dkim_result TEXT,
    spf_result TEXT,
    dkim_domain TEXT,
    spf_domain TEXT,
    FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_records_report_id ON report_records(report_id);
CREATE INDEX IF NOT EXISTS idx_records_source_ip ON report_records(source_ip);

CREATE TABLE IF NOT EXISTS download_state (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_uid TEXT NOT NULL,
    folder TEXT NOT NULL,
    downloaded_at INTEGER NOT NULL,
    UNIQUE(message_uid, folder)
);

CREATE INDEX IF NOT EXISTS idx_download_state_uid ON download_state(message_uid);

CREATE TABLE IF NOT EXISTS statistics_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_name TEXT NOT NULL,
    time_period TEXT NOT NULL, -- 'day', 'week', 'month'
    value INTEGER NOT NULL,
    calculated_at INTEGER NOT NULL,
    UNIQUE(metric_name, time_period)
);
```

**Models** (`internal/database/models.go`):
```go
type Report struct {
    ID          int64
    MessageUID  string
    ReportType  string // "rua" or "ruf"
    OrgName     string
    ReportID    string
    DateBegin   time.Time
    DateEnd     time.Time
    Email       string
    Domain      string
    RawXML      string
    CreatedAt   time.Time
}

type ReportRecord struct {
    ID          int64
    ReportID    int64
    SourceIP    string
    Count       int
    Disposition string
    DKIMResult  string
    SPFResult   string
    DKIMDomain  string
    SPFDomain   string
}

type DownloadState struct {
    ID           int64
    MessageUID   string
    Folder       string
    DownloadedAt time.Time
}

type StatisticsCache struct {
    ID           int64
    MetricName   string
    TimePeriod   string
    Value        int64
    CalculatedAt time.Time
}
```

**Database Operations** (`internal/database/db.go`):
```go
type DB struct {
    conn *sql.DB
}

func New(path string) (*DB, error)
func (db *DB) Close() error
func (db *DB) InsertReport(report *Report) (int64, error)
func (db *DB) GetReport(id int64) (*Report, error)
func (db *DB) ListReports(limit, offset int) ([]*Report, error)
func (db *DB) InsertReportRecords(records []*ReportRecord) error
func (db *DB) GetReportRecords(reportID int64) ([]*ReportRecord, error)
func (db *DB) IsDownloaded(messageUID, folder string) (bool, error)
func (db *DB) MarkDownloaded(messageUID, folder string) error
```

**Dependencies**:
- `modernc.org/sqlite` (pure Go SQLite driver)
- `database/sql` (standard library)

**Unit Tests** (`internal/database/db_test.go`):
- Test database initialization
- Test schema creation
- Test INSERT operations for all tables
- Test SELECT operations (single and list)
- Test UNIQUE constraints (duplicate reports)
- Test foreign key constraints
- Test download state tracking
- Test concurrent access (multiple goroutines)
- Use `:memory:` for in-memory testing

### What You Should See and Test

**After completing this task:**

1. Run unit tests: `go test ./internal/database/... -v`
2. All tests should pass with >85% coverage
3. Create a simple test program:
   ```go
   package main

   import (
       "dmarc-viewer/internal/database"
       "time"
   )

   func main() {
       db, _ := database.New("./test.db")
       defer db.Close()

       report := &database.Report{
           MessageUID: "12345",
           ReportType: "rua",
           OrgName:    "example.com",
           Domain:     "mydomain.com",
           DateBegin:  time.Now().Add(-24 * time.Hour),
           DateEnd:    time.Now(),
           CreatedAt:  time.Now(),
       }

       id, _ := db.InsertReport(report)
       println("Inserted report with ID:", id)

       fetched, _ := db.GetReport(id)
       println("Fetched report:", fetched.Domain)
   }
   ```
4. Run the test program and verify database file is created
5. Use `sqlite3` CLI to inspect the database:
   ```bash
   sqlite3 test.db ".schema"
   sqlite3 test.db "SELECT * FROM reports;"
   ```
6. Verify all tables and indexes exist
7. Test duplicate insertion (should fail with UNIQUE constraint)

**Success Criteria**:
- All unit tests pass
- Database schema is created correctly
- CRUD operations work for all tables
- UNIQUE constraints prevent duplicates
- Foreign keys work correctly (cascade delete)
- In-memory database tests complete quickly (<1 second)

---

## TASK 3: DMARC Report Parser (RUA/XML)

### Objectives
- Parse DMARC aggregate (RUA) reports in XML format
- Extract all relevant data fields
- Handle compressed attachments (gzip, zip)
- Create robust error handling for malformed XML

### Files to Create
1. `internal/parser/rua.go` - RUA XML parser
2. `internal/parser/rua_test.go` - Unit tests with sample reports
3. `internal/parser/compression.go` - Decompression utilities
4. `internal/parser/compression_test.go` - Decompression tests
5. `testdata/sample_rua.xml` - Sample RUA report
6. `testdata/sample_rua.xml.gz` - Gzipped sample

### Implementation Requirements

**RUA XML Structure** (based on RFC 7489):
```xml
<?xml version="1.0"?>
<feedback>
  <report_metadata>
    <org_name>example.com</org_name>
    <email>noreply@example.com</email>
    <report_id>12345</report_id>
    <date_range>
      <begin>1234567890</begin>
      <end>1234654290</end>
    </date_range>
  </report_metadata>
  <policy_published>
    <domain>mydomain.com</domain>
    <p>reject</p>
    <sp>reject</sp>
    <pct>100</pct>
  </policy_published>
  <record>
    <row>
      <source_ip>192.0.2.1</source_ip>
      <count>5</count>
      <policy_evaluated>
        <disposition>none</disposition>
        <dkim>pass</dkim>
        <spf>fail</spf>
      </policy_evaluated>
    </row>
    <identifiers>
      <header_from>mydomain.com</header_from>
    </identifiers>
    <auth_results>
      <dkim>
        <domain>mydomain.com</domain>
        <result>pass</result>
      </dkim>
      <spf>
        <domain>mydomain.com</domain>
        <result>fail</result>
      </spf>
    </auth_results>
  </record>
</feedback>
```

**Parser Functions** (`internal/parser/rua.go`):
```go
type RUAReport struct {
    Metadata      ReportMetadata
    PolicyPublished PolicyPublished
    Records       []ReportRecord
    RawXML        string
}

type ReportMetadata struct {
    OrgName   string
    Email     string
    ReportID  string
    DateBegin time.Time
    DateEnd   time.Time
}

type PolicyPublished struct {
    Domain string
    Policy string
    SubdomainPolicy string
    Percentage int
}

type ReportRecord struct {
    SourceIP    string
    Count       int
    Disposition string
    DKIMResult  string
    SPFResult   string
    DKIMDomain  string
    SPFDomain   string
}

func ParseRUA(xmlData []byte) (*RUAReport, error)
func ParseRUAFromFile(filepath string) (*RUAReport, error)
```

**Decompression** (`internal/parser/compression.go`):
```go
func DecompressGzip(data []byte) ([]byte, error)
func DecompressZip(data []byte) ([]byte, error)
func DetectAndDecompress(data []byte) ([]byte, error)
```

**Dependencies**:
- `encoding/xml` (standard library)
- `compress/gzip` (standard library)
- `archive/zip` (standard library)

**Unit Tests** (`internal/parser/rua_test.go`):
- Test parsing valid RUA XML
- Test parsing with missing optional fields
- Test parsing malformed XML (should return error)
- Test parsing empty records
- Test parsing multiple records
- Test gzip decompression
- Test zip decompression
- Test invalid compression format
- Use table-driven tests with multiple XML samples

### What You Should See and Test

**After completing this task:**

1. Create sample RUA report in `testdata/sample_rua.xml` (use the XML structure above)
2. Run unit tests: `go test ./internal/parser/... -v`
3. All tests should pass with >90% coverage
4. Create a test program:
   ```go
   package main

   import (
       "dmarc-viewer/internal/parser"
       "fmt"
       "os"
   )

   func main() {
       data, _ := os.ReadFile("testdata/sample_rua.xml")
       report, err := parser.ParseRUA(data)
       if err != nil {
           panic(err)
       }

       fmt.Printf("Organization: %s\n", report.Metadata.OrgName)
       fmt.Printf("Domain: %s\n", report.PolicyPublished.Domain)
       fmt.Printf("Records: %d\n", len(report.Records))

       for i, record := range report.Records {
           fmt.Printf("  Record %d: IP=%s, Count=%d, DKIM=%s, SPF=%s\n",
               i+1, record.SourceIP, record.Count,
               record.DKIMResult, record.SPFResult)
       }
   }
   ```
5. Run the test program and verify correct parsing
6. Test with gzipped file:
   ```bash
   gzip -c testdata/sample_rua.xml > testdata/sample_rua.xml.gz
   ```
7. Modify test program to decompress first, then parse
8. Test with intentionally malformed XML and verify error handling

**Success Criteria**:
- Parser correctly extracts all fields from valid RUA reports
- Parser handles missing optional fields gracefully
- Decompression works for gzip and zip formats
- Error messages are clear for malformed XML
- All unit tests pass
- Real-world RUA reports parse successfully

---

## TASK 4: IMAP Client and Email Fetching

### Objectives
- Connect to IMAP server with TLS support
- Search for messages in specified folder
- Download email attachments
- Track downloaded messages to prevent re-downloading
- Handle connection errors and retries

### Files to Create
1. `internal/imap/client.go` - IMAP client implementation
2. `internal/imap/state.go` - Download state tracking
3. `internal/imap/client_test.go` - Unit tests with mock IMAP server
4. `internal/imap/state_test.go` - Download state tests

### Implementation Requirements

**IMAP Client** (`internal/imap/client.go`):
```go
type Client struct {
    config *config.IMAPConfig
    conn   *client.Client
    db     *database.DB
}

func NewClient(cfg *config.IMAPConfig, db *database.DB) (*Client, error)
func (c *Client) Connect() error
func (c *Client) Disconnect() error
func (c *Client) FetchNewMessages() ([]*Message, error)
func (c *Client) GetAttachments(msg *Message) ([]Attachment, error)

type Message struct {
    UID     uint32
    From    string
    Subject string
    Date    time.Time
}

type Attachment struct {
    Filename string
    Data     []byte
}
```

**State Tracking** (`internal/imap/state.go`):
```go
type StateTracker struct {
    db *database.DB
}

func NewStateTracker(db *database.DB) *StateTracker
func (s *StateTracker) IsDownloaded(uid string, folder string) (bool, error)
func (s *StateTracker) MarkDownloaded(uid string, folder string) error
func (s *StateTracker) GetLastUID(folder string) (uint32, error)
```

**Dependencies**:
- `github.com/emersion/go-imap` (pure Go IMAP client)
- `github.com/emersion/go-message` (email parsing)

**Implementation Details**:
- Use IMAP SEARCH to find messages efficiently
- Only fetch messages with UIDs not in download_state table
- Extract attachments that look like DMARC reports (XML, gz, zip)
- Support both implicit TLS (port 993) and STARTTLS
- Implement exponential backoff for retries
- Use SELECT to open folder in read-only mode

**Unit Tests** (`internal/imap/client_test.go`):
- Test connection with valid credentials
- Test connection failure with invalid credentials
- Test folder selection
- Test message search
- Test attachment extraction
- Test state tracking (don't re-download)
- Mock IMAP server responses for testing
- Test TLS connection

### What You Should See and Test

**After completing this task:**

1. Run unit tests: `go test ./internal/imap/... -v`
2. Create a test program (requires actual IMAP server):
   ```go
   package main

   import (
       "dmarc-viewer/internal/config"
       "dmarc-viewer/internal/database"
       "dmarc-viewer/internal/imap"
       "fmt"
   )

   func main() {
       cfg := &config.IMAPConfig{
           Host:     "imap.gmail.com",
           Port:     993,
           Username: "your-email@gmail.com",
           Password: "your-app-password",
           Folder:   "INBOX",
           UseTLS:   true,
       }

       db, _ := database.New("./test.db")
       defer db.Close()

       client, _ := imap.NewClient(cfg, db)
       defer client.Disconnect()

       err := client.Connect()
       if err != nil {
           panic(err)
       }

       messages, _ := client.FetchNewMessages()
       fmt.Printf("Found %d new messages\n", len(messages))

       for _, msg := range messages {
           attachments, _ := client.GetAttachments(msg)
           fmt.Printf("Message UID %d has %d attachments\n",
               msg.UID, len(attachments))
       }
   }
   ```
3. Test with a real IMAP server (Gmail, Outlook, etc.)
4. Verify connection succeeds
5. Verify messages are found in the specified folder
6. Run the program twice - second run should show 0 new messages (already downloaded)
7. Add a new email to the folder and verify it's detected
8. Test with invalid credentials and verify error handling

**Success Criteria**:
- Successfully connects to IMAP server with TLS
- Can list messages in specified folder
- Extracts attachments from messages
- Tracks downloaded UIDs in database
- Re-running doesn't re-download same messages
- Handles connection errors gracefully
- All unit tests pass

---

## TASK 5: Integration - IMAP to Database Pipeline

### Objectives
- Combine IMAP client, parser, and database modules
- Create sync service that fetches, parses, and stores reports
- Implement scheduled syncing
- Add logging throughout the pipeline

### Files to Create
1. `internal/sync/service.go` - Sync orchestration service
2. `internal/sync/service_test.go` - Integration tests
3. Update `cmd/dmarc-viewer/main.go` - Add sync command

### Implementation Requirements

**Sync Service** (`internal/sync/service.go`):
```go
type Service struct {
    config *config.Config
    db     *database.DB
    imap   *imap.Client
    logger *zerolog.Logger
}

func NewService(cfg *config.Config, db *database.DB, logger *zerolog.Logger) (*Service, error)
func (s *Service) SyncOnce() error
func (s *Service) StartScheduled(ctx context.Context) error
func (s *Service) processMessage(msg *imap.Message) error
```

**Pipeline Flow**:
1. Connect to IMAP server
2. Fetch new messages from folder
3. For each message:
   - Get attachments
   - Decompress if needed
   - Parse XML (RUA)
   - Insert into database (report + records)
   - Mark as downloaded
4. Disconnect from IMAP
5. Log statistics (messages processed, errors, duration)

**Logging** (use `github.com/rs/zerolog`):
- INFO: Sync started/completed, message counts
- DEBUG: Each message processed, attachment details
- WARN: Parse failures, skipped messages
- ERROR: Connection failures, database errors

**Dependencies**:
- `github.com/rs/zerolog` (structured logging)
- `context` (standard library for cancellation)

**Unit/Integration Tests** (`internal/sync/service_test.go`):
- Test full pipeline with sample email
- Test handling of multiple reports
- Test error handling (bad XML, database failure)
- Test duplicate prevention
- Test scheduled sync with context cancellation
- Use in-memory database and mock IMAP

### What You Should See and Test

**After completing this task:**

1. Update main.go to support sync command:
   ```bash
   go run cmd/dmarc-viewer/main.go sync --config config.yaml
   ```
2. Run unit tests: `go test ./internal/sync/... -v`
3. Prepare test environment:
   - Create config.yaml with real IMAP credentials
   - Send yourself a DMARC report email (use sample XML as attachment)
4. Run sync command:
   ```bash
   go run cmd/dmarc-viewer/main.go sync --config config.yaml
   ```
5. Verify console output shows:
   - "Sync started"
   - "Connected to IMAP server"
   - "Processing message UID: ..."
   - "Parsed RUA report from ..."
   - "Inserted report ID: ..."
   - "Sync completed: 1 reports processed"
6. Check database:
   ```bash
   sqlite3 dmarc-reports.db "SELECT * FROM reports;"
   sqlite3 dmarc-reports.db "SELECT * FROM report_records;"
   ```
7. Verify report and records are inserted
8. Run sync again - should show "0 new messages"
9. Test scheduled sync:
   ```bash
   go run cmd/dmarc-viewer/main.go sync --config config.yaml --scheduled
   ```
10. Verify it runs continuously (Ctrl+C to stop)

**Success Criteria**:
- Sync command successfully processes DMARC reports
- Reports and records are inserted into database
- Duplicate messages are not re-processed
- Logs show clear progress and errors
- Scheduled sync runs at configured interval
- Graceful shutdown on SIGINT/SIGTERM
- All integration tests pass

---

## TASK 6: Web Server Foundation and Routing

### Objectives
- Create HTTP server with chi router
- Set up HTML templating
- Create base layout and navigation
- Implement health check endpoint

### Files to Create
1. `internal/web/server.go` - HTTP server setup
2. `internal/web/handlers.go` - HTTP handler functions
3. `internal/web/templates/layout.html` - Base HTML layout
4. `internal/web/templates/dashboard.html` - Dashboard page
5. `internal/web/server_test.go` - HTTP handler tests

### Implementation Requirements

**Server Setup** (`internal/web/server.go`):
```go
type Server struct {
    config *config.Config
    db     *database.DB
    router *chi.Mux
    logger *zerolog.Logger
}

func NewServer(cfg *config.Config, db *database.DB, logger *zerolog.Logger) *Server
func (s *Server) setupRoutes()
func (s *Server) Start() error
func (s *Server) Shutdown(ctx context.Context) error
```

**Routes**:
- `GET /` - Dashboard
- `GET /health` - Health check (JSON)
- `GET /reports` - Report list page
- `GET /reports/{id}` - Report detail page
- `GET /api/stats` - Statistics API (JSON)
- `POST /api/sync` - Trigger manual sync

**Base Layout** (`internal/web/templates/layout.html`):
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - DMARC Viewer</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        /* Basic CSS - clean, professional */
        body { font-family: system-ui; max-width: 1200px; margin: 0 auto; padding: 20px; }
        nav { background: #333; color: white; padding: 15px; margin-bottom: 20px; }
        nav a { color: white; margin-right: 15px; text-decoration: none; }
    </style>
</head>
<body>
    <nav>
        <a href="/">Dashboard</a>
        <a href="/reports">Reports</a>
    </nav>
    <main>
        {{template "content" .}}
    </main>
</body>
</html>
```

**Dashboard Template** (`internal/web/templates/dashboard.html`):
```html
{{define "content"}}
<h1>DMARC Report Dashboard</h1>
<div id="stats" hx-get="/api/stats" hx-trigger="load, every 30s">
    Loading statistics...
</div>
{{end}}
```

**Dependencies**:
- `github.com/go-chi/chi` (HTTP router)
- `html/template` (standard library)
- `net/http` (standard library)

**Unit Tests** (`internal/web/server_test.go`):
- Test server initialization
- Test health check endpoint
- Test routing (all routes exist)
- Test template rendering
- Use `httptest` package

### What You Should See and Test

**After completing this task:**

1. Update main.go to support web command:
   ```bash
   go run cmd/dmarc-viewer/main.go web --config config.yaml
   ```
2. Run unit tests: `go test ./internal/web/... -v`
3. Start the web server:
   ```bash
   go run cmd/dmarc-viewer/main.go web --config config.yaml
   ```
4. Open browser to `http://localhost:8080`
5. Verify you see:
   - Navigation bar with "Dashboard" and "Reports" links
   - "DMARC Report Dashboard" heading
   - "Loading statistics..." message
6. Check health endpoint:
   ```bash
   curl http://localhost:8080/health
   ```
   Should return: `{"status":"ok"}`
7. Test graceful shutdown (Ctrl+C) - should log "Server stopped"
8. Test different port:
   ```bash
   go run cmd/dmarc-viewer/main.go web --web-port 9090
   ```
   Verify it runs on port 9090

**Success Criteria**:
- Web server starts and listens on configured port
- Health check endpoint returns JSON
- Base layout renders correctly
- Navigation links are visible
- Templates use HTMX script
- Graceful shutdown works
- All unit tests pass

---

## TASK 7: Dashboard and Statistics Display

### Objectives
- Implement statistics calculation module
- Create dashboard with key metrics
- Display statistics using HTMX for auto-refresh
- Show failure counts for last day/week/month

### Files to Create
1. `internal/stats/calculator.go` - Statistics calculation
2. `internal/stats/calculator_test.go` - Statistics tests
3. Update `internal/web/handlers.go` - Add stats handlers
4. Update `internal/web/templates/dashboard.html` - Stats display

### Implementation Requirements

**Statistics Calculator** (`internal/stats/calculator.go`):
```go
type Calculator struct {
    db *database.DB
}

type Statistics struct {
    TotalReports    int64
    TotalRecords    int64
    FailuresDay     int64
    FailuresWeek    int64
    FailuresMonth   int64
    PassRateDay     float64
    PassRateWeek    float64
    PassRateMonth   float64
    TopFailingSources []SourceStat
}

type SourceStat struct {
    SourceIP string
    Failures int64
}

func NewCalculator(db *database.DB) *Calculator
func (c *Calculator) Calculate() (*Statistics, error)
func (c *Calculator) GetFailures(since time.Time) (int64, error)
func (c *Calculator) GetPassRate(since time.Time) (float64, error)
func (c *Calculator) GetTopFailingSources(limit int) ([]SourceStat, error)
```

**Calculation Logic**:
- Failures = records where disposition != "none" OR dkim_result != "pass" OR spf_result != "pass"
- Pass rate = (total_count - failure_count) / total_count * 100
- Last day = now - 24 hours
- Last week = now - 7 days
- Last month = now - 30 days

**Dashboard Template** (`internal/web/templates/dashboard.html`):
```html
{{define "content"}}
<h1>DMARC Report Dashboard</h1>

<div id="stats" hx-get="/api/stats" hx-trigger="load, every 30s" hx-swap="innerHTML">
    <p>Loading statistics...</p>
</div>

<div id="stats-template" style="display: none;">
    <div class="stats-grid">
        <div class="stat-card">
            <h3>Total Reports</h3>
            <p class="stat-value">{{.TotalReports}}</p>
        </div>
        <div class="stat-card">
            <h3>Failures (24h)</h3>
            <p class="stat-value">{{.FailuresDay}}</p>
        </div>
        <div class="stat-card">
            <h3>Failures (7d)</h3>
            <p class="stat-value">{{.FailuresWeek}}</p>
        </div>
        <div class="stat-card">
            <h3>Failures (30d)</h3>
            <p class="stat-value">{{.FailuresMonth}}</p>
        </div>
        <div class="stat-card">
            <h3>Pass Rate (24h)</h3>
            <p class="stat-value">{{printf "%.1f%%" .PassRateDay}}</p>
        </div>
    </div>

    <h2>Top Failing Sources</h2>
    <table>
        <tr><th>Source IP</th><th>Failures</th></tr>
        {{range .TopFailingSources}}
        <tr><td>{{.SourceIP}}</td><td>{{.Failures}}</td></tr>
        {{end}}
    </table>
</div>
{{end}}
```

**API Handler** (`internal/web/handlers.go`):
```go
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
    calc := stats.NewCalculator(s.db)
    stats, err := calc.Calculate()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // If HTMX request, return HTML fragment
    if r.Header.Get("HX-Request") == "true" {
        s.renderTemplate(w, "stats-fragment", stats)
        return
    }

    // Otherwise return JSON
    json.NewEncoder(w).Encode(stats)
}
```

**Unit Tests** (`internal/stats/calculator_test.go`):
- Test calculation with empty database
- Test failure counting
- Test pass rate calculation
- Test time filtering (day/week/month)
- Test top failing sources
- Use in-memory database with sample data

### What You Should See and Test

**After completing this task:**

1. Insert test data into database:
   ```sql
   -- Create some sample reports
   INSERT INTO reports (message_uid, report_type, domain, date_begin, date_end, created_at)
   VALUES ('test1', 'rua', 'example.com', strftime('%s', 'now', '-2 days'), strftime('%s', 'now', '-1 day'), strftime('%s', 'now'));

   -- Create some sample records (failures)
   INSERT INTO report_records (report_id, source_ip, count, disposition, dkim_result, spf_result)
   VALUES (1, '192.0.2.1', 10, 'reject', 'fail', 'fail');

   INSERT INTO report_records (report_id, source_ip, count, disposition, dkim_result, spf_result)
   VALUES (1, '192.0.2.2', 5, 'none', 'pass', 'pass');
   ```
2. Run unit tests: `go test ./internal/stats/... -v`
3. Start web server: `go run cmd/dmarc-viewer/main.go web`
4. Open `http://localhost:8080` in browser
5. Verify dashboard shows:
   - Total Reports: 1
   - Failures (24h): 10 (from the failed record)
   - Pass Rate: Shows calculated percentage
   - Top Failing Sources table with IPs
6. Open browser dev tools network tab
7. Wait 30 seconds and verify automatic refresh (HTMX request)
8. Test JSON API:
   ```bash
   curl http://localhost:8080/api/stats
   ```
   Should return JSON with statistics
9. Add more test data and refresh - verify numbers update

**Success Criteria**:
- Statistics are calculated correctly
- Dashboard displays all metrics
- HTMX auto-refreshes every 30 seconds
- JSON API returns same data
- Works with empty database (shows zeros)
- All unit tests pass with >85% coverage

---

## TASK 8: Report List View with Pagination

### Objectives
- Display list of all reports sorted by date
- Implement pagination (20 reports per page)
- Use HTMX for infinite scroll or "Load More" button
- Show key info for each report (domain, org, date, failure count)

### Files to Create
1. Update `internal/database/db.go` - Add pagination queries
2. Update `internal/web/handlers.go` - Add report list handler
3. `internal/web/templates/reports.html` - Report list template
4. `internal/web/templates/report_row.html` - Single report row template

### Implementation Requirements

**Database Queries** (`internal/database/db.go`):
```go
func (db *DB) ListReports(limit, offset int) ([]*Report, error)
func (db *DB) CountReports() (int64, error)
func (db *DB) GetReportWithStats(id int64) (*ReportWithStats, error)

type ReportWithStats struct {
    Report
    TotalCount    int
    FailureCount  int
    PassCount     int
}
```

**Report List Handler** (`internal/web/handlers.go`):
```go
func (s *Server) handleReportList(w http.ResponseWriter, r *http.Request) {
    page := getQueryInt(r, "page", 0)
    limit := 20
    offset := page * limit

    reports, err := s.db.ListReports(limit, offset)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // If HTMX request, return rows only
    if r.Header.Get("HX-Request") == "true" {
        s.renderTemplate(w, "report-rows", reports)
        return
    }

    // Full page render
    data := map[string]interface{}{
        "Reports": reports,
        "NextPage": page + 1,
    }
    s.renderTemplate(w, "reports", data)
}
```

**Report List Template** (`internal/web/templates/reports.html`):
```html
{{define "content"}}
<h1>DMARC Reports</h1>

<table class="reports-table">
    <thead>
        <tr>
            <th>Domain</th>
            <th>Organization</th>
            <th>Date Range</th>
            <th>Total Messages</th>
            <th>Failures</th>
            <th></th>
        </tr>
    </thead>
    <tbody id="report-list">
        {{range .Reports}}
        {{template "report-row" .}}
        {{end}}
    </tbody>
</table>

<div id="load-more">
    <button
        hx-get="/reports?page={{.NextPage}}"
        hx-target="#report-list"
        hx-swap="beforeend"
        hx-select="#report-list > *">
        Load More
    </button>
</div>
{{end}}

{{define "report-row"}}
<tr>
    <td>{{.Domain}}</td>
    <td>{{.OrgName}}</td>
    <td>{{.DateBegin.Format "2006-01-02"}} - {{.DateEnd.Format "2006-01-02"}}</td>
    <td>{{.TotalCount}}</td>
    <td class="{{if gt .FailureCount 0}}text-danger{{end}}">{{.FailureCount}}</td>
    <td>
        <a href="/reports/{{.ID}}"
           hx-get="/reports/{{.ID}}"
           hx-target="#detail-pane"
           hx-swap="innerHTML">
            View
        </a>
    </td>
</tr>
{{end}}
```

**Unit Tests**:
- Test pagination with various page numbers
- Test empty result set
- Test HTMX vs full page render
- Test report count query

### What You Should See and Test

**After completing this task:**

1. Populate database with multiple reports:
   ```go
   for i := 0; i < 50; i++ {
       db.InsertReport(&database.Report{
           MessageUID: fmt.Sprintf("test%d", i),
           Domain:     "example.com",
           DateBegin:  time.Now().Add(-time.Duration(i) * 24 * time.Hour),
           DateEnd:    time.Now().Add(-time.Duration(i-1) * 24 * time.Hour),
           // ...
       })
   }
   ```
2. Run unit tests: `go test ./internal/web/... -v`
3. Start web server: `go run cmd/dmarc-viewer/main.go web`
4. Click "Reports" in navigation
5. Verify you see:
   - Table with report columns
   - First 20 reports
   - "Load More" button at bottom
6. Click "Load More"
7. Verify:
   - Next 20 reports are appended (not replaced)
   - Button remains for more pages
8. Open browser dev tools
9. Click "Load More" again
10. Verify HTMX request in network tab:
    - Request header: `HX-Request: true`
    - Response contains only table rows (no full page)
11. Test with fewer than 20 reports - "Load More" should disappear

**Success Criteria**:
- Report list displays correctly
- Pagination works (20 per page)
- "Load More" appends reports (infinite scroll style)
- HTMX requests return partial HTML
- Full page requests return complete page
- Sorting is by date (newest first)
- All unit tests pass

---

## TASK 9: Report Detail View

### Objectives
- Display full details of a single report
- Show all report records with source IPs and results
- Highlight failures in red/yellow
- Load detail view in a side panel using HTMX

### Files to Create
1. Update `internal/web/handlers.go` - Add report detail handler
2. `internal/web/templates/report_detail.html` - Detail view template
3. Update `internal/web/templates/reports.html` - Add detail pane

### Implementation Requirements

**Database Queries** (`internal/database/db.go`):
```go
func (db *DB) GetReportWithRecords(id int64) (*ReportDetail, error)

type ReportDetail struct {
    Report  *Report
    Records []*ReportRecord
    Stats   struct {
        TotalMessages int
        PassCount     int
        FailCount     int
        PassRate      float64
    }
}
```

**Detail Handler** (`internal/web/handlers.go`):
```go
func (s *Server) handleReportDetail(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid report ID", http.StatusBadRequest)
        return
    }

    detail, err := s.db.GetReportWithRecords(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    s.renderTemplate(w, "report-detail", detail)
}
```

**Detail Template** (`internal/web/templates/report_detail.html`):
```html
{{define "report-detail"}}
<div class="report-detail">
    <h2>Report Details</h2>

    <div class="report-metadata">
        <p><strong>Domain:</strong> {{.Report.Domain}}</p>
        <p><strong>Organization:</strong> {{.Report.OrgName}}</p>
        <p><strong>Report ID:</strong> {{.Report.ReportID}}</p>
        <p><strong>Date Range:</strong>
            {{.Report.DateBegin.Format "2006-01-02 15:04"}} -
            {{.Report.DateEnd.Format "2006-01-02 15:04"}}
        </p>
    </div>

    <div class="report-stats">
        <h3>Summary</h3>
        <p>Total Messages: {{.Stats.TotalMessages}}</p>
        <p>Passed: {{.Stats.PassCount}} ({{printf "%.1f%%" .Stats.PassRate}})</p>
        <p class="text-danger">Failed: {{.Stats.FailCount}}</p>
    </div>

    <h3>Records</h3>
    <table class="records-table">
        <thead>
            <tr>
                <th>Source IP</th>
                <th>Count</th>
                <th>Disposition</th>
                <th>DKIM</th>
                <th>SPF</th>
                <th>DKIM Domain</th>
                <th>SPF Domain</th>
            </tr>
        </thead>
        <tbody>
            {{range .Records}}
            <tr class="{{if or (ne .Disposition "none") (ne .DKIMResult "pass") (ne .SPFResult "pass")}}failure-row{{end}}">
                <td>{{.SourceIP}}</td>
                <td>{{.Count}}</td>
                <td>{{.Disposition}}</td>
                <td class="{{if ne .DKIMResult "pass"}}text-danger{{end}}">{{.DKIMResult}}</td>
                <td class="{{if ne .SPFResult "pass"}}text-danger{{end}}">{{.SPFResult}}</td>
                <td>{{.DKIMDomain}}</td>
                <td>{{.SPFDomain}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <details>
        <summary>Raw XML</summary>
        <pre>{{.Report.RawXML}}</pre>
    </details>

    <button onclick="document.getElementById('detail-pane').innerHTML=''">Close</button>
</div>
{{end}}
```

**Updated Reports Page** (`internal/web/templates/reports.html`):
```html
{{define "content"}}
<div class="reports-layout">
    <div class="reports-list">
        <h1>DMARC Reports</h1>
        <!-- existing table code -->
    </div>

    <div id="detail-pane" class="detail-pane">
        <!-- Detail view loads here via HTMX -->
    </div>
</div>

<style>
    .reports-layout {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
    }

    .detail-pane {
        position: sticky;
        top: 20px;
        max-height: 90vh;
        overflow-y: auto;
        border-left: 2px solid #ccc;
        padding-left: 20px;
    }

    .failure-row {
        background-color: #ffe6e6;
    }

    .text-danger {
        color: #d32f2f;
        font-weight: bold;
    }
</style>
{{end}}
```

**Unit Tests**:
- Test detail view with valid ID
- Test with invalid ID (404)
- Test with report that has no records
- Test stats calculation
- Test highlighting of failures

### What You Should See and Test

**After completing this task:**

1. Ensure database has reports with records
2. Run unit tests: `go test ./internal/web/... -v`
3. Start web server: `go run cmd/dmarc-viewer/main.go web`
4. Navigate to Reports page
5. Click "View" on any report
6. Verify:
   - Detail pane appears on the right side
   - Report metadata is displayed (domain, org, dates)
   - Summary stats are shown
   - Records table lists all source IPs
   - Failed records are highlighted in red
   - DKIM/SPF failures are shown in red
7. Click "Close" button - detail pane clears
8. Click another report - detail pane updates (no page reload)
9. Test with browser dev tools:
   - Network tab shows only HTML fragment loaded
   - No full page refresh
10. Expand "Raw XML" section - verify XML is displayed
11. Test with invalid report ID:
    ```bash
    curl http://localhost:8080/reports/99999
    ```
    Should return 404

**Success Criteria**:
- Detail view loads in side panel
- No full page refresh (HTMX magic)
- All report data is displayed correctly
- Failures are visually highlighted
- Records table shows all source IPs
- Raw XML is available (collapsed by default)
- Close button works
- Multiple detail views can be loaded sequentially
- All unit tests pass

---

## TASK 10: Manual Sync Trigger and Progress Indicator

### Objectives
- Add button to trigger manual sync from web interface
- Show sync progress with HTMX
- Display success/error messages
- Update stats and report list after sync

### Files to Create
1. Update `internal/web/handlers.go` - Add sync trigger handler
2. Update `internal/web/templates/dashboard.html` - Add sync button
3. Add `internal/web/templates/sync_status.html` - Sync progress template

### Implementation Requirements

**Sync Handler** (`internal/web/handlers.go`):
```go
func (s *Server) handleSyncTrigger(w http.ResponseWriter, r *http.Request) {
    // Run sync in background
    go func() {
        err := s.syncService.SyncOnce()
        if err != nil {
            s.logger.Error().Err(err).Msg("Manual sync failed")
        }
    }()

    // Return immediate response
    w.Header().Set("HX-Trigger", "syncStarted")
    s.renderTemplate(w, "sync-status", map[string]string{
        "Status": "Sync started...",
        "Class":  "info",
    })
}

func (s *Server) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
    // Check if sync is running
    status := s.syncService.GetStatus()
    s.renderTemplate(w, "sync-status", status)
}
```

**Sync Button** (`internal/web/templates/dashboard.html`):
```html
{{define "content"}}
<h1>DMARC Report Dashboard</h1>

<div class="actions">
    <button
        hx-post="/api/sync"
        hx-target="#sync-status"
        hx-swap="innerHTML"
        hx-indicator="#sync-spinner">
        Sync Now
    </button>
    <span id="sync-spinner" class="htmx-indicator">Syncing...</span>
</div>

<div id="sync-status"></div>

<div id="stats" hx-get="/api/stats" hx-trigger="load, every 30s, syncComplete from:body">
    <p>Loading statistics...</p>
</div>
{{end}}

{{define "sync-status"}}
<div class="alert alert-{{.Class}}">
    {{.Status}}
</div>
{{end}}
```

**HTMX Events**:
- `syncStarted` - Triggers when sync begins
- `syncComplete` - Triggers when sync finishes (refresh stats)
- `syncError` - Triggers on sync failure

**Unit Tests**:
- Test sync trigger endpoint
- Test sync status endpoint
- Test concurrent sync requests
- Test error handling

### What You Should See and Test

**After completing this task:**

1. Ensure you have test emails with DMARC reports in your IMAP folder
2. Run unit tests: `go test ./internal/web/... -v`
3. Start web server: `go run cmd/dmarc-viewer/main.go web`
4. Open dashboard `http://localhost:8080`
5. Click "Sync Now" button
6. Verify:
   - Button shows "Syncing..." indicator
   - Status message appears: "Sync started..."
   - After a few seconds, status updates (check console logs)
   - Statistics automatically refresh when sync completes
7. Check browser dev tools:
   - Network tab shows POST to `/api/sync`
   - HTMX triggers can be seen in console
8. Navigate to Reports page
9. Verify new reports appear after sync
10. Test clicking "Sync Now" multiple times rapidly
11. Verify it handles concurrent requests gracefully
12. Test with no new messages:
    - Click "Sync Now"
    - Status should show "No new reports found"

**Success Criteria**:
- Sync button triggers background sync
- Progress indicator shows during sync
- Success/error messages are displayed
- Statistics refresh automatically after sync
- Report list updates with new reports
- Concurrent sync requests are handled
- All unit tests pass

---

## TASK 11: Comprehensive Testing and Documentation

### Objectives
- Achieve >80% test coverage across all modules
- Write integration tests for full workflow
- Create README with setup instructions
- Add code documentation and examples

### Files to Create
1. `integration_test.go` - End-to-end integration test
2. `README.md` - Project documentation
3. `Makefile` - Build and test automation
4. `docker-compose.yml` - Optional Docker setup
5. Update all `*_test.go` files to improve coverage

### Implementation Requirements

**Integration Test** (`integration_test.go`):
```go
// Test full workflow: IMAP → Parse → Database → Web API
func TestFullPipeline(t *testing.T) {
    // 1. Setup test IMAP server with sample message
    // 2. Setup in-memory database
    // 3. Run sync
    // 4. Verify report in database
    // 5. Query via web API
    // 6. Verify JSON response
}
```

**README.md Structure**:
```markdown
# DMARC Report Viewer

## Features
- Automatic IMAP report fetching
- RUA (aggregate) report parsing
- Web dashboard with statistics
- Real-time updates with HTMX
- Pure Go implementation

## Installation
## Configuration
## Usage
## Development
## Testing
## License
```

**Makefile**:
```makefile
.PHONY: build test coverage run clean

build:
	go build -o dmarc-viewer cmd/dmarc-viewer/main.go

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run:
	go run cmd/dmarc-viewer/main.go web

clean:
	rm -f dmarc-viewer coverage.out dmarc-reports.db
```

**Documentation Requirements**:
- Every exported function has GoDoc comment
- Complex algorithms have inline comments
- Examples in README for common use cases
- Configuration options documented
- Troubleshooting section

**Unit Tests to Add**:
- Edge cases for all parsers
- Error handling paths
- Boundary conditions (0 records, max int, etc.)
- Concurrent access tests
- Resource cleanup tests (defer, Close)

### What You Should See and Test

**After completing this task:**

1. Run coverage report:
   ```bash
   make coverage
   ```
   Open `coverage.out` in browser and verify >80% coverage for all packages

2. Run integration test:
   ```bash
   go test -v -run TestFullPipeline
   ```
   Should pass with full workflow working

3. Follow README instructions from scratch:
   - Clone repository (simulated)
   - Install dependencies
   - Configure IMAP settings
   - Run application
   - Verify everything works as documented

4. Test Makefile targets:
   ```bash
   make build   # Should create binary
   make test    # Should run all tests
   make run     # Should start web server
   make clean   # Should remove artifacts
   ```

5. Verify documentation:
   ```bash
   go doc dmarc-viewer/internal/parser
   go doc dmarc-viewer/internal/database
   ```
   All exported functions should have clear documentation

6. Run linter:
   ```bash
   go vet ./...
   golangci-lint run  # if installed
   ```
   Should report no issues

7. Test with real IMAP account end-to-end:
   - Configure with your email
   - Run sync
   - Verify reports appear in web UI
   - Check all statistics are accurate
   - Test filtering and pagination

8. Performance test:
   ```bash
   go test -bench=. ./internal/parser
   ```
   Verify parsing performance is acceptable

**Success Criteria**:
- Test coverage >80% for all packages
- Integration test covers full workflow
- README is complete and accurate
- All exported functions have documentation
- Makefile automates common tasks
- No linter errors
- Application works end-to-end with real IMAP server
- Performance is acceptable (can process 100 reports/second)

---

## TASK 12: Polish and Production Readiness

### Objectives
- Add proper error pages (404, 500)
- Implement request logging
- Add graceful shutdown
- Improve CSS styling
- Add security headers
- Create systemd service file

### Files to Create
1. `internal/web/middleware.go` - HTTP middleware (logging, recovery)
2. `internal/web/templates/error.html` - Error page template
3. `dmarc-viewer.service` - Systemd service file
4. Update `internal/web/templates/layout.html` - Improved CSS
5. `docs/deployment.md` - Deployment guide

### Implementation Requirements

**Middleware** (`internal/web/middleware.go`):
```go
func LoggingMiddleware(logger *zerolog.Logger) func(http.Handler) http.Handler
func RecoveryMiddleware(logger *zerolog.Logger) func(http.Handler) http.Handler
func SecurityHeadersMiddleware(next http.Handler) http.Handler

// Security headers:
// X-Content-Type-Options: nosniff
// X-Frame-Options: DENY
// X-XSS-Protection: 1; mode=block
// Content-Security-Policy: default-src 'self'
```

**Error Pages** (`internal/web/templates/error.html`):
```html
{{define "content"}}
<div class="error-page">
    <h1>{{.StatusCode}}</h1>
    <p>{{.Message}}</p>
    <a href="/">Return to Dashboard</a>
</div>
{{end}}
```

**Graceful Shutdown** (`cmd/dmarc-viewer/main.go`):
```go
// Handle SIGINT, SIGTERM
// Close database connection
// Stop HTTP server
// Wait for in-flight requests
// Timeout after 30 seconds
```

**Improved CSS** (clean, professional design):
- Responsive layout (mobile-friendly)
- Color scheme (blues and grays)
- Card-based design for stats
- Proper spacing and typography
- Loading indicators
- Hover effects

**Systemd Service** (`dmarc-viewer.service`):
```ini
[Unit]
Description=DMARC Report Viewer
After=network.target

[Service]
Type=simple
User=dmarc
WorkingDirectory=/opt/dmarc-viewer
ExecStart=/opt/dmarc-viewer/dmarc-viewer web --config /etc/dmarc-viewer/config.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

**Unit Tests**:
- Test middleware functions
- Test error page rendering
- Test graceful shutdown
- Test security headers

### What You Should See and Test

**After completing this task:**

1. Run unit tests: `go test ./... -v`
2. Build production binary:
   ```bash
   make build
   CGO_ENABLED=0 go build -ldflags="-w -s" -o dmarc-viewer cmd/dmarc-viewer/main.go
   ```
3. Start server and test graceful shutdown:
   ```bash
   ./dmarc-viewer web &
   PID=$!
   sleep 2
   kill -TERM $PID  # Should log "Shutting down gracefully..."
   ```
4. Test error pages:
   ```bash
   curl http://localhost:8080/nonexistent  # Should show 404 page
   ```
5. Verify request logging:
   - All HTTP requests logged with: method, path, status, duration
   - Structured JSON logs
6. Check security headers:
   ```bash
   curl -I http://localhost:8080/
   ```
   Verify headers: X-Content-Type-Options, X-Frame-Options, etc.
7. Test responsive design:
   - Open in browser
   - Resize window to mobile size
   - Verify layout adapts properly
8. Test on different browsers:
   - Chrome
   - Firefox
   - Safari
   - Verify HTMX works consistently
9. Load test:
   ```bash
   # Install `hey` or `ab` for load testing
   hey -n 1000 -c 10 http://localhost:8080/
   ```
   Verify performance is acceptable
10. Test systemd service (on Linux):
    ```bash
    sudo cp dmarc-viewer.service /etc/systemd/system/
    sudo systemctl daemon-reload
    sudo systemctl start dmarc-viewer
    sudo systemctl status dmarc-viewer
    ```

**Success Criteria**:
- All middleware works correctly
- Error pages render properly
- Graceful shutdown completes cleanly
- Security headers are present
- CSS is polished and responsive
- Logs are structured and informative
- Application handles load well (>100 req/s)
- Systemd service starts and stops correctly
- All tests pass

---

## Final Checklist

After completing all tasks, verify:

- [ ] Application starts and runs without errors
- [ ] Can connect to IMAP server and fetch reports
- [ ] Reports are parsed and stored correctly
- [ ] Web interface displays all data accurately
- [ ] Statistics calculations are correct
- [ ] HTMX provides smooth updates
- [ ] All configuration methods work (YAML, env, CLI)
- [ ] Test coverage >80% for all packages
- [ ] Documentation is complete
- [ ] No linter errors
- [ ] Application is production-ready
- [ ] End-to-end workflow tested with real IMAP account

## Estimated Timeline

- Task 1: 2-3 hours
- Task 2: 3-4 hours
- Task 3: 3-4 hours
- Task 4: 4-5 hours
- Task 5: 2-3 hours
- Task 6: 2-3 hours
- Task 7: 2-3 hours
- Task 8: 2-3 hours
- Task 9: 2-3 hours
- Task 10: 2-3 hours
- Task 11: 4-5 hours
- Task 12: 3-4 hours

**Total: ~30-40 hours of development time**
