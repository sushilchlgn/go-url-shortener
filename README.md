# go-url-shortener

A tiny URL shortener: Go serverless functions on Netlify (Lambda-compatible
runtime) + Postgres for storage.

```
POST /api/shorten   { "url": "https://..." }  -> { "code", "short_url" }
GET  /r/<code>                                -> 302 redirect to original URL
```

## How it's wired

- `netlify/functions/shorten/main.go` — creates a random 6-char code, inserts
  it into Postgres, returns the short URL.
- `netlify/functions/redirect/main.go` — looks up a code, 302-redirects,
  increments a click counter.
- `internal/store` — the Postgres connection + queries (shared by both
  functions).
- `internal/shortcode` — crypto/rand-based code generator.
- `build.sh` — compiles each function directory into a Linux binary under
  `functions-build/`, which is what Netlify's Go (Lambda) runtime expects.
- `netlify.toml` — build command + clean URL redirects (`/api/shorten`,
  `/r/*`) so callers never see the raw `/.netlify/functions/...` paths.
- `public/index.html` — minimal terminal-styled test page.

## 1. Get a free Postgres database

Any of these work — you just need a connection string:

- **Neon** (neon.tech) — free tier, fastest to set up, serverless-friendly
- **Supabase** (supabase.com) — free tier, includes a UI for browsing data
- **Netlify DB** — Netlify's own managed Postgres, provisioned from your
  site's dashboard if you'd rather keep everything in one place

Whichever you pick, copy the connection string — it looks like:
```
postgres://user:password@host:5432/dbname?sslmode=require
```

## 2. Local setup

```bash
git clone <this repo>
cd go-url-shortener
go mod tidy        # resolves and locks dependency versions
```

`go mod tidy` needs normal internet access (it fetches from proxy.golang.org),
so run it on your own machine, not in a restricted sandbox.

## 3. Environment variables

Set these in Netlify: **Site settings → Environment variables**

| Variable          | Value                                      |
|--------------------|---------------------------------------------|
| `DATABASE_URL`     | your Postgres connection string from step 1 |
| `PUBLIC_BASE_URL`  | e.g. `https://links.yourdomain.com` (optional — falls back to the request's Host header if unset) |

## 4. Deploy

Push to GitHub and connect the repo in Netlify (Add new site → Import an
existing project), or deploy directly with the CLI:

```bash
npm install -g netlify-cli
netlify deploy --prod
```

Netlify will run `build.sh`, which compiles both functions with
`GOOS=linux GOARCH=amd64 CGO_ENABLED=0`.

## 5. Point a subdomain at it

Same flow as your portfolio: in the project's Netlify dashboard go to
**Domain management → Add domain alias**, enter something like
`links.yourdomain.com`, then add the CNAME record it gives you at your DNS
provider (or in Netlify's own DNS panel if it manages your domain).

## 6. Test it

```bash
curl -X POST https://links.yourdomain.com/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com/sushilchlgn/multisync"}'

# {"code":"aZ3kT9","short_url":"https://links.yourdomain.com/r/aZ3kT9","original_url":"..."}

curl -i https://links.yourdomain.com/r/aZ3kT9
# HTTP/2 302
# location: https://github.com/sushilchlgn/multisync
```

## Notes / possible extensions

- Codes are 6 chars from a 62-char alphabet (~5.6×10^10 combinations),
  generated with `crypto/rand`, with a retry loop on collision.
- No auth on `/api/shorten` yet — worth adding a bearer token check (you've
  already got that pattern from the OAuth2 microservice) before this is
  public-facing for real.
- `clicks` is tracked per code but not yet exposed via an endpoint — a
  `GET /api/stats/<code>` function would be a natural next addition.
