# @cloak-on-board/sdk

Typed ESM client for cloak-on-board API v0.4.

```ts
import { CloakOnBoardClient } from "@cloak-on-board/sdk";

const client = new CloakOnBoardClient({
  baseUrl: "https://onboarder.example.com/api/v1",
  token: async () => session.accessToken,
});

const applications = await client.listApplications();
```

The token provider is evaluated for every call. Pass an `AbortSignal` or
`requestId` in the final options argument. Only safe GET requests are retried;
mutations, especially secret rotation, are never retried.

Administrative realm credentials are write-only and never represented in
response types.
