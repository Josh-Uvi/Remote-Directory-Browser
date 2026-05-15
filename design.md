| Key | Value |
| --- | --- |
| authors | Engineering Team |
| state | draft |

# RFD - Remote Directory Browser

## What

A web application that allows authenticated users to securely browse directory contents on a remote server. The application emphasizes strong authentication, encryption in transit (TLS), and client-side filtering/sorting capabilities with URL-based navigation to maintain application state across page refreshes.

## Why

Organizations need a simple, secure way to explore remote filesystems through a web interface. This POC demonstrates security best practices including:
- End-to-end encryption (TLS)
- Strong authentication with session management
- Protection against common web vulnerabilities
- Client-side performance optimization to reduce server round-trips

This design showcases engineering practices aligned with security-first organizations like Teleport, proving that simple UX doesn't require sacrificing security posture.

## Details

### UX

#### User Stories

**Story 1: First-time User Logs In**

Alex is a new user who needs to browse a project directory on a shared server. They navigate to the application, see a login page with username and password fields. They enter credentials (`admin` / `admin123`), click "Login", and are redirected to the home directory view.

Expected flow:
```
1. User visits https://localhost:8443/
2. Unauthenticated → redirected to /login
3. User enters credentials and submits
4. API validates credentials and creates session cookie
5. User is redirected to /files (home directory)
6. Directory contents load and display
```

**Story 2: Navigate Directory Structure**

Once logged in, Alex sees a list of files and subdirectories. They click on a subdirectory named "Projects" in the table. The breadcrumb updates to show "Home > Projects", the URL changes to `/files/Projects`, and the table reloads with contents of that directory.

Expected interactions:
```
- Click on row: Navigate to subdirectory (if type=dir)
- Click breadcrumb link: Navigate to parent directory
- URL bar: User can bookmark and return to specific directory
- Page refresh: State is preserved via URL, no data loss
```

**Story 3: Filter and Sort Contents**

While browsing a directory with many files, Alex types "config" in the search box. The table instantly filters to show only items matching "config". Alex then clicks "Sort by: Size" to organize results by file size. All changes happen client-side without server requests.

Expected behavior:
```
- Filter: Real-time substring matching, case-insensitive
- Sort: By Name (A-Z, Z-A), Type (directories first), Size (ascending/descending)
- Query params capture state: ?sort=size&order=desc&filter=config
```

**Story 4: Session Timeout and Logout**

After browsing for a while, Alex clicks the "Logout" button in the top-right corner. The session is destroyed server-side, the browser cookie is cleared, and Alex is redirected to the login page. If Alex tries to access `/files` directly, they are redirected back to login.

Expected behavior:
```
- POST /api/logout → Session destroyed
- Redirect to /login
- Direct navigation to /files without session → Redirected to /login
```

#### Wireframes

**Main Application View**
```
┌─────────────────────────────────────────────────────────────┐
│  Remote Directory Browser                          [Logout] │
├─────────────────────────────────────────────────────────────┤
│ Breadcrumb: Home > Documents > Projects > MyApp             │
├─────────────────────────────────────────────────────────────┤
│ Filter: [Search...]     Sort by: [Name ▼]                   │
├─────────────────────────────────────────────────────────────┤
│ NAME              TYPE      SIZE           MODIFIED         │
├─────────────────────────────────────────────────────────────┤
│ 📁 src            dir       -              May 10, 2026      │
│ 📁 tests          dir       -              May 8, 2026       │
│ 📄 README.md      file      2.3 KB         May 15, 2026      │
│ 📄 package.json   file      1.1 KB         May 12, 2026      │
│ 📄 config.yaml    file      456 B          May 5, 2026       │
└─────────────────────────────────────────────────────────────┘
```

**Login Page**
```
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│            Remote Directory Browser                          │
│                                                              │
│            ┌──────────────────────────────┐                  │
│            │ Login                        │                  │
│            ├──────────────────────────────┤                  │
│            │ Username: [________]         │                  │
│            │ Password: [________]         │                  │
│            │                              │                  │
│            │ [Login]                      │                  │
│            └──────────────────────────────┘                  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

#### Components

The UI is composed of lightweight, single-purpose components:

**Breadcrumb Navigation**
- Clickable links to navigate to parent directories
- Shows the full path of the current location
- Implementation: Plain HTML with click handlers, no framework

**Directory Table**
- Sortable columns: Name, Type, Size, Modified
- Rows for each file/directory with icons
- Implementation: HTML table with client-side event listeners

**Search/Filter Bar**
- Single input field with real-time filtering
- Filters displayed items by substring match
- Implementation: Input listener that filters the in-memory list

**Sort Dropdown**
- Options: Name (A-Z, Z-A), Type, Size (ascending/descending)
- Updates table without server call
- Implementation: Plain `<select>` element

**Logout Button**
- Top-right corner, clears session and redirects to login
- Implementation: Form POST or fetch call to `/api/logout`

#### Day One vs. Day Two Experience

**Day One (First-time user):**
- Minimal UI with clear CTAs (Call-to-Actions)
- Intuitive login flow
- Immediate access to home directory with sample content
- Clear error messages (e.g., "Invalid username or password")

**Day Two (Power user):**
- Ability to bookmark URLs for quick access to specific directories
- Advanced filtering and sorting for large directories
- Understanding of URL structure to share directory links

### API

#### GET /api/list?path=/path/to/directory

Returns the contents of a specified directory. Requires valid session.

**Request:**
```
GET /api/list?path=/Users/alice/Documents
Cookie: session=abc123xyz789
```

**Response (200 OK):**
```json
{
  "name": "Documents",
  "type": "dir",
  "size": 0,
  "path": "/Users/alice/Documents",
  "contents": [
    {
      "name": "README.md",
      "type": "file",
      "size": 2048,
      "modified": "2026-05-15T10:30:00Z"
    },
    {
      "name": "Projects",
      "type": "dir",
      "size": 0,
      "modified": "2026-05-10T14:20:00Z"
    }
  ]
}
```

**Error Responses:**
- `401 Unauthorized`: Missing or invalid session
- `400 Bad Request`: Invalid path parameter
- `403 Forbidden`: Path outside allowed directory or permission denied
- `500 Internal Server Error`: Server error

#### POST /api/login

Authenticates user and creates session.

**Request:**
```json
POST /api/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Login successful"
}
```

Sets secure HTTP-only cookie: `Set-Cookie: session=<token>; HttpOnly; Secure; SameSite=Strict`

**Error Response (401 Unauthorized):**
```json
{
  "success": false,
  "error": "Invalid username or password"
}
```

#### POST /api/logout

Destroys user session.

**Request:**
```
POST /api/logout
Cookie: session=abc123xyz789
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

Clears session cookie.

### URL Structure

Navigation state is encoded in the URL for bookmarkability and refresh resilience:

```
https://localhost:8443/                    # Root (redirects to /login if not authenticated)
https://localhost:8443/login               # Login page
https://localhost:8443/files               # Home directory
https://localhost:8443/files/Documents     # Subdirectory
https://localhost:8443/files/Documents/Projects  # Nested directory

# Query parameters for UI state (optional)
https://localhost:8443/files?sort=name&order=asc&filter=test
```

**Design rationale:**
- Path-based navigation (`/files/*`) is the source of truth for which directory to display
- Query parameters capture transient UI state (sort/filter) without affecting bookmarkability
- Page refresh preserves all state via URL
- Clean, familiar URL pattern for users

### AI Proofing

#### CLI UX (API Consumption)

While the primary interface is a web UI, the API is designed to be agent-friendly:

**Structured JSON responses:**
All API endpoints return machine-readable JSON with consistent structure, enabling agents to parse and act on responses reliably.

**Clear error messages:**
Error responses include actionable information:
```json
{
  "success": false,
  "error": "Path traversal detected: cannot access /etc/passwd",
  "details": "Requested path is outside allowed root directory"
}
```

**Idempotent logout:**
Calling `/api/logout` multiple times returns success, enabling robust retry logic.

**No interactive prompts:**
API accepts all required parameters via request body or query parameters; no interactive prompts block automated workflows.

#### Future Agent Skills

A reference agent skill could demonstrate:
1. Authenticating to the application
2. Discovering directory structure
3. Filtering and sorting contents programmatically
4. Handling error conditions gracefully

### Security

#### Authentication

**Session management:**
- Session tokens: Cryptographically random 32+ byte strings using `crypto/rand`
- Tokens stored in-memory map: `map[token]SessionData{username, createdAt, expiresAt}`
- Session expiry: 1 hour of inactivity; automatic cleanup of expired sessions
- Logout: Immediate session destruction and cookie clearing

**Hard-coded test users (demo only):**
- Username: `admin` / Password: `admin123`
- Username: `user` / Password: `password`
- Production: Use strong password hashing (bcrypt) and integrate with external auth system

**Session cookies:**
```
Set-Cookie: session=<token>; 
  HttpOnly;              # Prevent JavaScript access (XSS mitigation)
  Secure;                # HTTPS-only transmission
  SameSite=Strict;       # CSRF mitigation
  Max-Age=3600;          # 1-hour expiration
  Path=/;                # Available across entire app
```

#### Encryption in Transit (TLS)

**Certificate generation (development):**
```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes
```

**TLS enforcement:**
- All traffic served over HTTPS (port 8443)
- HTTP requests redirected to HTTPS
- Minimum TLS 1.2 enforced
- Strong cipher suites preferred

**HSTS (HTTP Strict-Transport-Security):**
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

Instructs browsers to always use HTTPS, preventing downgrade attacks.

#### Protection Against Common Web Vulnerabilities

| Vulnerability | Attack Vector | Mitigation |
| --- | --- | --- |
| **CSRF** | Forged cross-site requests | SameSite=Strict cookies; Origin/Referer validation |
| **XSS** | JavaScript injection | HTML entity escaping; Content-Security-Policy headers |
| **Path Traversal** | `../../etc/passwd` | Validate path is within allowed root; use `filepath.Abs()` and comparison |
| **Session Hijacking** | Cookie theft | Secure + HttpOnly flags; HTTPS-only; token rotation on re-login |
| **Clickjacking** | Embedding app in malicious frame | X-Frame-Options: DENY header |
| **Information Disclosure** | Detailed error messages | Generic error messages; detailed logs server-side only |
| **Brute Force** | Many login attempts | Rate limiting on `/api/login` (e.g., 5 attempts per minute per IP) |
| **SQL Injection** | N/A | No database; filesystem operations use validated paths only |

**Content-Security-Policy (CSP) header:**
```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'
```

**Path Traversal Protection:**
```go
// Pseudocode
allowedRoot := "/Users/alice"
requestedPath := filepath.Join(allowedRoot, userInput)
cleanedPath := filepath.Abs(requestedPath)

if !strings.HasPrefix(cleanedPath, allowedRoot) {
    return error("Path traversal detected")
}
// Safe to use cleanedPath
```

### Privacy

**Data Collection:**
- No personal data collected beyond username for session management
- No analytics or telemetry
- No tracking of file access patterns (no audit logs to external systems)

**Data Retention:**
- Session tokens stored in-memory; cleared on logout or expiry
- No persistent storage beyond application lifetime
- Server restart clears all session data

**Data Access:**
- Sessions accessible only to authenticated requests bearing valid token
- API validates session before returning directory contents
- No other users can access sessions or directory contents

**Recommendations for production:**
- Implement audit logging of directory access with PII masking
- Document data retention policies
- Consider encryption at rest for filesystem backups

### Scale

#### Current Limitations

**Directory listing:**
- API returns all files in a directory without pagination
- Suitable for directories with hundreds of files
- For larger directories (thousands+ of files), pagination should be added

**In-memory sessions:**
- Stores all active sessions in application memory
- Cannot scale horizontally across multiple server instances
- Suitable for single-server deployment or internal POC

#### Optimization Opportunities

**Client-side filtering/sorting:**
- All filtering and sorting happens in the browser after initial fetch
- Reduces server load; increases interactivity
- One API call per directory navigation

**Potential UI bottlenecks:**
- Rendering large tables (1000+ rows) may be slow in older browsers
- Virtual scrolling could improve performance for very large directories

#### Recommendations for Scale

- Add pagination to directory listing API: `GET /api/list?path=/dir&limit=100&offset=0`
- Use database for session storage to support horizontal scaling
- Implement server-side sorting/filtering with index support
- Add caching headers for immutable directory content (Cache-Control)

### Backward Compatibility

**Not applicable.** This is a new application with no existing users or versions.

**Future considerations:**
- API versioning (`/api/v1/list`, `/api/v2/list`) to support client updates
- Graceful session migration if moving from in-memory to database storage

### Audit Events

**Recommended audit events to log (server-side, not exposed via API):**

```json
{
  "event_type": "LOGIN_SUCCESS",
  "timestamp": "2026-05-15T10:30:00Z",
  "username": "admin",
  "ip_address": "192.168.1.100"
}

{
  "event_type": "LOGIN_FAILURE",
  "timestamp": "2026-05-15T10:30:05Z",
  "username": "unknown",
  "ip_address": "192.168.1.101",
  "reason": "Invalid password"
}

{
  "event_type": "DIRECTORY_ACCESS",
  "timestamp": "2026-05-15T10:31:00Z",
  "username": "admin",
  "path": "/Users/alice/Documents",
  "ip_address": "192.168.1.100"
}

{
  "event_type": "LOGOUT",
  "timestamp": "2026-05-15T10:35:00Z",
  "username": "admin",
  "ip_address": "192.168.1.100"
}
```

### Observability

**Key metrics to monitor:**

```
- app_login_attempts_total (counter) - Total login attempts
- app_login_failures_total (counter) - Failed login attempts by reason
- app_directory_access_total (counter) - Total directory access requests
- app_request_duration_seconds (histogram) - Request latency by endpoint
- app_active_sessions (gauge) - Current number of active sessions
- app_tls_handshake_duration_seconds (histogram) - TLS handshake performance
```

**Logging strategy:**
- Structured JSON logs with timestamp, level, message, and context
- Errors include stack traces and request IDs for debugging
- Info-level logs for successful operations (login, logout, directory access)
- Debug-level logs for internal state transitions

**Example log lines:**
```
{"timestamp":"2026-05-15T10:30:00Z","level":"INFO","message":"Login successful","username":"admin","ip":"192.168.1.100"}
{"timestamp":"2026-05-15T10:31:00Z","level":"INFO","message":"Directory listed","path":"/Users/alice/Documents","count":5,"duration_ms":12}
```

### Product Usage

**Metrics to track adoption:**

```
- Daily Active Users (DAU) - Unique users per day
- Login success rate - (Successful logins) / (Total login attempts)
- Session duration - Average time from login to logout
- Most accessed paths - Top 10 directories visited
- Features used - Sorting/filtering usage frequency
```

**Collection method:**
- Application metrics exported via `/metrics` endpoint (Prometheus format)
- No client-side analytics or third-party tracking

### Test Plan

#### Unit Tests

**Backend:**
- Session token generation uniqueness and randomness
- Path traversal validation logic
- File size formatting and human-readable output
- Login/logout state transitions

**Frontend:**
- Filter logic (substring matching, case-sensitivity)
- Sort logic (by name, type, size)
- Breadcrumb path parsing and navigation
- URL state encoding/decoding

#### Integration Tests

- Login flow: valid credentials → session created → cookie set
- Logout flow: valid session → session destroyed → redirect to login
- Directory listing: valid path → correct contents returned; invalid path → 403 error
- Authentication required: unauthenticated request → 401 response
- Session expiry: access with expired session → 401 response

#### End-to-End Tests

```
Scenario 1: Complete user journey
1. Start at login page
2. Submit invalid credentials → error message
3. Submit valid credentials → redirected to /files
4. Navigate to subdirectory → URL updates, contents load
5. Filter by filename → table updates client-side
6. Sort by size → table re-orders without API call
7. Click logout → redirected to login page
8. Try to access /files without session → redirected to login

Scenario 2: Session security
1. Login and get session cookie
2. Extract session token and use in another browser/incognito window → works
3. Logout → session destroyed
4. Retry with saved session token → 401 Unauthorized

Scenario 3: Path traversal protection
1. Attempt: GET /api/list?path=../../etc/passwd → 403 Forbidden
2. Attempt: GET /api/list?path=/Users/bob → 403 Forbidden (different user)
3. Success: GET /api/list?path=/Users/alice/Documents → 200 OK
```

#### Cross-browser Testing

- Chrome/Chromium
- Firefox
- Safari
- Edge

#### Security Testing

- SSL/TLS certificate validity
- HTTPS enforcement (HTTP → HTTPS redirect)
- Secure cookie flags (HttpOnly, Secure, SameSite)
- HSTS header presence
- X-Frame-Options header presence
- CSP policy effectiveness


