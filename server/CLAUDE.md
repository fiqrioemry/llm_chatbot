# CLAUDE.md — Go Gin Server Architecture Guide

## Stack

Go + Gin · GORM (PostgreSQL) · Redis · MinIO · JWT (stateless AT) · Opaque RT (DB-stored, hashed) · OAuth2 · SMTP Mailer · OpenAI · VectorDB (pgvector)

---

## Folder Structure

```
/cmd/main.go                              # entry: load config, call lib.Init(), wire handlers, start server

/internal/
  config/
    config.go                             # struct Config, Load() — baca .env sekali di main
    constant/{auth,user,post,file}.go     # error message + error code strings

  lib/                                    # infrastructure init — dikembalikan dari lib.Init()
    init.go                               # Init(cfg) → (db, cache, minio, mailer, oauth)
    database.go                           # NewDatabase → *gorm.DB + AutoMigrate
    redis.go                              # NewRedis → *redis.Client
    storage.go                            # NewMinio → *minio.Client
    mailer.go                             # NewMailer → *Mailer (SMTP)
    oauth.go                              # NewOAuthRegistry → *OAuthRegistry (Google, GitHub)

  middleware/
    auth.go                               # Protect(cfg) — JWT verify; RequireRole(...); GetAuthUser(c)
    cors.go                               # CORS(cfg)
    error.go                              # ErrorHandler() — panic recovery; NotFound()
    file.go                               # file upload validation (size, type, field name)
    limiter.go                            # RateLimiter(cfg, cache, RateLimitConfig) — Redis incr/expire

  models/                                 # GORM structs — source of truth untuk skema DB
    user.go                               # User, UserStatus, UserRole, ProviderAccount, UserVerification, UserActivityLog
    session.go                            # Session
    file.go                               # FileStorage
    notification.go

  dto/                                    # Request + Response struct per domain (binding tags)
    auth.go                               # RegisterRequest, LoginRequest, LoginResponse, TokenPair, MeResponse, SessionResponse, ...
    user.go
    product.go

  repositories/                           # DB layer — interface + impl (GORM)
    auth.go                               # AuthRepositoryContract interface + AuthRepository struct
    user.go
    product.go

  services/                               # Business logic — inject repo + infra
    auth.go                               # AuthService: Register, Login, Logout, RefreshToken, OAuth, ...
    user.go
    product.go

  handlers/                               # HTTP layer — bind req, call service, return response
    auth.go                               # AuthHandler; NewAuthHandler(svc, cfg)
    user.go
    product.go
    handlers.go                           # struct Handlers { Auth *AuthHandler; User *UserHandler; ... }

  routes/
    router.go                             # InitRoutes(engine, cfg, handlers) — health, NoRoute, mount v1
    auth.go                               # InitAuthRoutes(rg, cfg, handlers)
    user.go

  # ── INTER-SERVICE COMMUNICATION ──────────────────────────────────────────────
  clients/                                # HTTP clients untuk panggil internal service lain
    base.go                               # newHTTPClient(baseURL, timeout) — *resty.Client wrapper
    auth/
      client.go                           # AuthClient: ValidateToken(token) → *AuthClaims, GetUser(id)
      types.go                            # request/response struct khusus inter-service auth
    notification/
      client.go                           # NotificationClient: Send(userID, payload)
      types.go
    # proto/                              # (future) gRPC stubs — generate dari .proto, taruh sini

  # ── THIRD-PARTY API CLIENTS ──────────────────────────────────────────────────
  providers/                              # Wrapper untuk external API (bukan infra, bukan OAuth)
    openai/
      client.go                           # OpenAIClient: Chat(msgs) → string, Embed(text) → []float32
      types.go
    gemini/
      client.go
      types.go
    maps/
      client.go                           # MapsClient: Geocode(addr), DistanceMatrix(origins, dests)
      types.go
    midtrans/
      client.go                           # MidtransClient: Charge(req), GetStatus(orderID), Refund(id)
      types.go

  workers/                                # Background goroutines + cron jobs
    email.go                              # async email queue consumer
    file_cleanup.go                       # hapus file orphan dari MinIO
    data_sync.go

  ws/                                     # WebSocket
    hub.go                                # Register/unregister client, broadcast
    client.go                             # per-connection read/write pump
    event.go                              # event type definitions

/pkg/                                     # zero-dependency utilities — tidak boleh import /internal
  cache/cache.go                          # Redis wrapper: Get, Set, Del, Incr, Expire, Exists
  crypto/crypto.go                        # HashPassword, VerifyPassword, GenerateOpaqueToken, HashToken (SHA256)
  jwt/jwt.go                              # GenerateAccessToken, VerifyAccessToken
  response/response.go                    # OK, Created, Error, HandleError, AppError, PaginationMeta
  validator/validator.go                  # Init(), ExtractErrors(err) → []FieldError
  pagination/pagination.go               # ParseQuery(c) → PaginationParams
  utils/
    generator.go                          # NewUUID, GenerateUsername, GenerateAvatarURL, RandomString
    string.go
    date.go

/tests/                                   # testing
  integration/                            # test yang butuh DB/Redis nyala (pakai testcontainers)
    auth_test.go
    user_test.go
  unit/                                   # pure unit test, mock repo/service
    services/
      auth_test.go
    pkg/
      crypto_test.go
  mocks/                                  # generated mock — pakai mockery atau manual
    auth_repository_mock.go
    auth_service_mock.go

/.env
/.env.example
/go.mod · /go.sum
/Makefile                                 # make run | test | build | mock | migrate
```

---

## Layer Rules

```
handler → service → repository → model
handler → pkg/response (langsung)
service → pkg/{crypto,jwt,cache,utils}
service → clients/  (inter-service HTTP)
service → providers/ (third-party API)
pkg/    → tidak boleh import internal/
```

---

## Patterns

### Error handling

```go
// service — return AppError, bukan raw error
if user == nil {
    return nil, response.NotFoundErr(constant.ErrUserNotFound, constant.CodeUserNotFound)
}

// handler — satu baris, tidak perlu switch
if err != nil {
    response.HandleError(c, err)
    return
}
```

### Repository contract

```go
// Selalu ada interface untuk tiap repo — memudahkan mock di test
type AuthRepositoryContract interface {
    FindUserByEmail(email string) (*models.User, error)
    Transaction(fn func(tx *gorm.DB) error) error
    // ...
}
```

### Inter-service call (`clients/`)

```go
// clients/auth/client.go
type AuthClient struct { base *resty.Client }

func (c *AuthClient) ValidateToken(token string) (*AuthClaims, error) {
    // POST /internal/auth/validate
    // return AppError jika gagal supaya handler bisa HandleError
}
```

### Third-party provider (`providers/`)

```go
// providers/openai/client.go
type OpenAIClient struct { apiKey string; http *resty.Client }

func (c *OpenAIClient) Chat(msgs []Message) (string, error) { ... }
// Wrap error ke response.InternalErr / response.BadRequestErr
```

### DTO per domain

```go
// Pisahkan request vs response — jangan pakai model GORM langsung ke handler
type RegisterRequest struct {
    Fullname string `json:"fullname" binding:"required,min=1,max=100"`
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=8,max=72"`
}
```

### Background job (workers)

```go
// Gunakan goroutine untuk side-effect non-blocking
go func() {
    link := fmt.Sprintf("%s/verify-email?token=%s", cfg.ClientURL, rawToken)
    mailer.SendVerificationLink(email, subject, link)
}()
// Worker long-running → jalankan di main.go sebelum server start
```

---

## Do

- Selalu return **AppError** dari service, bukan raw `errors.New`
- Selalu wrap DB ops dalam **Transaction** kalau ada >1 write
- **Interface** untuk semua repository — wajib untuk testability
- **DTO** terpisah dari model GORM — jangan expose struct DB ke response
- `clients/` untuk **internal service call** — `providers/` untuk **external API**
- Inject semua dependency lewat constructor (`NewXxx(dep1, dep2)`)
- Gunakan **goroutine** hanya untuk side-effect (email, log) — bukan untuk flow utama
- Test di `/tests/` — unit test mock repo/service, integration test pakai real DB

## Don't

- Jangan import `internal/` dari `pkg/`
- Jangan panggil `os.Exit` atau `log.Fatal` di luar `main.go` / `lib/`
- Jangan return model GORM langsung dari handler — selalu map ke DTO
- Jangan hardcode string error/code — pakai `constant/`
- Jangan buat logic bisnis di handler — handler hanya: bind → call service → respond
- Jangan buat HTTP call ke service lain langsung dari service — gunakan `clients/`
- Jangan skip interface untuk repository meski module kecil

---

## Logging Rules

Gunakan `log.Printf` standar Go. Semua log wajib ada **prefix layer** supaya mudah di-grep.

### Format prefix

```
[layer] pesan
```

| Layer      | Prefix                                           |
| ---------- | ------------------------------------------------ |
| Handler    | `[auth handler]`, `[user handler]`               |
| Service    | `[auth service]`, `[user service]`               |
| Repository | `[auth repo]` (opsional, hanya untuk slow query) |
| Worker     | `[worker:file_cleanup]`, `[worker:email]`        |
| Middleware | `[middleware:auth]`, `[middleware:limiter]`      |
| Infra/lib  | `[db]`, `[redis]`, `[minio]`, `[mailer]`         |

### Wajib di-log

**Handler** — request masuk + hasil akhir:

```go
log.Printf("[auth handler] Login request — email: %s", req.Email)
log.Printf("[auth handler] Login success — email: %s", req.Email)
// error tidak perlu log di handler — sudah di HandleError
```

**Service** — keputusan bisnis penting:

```go
log.Printf("[auth service] Register — user created: %s", userID)
log.Printf("[auth service] Login — invalid credentials: %s", email)
log.Printf("[auth service] RefreshToken — session rotated: %s", sessionID)
log.Printf("[auth service] OAuthCallback — new user created via %s: %s", provider, userID)
```

**HandleError** — unexpected error (sudah ada, jangan duplikat):

```go
// pkg/response/response.go — sudah log di sini untuk non-AppError
log.Printf("[ERROR] unhandled error on %s %s: %v", method, path, err)
```

**Worker** — lifecycle job:

```go
log.Printf("[worker:file_cleanup] started")
log.Printf("[worker:file_cleanup] deleted %d orphan files", n)
log.Printf("[worker:file_cleanup] error: %v", err)
```

**Infra (lib/)** — sudah ada, pertahankan pola `✅` / `❌`:

```go
log.Println("✅ MySQL connected and migrated")
log.Fatalf("❌ Failed to connect to MySQL: %v", err)
```

### Jangan di-log

- Password, token, hash — **jangan pernah**
- Setiap DB query (sudah ditangani GORM logger di mode development)
- Successful no-op (misal resend email ke user tidak ditemukan — intentional silent)

---

## File naming convention

| Layer                  | Pattern                  | Contoh                    |
| ---------------------- | ------------------------ | ------------------------- |
| Model                  | noun                     | `user.go`                 |
| DTO                    | noun                     | `auth.go`                 |
| Repository             | noun                     | `auth.go`                 |
| Service                | noun                     | `auth.go`                 |
| Handler                | noun                     | `auth.go`                 |
| Client (inter-service) | `client.go` + `types.go` | `clients/auth/`           |
| Provider (3rd party)   | `client.go` + `types.go` | `providers/openai/`       |
| Worker                 | `noun_verb.go`           | `file_cleanup.go`         |
| Test                   | `*_test.go`              | `auth_test.go`            |
| Mock                   | `*_mock.go`              | `auth_repository_mock.go` |

x
