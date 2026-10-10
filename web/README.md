# Web

Vite+ workspace managed with pnpm. Requires Node.js 22.18+ (Node.js 24 recommended)
and Go as described in the repository README.

## Commands

Run from `web/` using the project-local Vite+ CLI:

```sh
pnpm install
cp .env.example .env
pnpm exec vp run generate
pnpm exec vp check
pnpm exec vp test run
pnpm exec vp run build
```

With a current global Vite+ CLI, use `vp install` and omit `pnpm exec` above.
`pnpm exec vp run ready` runs generation, checks, tests, and all builds.
Applications use `vp build`; the shared package uses `vp pack`.

## Applications

Four React apps use Astryx and StyleX, with dependency versions centralized in
the pnpm catalog in `pnpm-workspace.yaml`. Identity has a cookie-backed login
and registration forms; the other apps currently render just an `h1` with their name.

| App        | Directory         | Development URL       |
| ---------- | ----------------- | --------------------- |
| Customer   | `apps/customer`   | http://localhost:5173 |
| Admin      | `apps/admin`      | http://localhost:5174 |
| Management | `apps/management` | http://localhost:5175 |
| Identity   | `apps/identity`   | http://localhost:5176 |

Start all apps, or just one, from `web/`:

```sh
pnpm exec vp run dev
pnpm exec vp run customer#dev
```

Replace `customer` with another app name to target it. `vp run customer#build`
builds just that app to `apps/customer/dist`; `vp run customer#preview` serves
the production build. `vp run shared#dev` watches the shared package.

`vite.app.config.ts` shares the React and Astryx StyleX compiler configuration
across the apps, including CSS cascade layers and separate development ports.
All apps load environment files from `web/`. Set `VITE_API_BASE_URL` in `web/.env`
(for example `http://localhost:8080`). Requests go directly to that URL; there is
no development proxy. Restart Vite after changing it. Production builds embed
the value, so set it before building; it is public, not a secret.

### Identity login

Start the Go API with `SESSION_SIGNING_KEY` and
`CORS_ALLOWED_ORIGINS=http://localhost:5176` configured (see the repository
README), then run `pnpm exec vp run identity#dev` and open
`http://localhost:5176/login`. Follow “Create an account” to `/register` for a
new account. Registration checks password confirmation, creates the account,
and returns to login with a success message. `/` redirects to `/login`.

Identity uses TanStack Router's file-based routes in `apps/identity/src/routes`
for `/login`, `/register`, and the protected `/session` page. Its Vite plugin generates
`src/routeTree.gen.ts` during dev/build; do not edit that file by hand.
It uses TanStack Query for login, registration, and logout mutations and the current-user query.
Successful login shows the account and session expiry; sign out revokes the
session and returns to the form. Passwords and session tokens are not persisted
in browser storage. Production hosting must serve `index.html` for frontend
routes. Set the API's `CORS_ALLOWED_ORIGINS` to the exact frontend origins
(comma-separated; no trailing slash or wildcard).

### Shared theme

`packages/shared` combines the theme and generated API client in the private
`shared` package, consumed directly from source by each app. Import the theme
from `shared/theme`. Change `src/theme.tsx` to configure the Astryx theme or color
mode (currently Neutral with system light/dark mode). `src/styles.css` loads the
reset, prebuilt theme CSS, and fonts once per app. The Astryx Vite plugin compiles
component styles into the `astryx-base` layer.

Every app wraps its content in `AppTheme` and imports `shared/styles.css`.
Headings use Astryx's `Heading` component with `level={1}` and a StyleX override.
When changing the theme package, update both its built module import and its
theme/font stylesheet imports together.

## API client

`shared/api` exports the framework-independent API client. Orval generates its
native-fetch functions and TypeScript models from `../dist/openapi.yaml`. This
entry point does not import React or the theme.

`vp run generate` first runs the Go OpenAPI stitcher, then Orval, then Vite+'s
formatter. Generated source is checked in under `packages/shared/src/generated/`;
do not edit it by hand. Rerun generation after changing contracts under `api/`.
`vp run build` also regenerates before packaging JavaScript and declarations.

Orval's [runtime base URL](https://orval.dev/docs/guides/set-base-url#runtime-base-url)
is configured as `output.baseUrl.runtime` in `packages/shared/orval.config.ts`.
It emits calls to `getApiBaseUrl()` from `src/config.ts`, which reads
`import.meta.env.VITE_API_BASE_URL` and normalizes trailing slashes. Environment
changes do not require regenerating the client. Orval's `override.requestOptions`
sets `credentials: "include"` for every request so the browser accepts and sends
the HttpOnly session cookie across origins.

Workspace consumers should add `"shared": "workspace:*"` to their dependencies.
Source exports work without a separate package build:

```ts
import { login, listEntries } from "shared/api";

const session = await login({
  email: "person@example.com",
  password: "correct horse battery staple",
});

if (session.status === 200) {
  const entries = await listEntries();
  if (entries.status === 200) {
    console.log(entries.data);
  }
}
```

Requests use absolute URLs prefixed by `VITE_API_BASE_URL` (an HTTP(S) URL,
optionally with a base path, without credentials, query strings, or fragments).
Functions accept `RequestInit` for headers, cancellation signals, and other fetch
options. They return `{ data, status, headers }`; check `status` to distinguish
successful responses from documented API errors. Network failures reject.
Authentication uses the HttpOnly `session` cookie set by login; no Authorization
header is needed. The API must allow the frontend origin and credentials through
CORS. The existing SameSite=Lax cookie supports same-site origins (such as
localhost on different ports, or HTTPS subdomains of the same domain), not
unrelated cross-site domains. Use matching hostnames locally; do not mix
`localhost` with `127.0.0.1`.

Formatting, linting, type checks, tests, packaging, and task orchestration are all
provided by Vite+. No separate ESLint, Prettier, Vitest, or build CLI is configured.
