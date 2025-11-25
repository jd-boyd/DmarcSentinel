# DMARC Report Viewer - Design Documentation

## Project Overview
A locally-hosted Go application for viewing DMARC RUA (aggregate) and RUF (forensic) reports. The application will extract reports from an IMAP server, store them in a database, and provide a web interface for viewing and analyzing the reports.

## Core Requirements

### 1. Report Extraction
- Connect to IMAP server
- Extract DMARC reports from a specified folder
- Support both RUA (XML aggregate reports) and RUF (forensic reports)
- Parse compressed attachments (gzip, zip)
- Track downloaded reports to avoid re-downloading
- Store reports in database

### 2. Data Storage
- Database for storing:
  - Report metadata (date, source, status)
  - Full report content
  - Download tracking (to prevent duplicates)
  - Parsed report data for quick querying
- Consider SQLite for simplicity (single file, no external dependencies)

### 3. Web Interface
- List view: Reports sorted by time (newest first)
- Detail view: Full report display when clicked
- Statistics dashboard:
  - Failure counts (last day/week/month)
  - Pass/fail ratios
  - Top failing sources
  - DKIM/SPF/DMARC alignment statistics

### 4. Configuration
- Support three configuration methods (priority order):
  1. Command line flags (highest priority)
  2. Environment variables (medium priority)
  3. YAML config file (lowest priority)
- Configuration items:
  - IMAP server details (host, port, username, password, folder)
  - Database path
  - Web server port
  - Sync interval
  - Log level

### 5. Technology Constraints
- **Pure Go libraries only** (no CGO dependencies)
- HTMX for frontend interactivity
- Emphasis on testability with comprehensive unit tests

## Technical Architecture

### Components

#### 1. IMAP Client Module
- **Purpose**: Connect to IMAP server and fetch emails
- **Pure Go Library**: `github.com/emersion/go-imap` (most popular, pure Go)
- **Responsibilities**:
  - Authenticate with IMAP server
  - Search for DMARC report emails
  - Download email attachments
  - Extract compressed files (gzip, zip)
  - Track message UIDs for download state

#### 2. Report Parser Module
- **Purpose**: Parse DMARC RUA (XML) and RUF reports
- **Pure Go Library**: `encoding/xml` (standard library)
- **Responsibilities**:
  - Validate XML structure
  - Parse aggregate reports (RUA)
  - Parse forensic reports (RUF)
  - Extract relevant metadata
  - Handle malformed reports gracefully

#### 3. Database Module
- **Purpose**: Store and retrieve report data
- **Pure Go Library**: `modernc.org/sqlite` (pure Go SQLite implementation)
- **Schema**:
  ```
  tables:
    - reports: id, message_uid, report_type, org_name, report_id, date_begin, date_end, email, domain, created_at
    - report_records: id, report_id, source_ip, count, disposition, dkim_result, spf_result
    - download_state: message_uid, folder, downloaded_at
    - statistics_cache: metric_name, time_period, value, calculated_at
  ```

#### 4. Configuration Module
- **Purpose**: Load and merge configuration from multiple sources
- **Pure Go Libraries**:
  - `github.com/spf13/viper` (configuration management)
  - `github.com/spf13/pflag` (POSIX/GNU style flags)
  - `gopkg.in/yaml.v3` (YAML parsing)
- **Priority**: CLI flags > Environment variables > YAML file

#### 5. Web Server Module
- **Purpose**: Serve web interface and API endpoints
- **Pure Go Libraries**:
  - `net/http` (standard library)
  - `github.com/go-chi/chi` (lightweight router)
  - `html/template` (standard library for templates)
- **Endpoints**:
  - `GET /` - Dashboard with statistics
  - `GET /reports` - List of reports (with pagination)
  - `GET /reports/{id}` - Detailed report view
  - `GET /api/stats` - JSON statistics API
  - `POST /sync` - Manual sync trigger

#### 6. Statistics Module
- **Purpose**: Calculate and cache statistics
- **Responsibilities**:
  - Count failures by time period
  - Calculate pass/fail ratios
  - Identify top failing sources
  - Cache results for performance
  - Provide aggregated metrics

### Data Flow

1. **Sync Process**:
   ```
   IMAP Server → IMAP Client → Download State Check → New Messages
   → Extract Attachments → Decompress → Parser → Database
   ```

2. **Web Request Flow**:
   ```
   Browser → HTTP Router → Handler → Database Query → Template Rendering
   → HTMX Response → Browser DOM Update
   ```

3. **Statistics Flow**:
   ```
   Scheduled Task → Statistics Module → Database Aggregation
   → Cache Update → Web Display
   ```

## HTMX Integration

### Why HTMX
- Minimal JavaScript required
- Server-side rendering with dynamic updates
- Progressive enhancement
- Simple mental model

### HTMX Patterns to Use

1. **Report List**:
   - Infinite scroll with `hx-get` and `hx-trigger="revealed"`
   - Click to load detail with `hx-get="/reports/{id}" hx-target="#detail-pane"`

2. **Statistics**:
   - Auto-refresh with `hx-get="/api/stats" hx-trigger="every 30s"`
   - Show loading indicator with `hx-indicator`

3. **Manual Sync**:
   - Button with `hx-post="/sync" hx-swap="outerHTML"`
   - Show progress with server-sent events (SSE)

## Testing Strategy

### Unit Tests
- Each module should have 80%+ code coverage
- Test files: `*_test.go` alongside source files
- Use table-driven tests for multiple scenarios
- Mock external dependencies (IMAP, database)

### Test Categories

1. **IMAP Client Tests**:
   - Mock IMAP server responses
   - Test connection handling
   - Test message parsing
   - Test download state tracking

2. **Parser Tests**:
   - Sample RUA/RUF XML files
   - Malformed XML handling
   - Edge cases (empty reports, missing fields)

3. **Database Tests**:
   - In-memory SQLite for tests
   - CRUD operations
   - Query performance
   - Transaction handling

4. **Configuration Tests**:
   - Priority ordering (CLI > ENV > YAML)
   - Missing configuration handling
   - Invalid values
   - Default values

5. **Web Handler Tests**:
   - HTTP response codes
   - Template rendering
   - HTMX header handling
   - Pagination logic

6. **Statistics Tests**:
   - Calculation accuracy
   - Time period filtering
   - Cache invalidation
   - Edge cases (no data)

## Configuration Schema

### YAML Example (`config.yaml`)
```yaml
imap:
  host: imap.example.com
  port: 993
  username: user@example.com
  password: secret
  folder: INBOX.DMARC
  use_tls: true

database:
  path: ./dmarc-reports.db

web:
  port: 8080
  host: localhost

sync:
  interval: 15m
  on_startup: true

logging:
  level: info
  format: json
```

### Environment Variables
```
DMARC_IMAP_HOST=imap.example.com
DMARC_IMAP_PORT=993
DMARC_IMAP_USERNAME=user@example.com
DMARC_IMAP_PASSWORD=secret
DMARC_IMAP_FOLDER=INBOX.DMARC
DMARC_DATABASE_PATH=./dmarc-reports.db
DMARC_WEB_PORT=8080
DMARC_LOG_LEVEL=info
```

### Command Line Flags
```bash
./dmarc-viewer \
  --imap-host imap.example.com \
  --imap-port 993 \
  --imap-username user@example.com \
  --imap-password secret \
  --database ./dmarc-reports.db \
  --web-port 8080 \
  --log-level info
```

## Security Considerations

1. **Credentials**:
   - Never log passwords
   - Support reading password from file
   - Warn if credentials in config file are world-readable

2. **Web Interface**:
   - Consider adding basic auth option
   - Bind to localhost by default
   - Sanitize all HTML output

3. **Database**:
   - Use parameterized queries (prevent SQL injection)
   - Regular backups recommended
   - File permissions on database

## Performance Considerations

1. **IMAP Sync**:
   - Use IMAP IDLE for push notifications (optional)
   - Batch processing of messages
   - Concurrent download of attachments (with limit)

2. **Database**:
   - Index on frequently queried fields (date, domain, source_ip)
   - Archive old reports (optional feature)
   - Vacuum database periodically

3. **Web Interface**:
   - Pagination for large result sets
   - Cache statistics calculations
   - Compress HTTP responses

## Future Enhancements (Out of Scope for Initial Version)

1. Email notifications for failures
2. Export reports to CSV/PDF
3. Multi-domain support
4. API for external integrations
5. Docker container
6. Alerting rules engine
7. Comparison with previous time periods
8. Whitelisting known good sources

## Project Structure

```
dmarc-viewer/
├── cmd/
│   └── dmarc-viewer/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go              # Configuration management
│   │   └── config_test.go
│   ├── imap/
│   │   ├── client.go              # IMAP client
│   │   ├── client_test.go
│   │   └── state.go               # Download state tracking
│   ├── parser/
│   │   ├── rua.go                 # RUA parser
│   │   ├── rua_test.go
│   │   ├── ruf.go                 # RUF parser
│   │   └── ruf_test.go
│   ├── database/
│   │   ├── db.go                  # Database operations
│   │   ├── db_test.go
│   │   ├── migrations.go          # Schema migrations
│   │   └── models.go              # Data models
│   ├── stats/
│   │   ├── calculator.go          # Statistics calculation
│   │   └── calculator_test.go
│   └── web/
│       ├── server.go              # HTTP server
│       ├── server_test.go
│       ├── handlers.go            # HTTP handlers
│       ├── handlers_test.go
│       └── templates/             # HTML templates
│           ├── layout.html
│           ├── dashboard.html
│           ├── report_list.html
│           └── report_detail.html
├── testdata/
│   ├── sample_rua.xml             # Test fixtures
│   └── sample_ruf.xml
├── config.yaml.example            # Example configuration
├── go.mod
├── go.sum
├── README.md
└── Makefile                       # Build automation
```

## Dependencies (Pure Go)

```
github.com/emersion/go-imap         # IMAP client
github.com/emersion/go-message      # Email message parsing
modernc.org/sqlite                  # Pure Go SQLite
github.com/spf13/viper             # Configuration
github.com/spf13/pflag             # CLI flags
github.com/go-chi/chi              # HTTP router
gopkg.in/yaml.v3                   # YAML parsing
github.com/rs/zerolog              # Structured logging
```

## Success Criteria

The application will be considered successful when:

1. It can connect to an IMAP server and download DMARC reports
2. It correctly parses RUA XML reports
3. It stores reports in the database without duplicates
4. The web interface displays a list of reports
5. Users can click on a report to see details
6. Statistics show accurate failure counts
7. Configuration works via CLI, environment, and YAML
8. Unit tests achieve >80% coverage
9. The application runs without external dependencies (except IMAP server)
10. HTMX provides smooth, dynamic updates without full page reloads
