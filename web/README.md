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
`pnpm exec vp run ready` runs generation, checks, tests, and the library build.
Builds use Vite+'s `vp pack`; there is no web application yet. `vp run dev`
regenerates the client and watches its TypeScript source for library builds.
Future Vite applications can use `vp dev` for their development server.

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
