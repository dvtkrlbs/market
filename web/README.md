# Web

Vite+ workspace managed with pnpm. Requires Node.js 22.18+ (Node.js 24 recommended)
and Go as described in the repository README.

## Commands

Run from `web/` using the project-local Vite+ CLI:

```sh
pnpm install
pnpm exec vp run generate
pnpm exec vp check
pnpm exec vp test run
pnpm exec vp run build
```

With a current global Vite+ CLI, use `vp install` and omit `pnpm exec` above.
`pnpm exec vp run ready` runs generation, checks, tests, and all builds.
Applications use `vp build`; the API client uses `vp pack`.

## Applications

Four React apps use Astryx and StyleX, with dependency versions centralized in
the pnpm catalog in `pnpm-workspace.yaml`. Each currently renders just an `h1`
with its name.

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
the production build. `vp run client#dev` still watches the API client.

`vite.app.config.ts` shares the React and Astryx StyleX compiler configuration
across the apps, including CSS cascade layers and separate development ports.

### Shared theme

`packages/theme` is the private `theme` package, consumed directly from source
by each app. Change `src/index.tsx` to configure the shared Astryx theme or color
mode (currently Neutral with system light/dark mode). `src/styles.css` loads the
reset, prebuilt theme CSS, and fonts once per app. The Astryx Vite plugin compiles
component styles into the `astryx-base` layer.

Every app wraps its content in `AppTheme` and imports `theme/styles.css`.
Headings use Astryx's `Heading` component with `level={1}` and a StyleX override.
When changing the theme package, update both its built module import and its
theme/font stylesheet imports together.

## API client

`packages/client` is the private, framework-independent `client` package. Orval
generates its native-fetch functions and TypeScript models from
`../dist/openapi.yaml`. It has no runtime dependencies.

`vp run generate` first runs the Go OpenAPI stitcher, then Orval, then Vite+'s
formatter. Generated source is checked in under `packages/client/src/generated/`;
do not edit it by hand. Rerun generation after changing contracts under `api/`.
`vp run build` also regenerates before packaging JavaScript and declarations.

Future workspace consumers should add `"client": "workspace:*"` to their
dependencies and build the package before importing it:

```ts
import { login, listEntries } from "client";

const session = await login({
  email: "person@example.com",
  password: "correct horse battery staple",
});

if (session.status === 200) {
  const entries = await listEntries({
    headers: { Authorization: `Bearer ${session.data.token}` },
  });
  if (entries.status === 200) {
    console.log(entries.data);
  }
}
```

Requests use relative paths, so serve the API on the same origin as the browser
application (or proxy its routes to `http://localhost:8080` in development).
Functions accept `RequestInit` for headers, cancellation signals, and other fetch
options. They return `{ data, status, headers }`; check `status` to distinguish
successful responses from documented API errors. Network failures reject.
Authentication is explicit: pass the Bearer token on protected requests.

Formatting, linting, type checks, tests, packaging, and task orchestration are all
provided by Vite+. No separate ESLint, Prettier, Vitest, or build CLI is configured.
