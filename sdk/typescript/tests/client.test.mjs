import assert from "node:assert/strict";
import test from "node:test";
import { ApiError, CloakOnBoardClient } from "../dist/index.js";

test("injects token and request ID", async () => {
  let request;
  const client = new CloakOnBoardClient({
    baseUrl: "https://api.example.test/api/v1/",
    token: async () => "token-123",
    fetch: async (url, init) => {
      request = { url, init };
      return new Response(JSON.stringify({ data: [] }), { status: 200, headers: { "Content-Type": "application/json" } });
    },
  });
  await client.listApplications({ requestId: "request-42" });
  assert.equal(request.url, "https://api.example.test/api/v1/applications");
  assert.equal(request.init.headers.Authorization, "Bearer token-123");
  assert.equal(request.init.headers["X-Request-ID"], "request-42");
});

test("returns typed errors and never retries mutations", async () => {
  let calls = 0;
  const client = new CloakOnBoardClient({
    baseUrl: "https://api.example.test",
    retries: 3,
    fetch: async () => {
      calls += 1;
      return new Response(JSON.stringify({ error: "conflict", request_id: "r1" }), { status: 409 });
    },
  });
  await assert.rejects(() => client.createApplication({ name: "demo" }), (error) => {
    assert.ok(error instanceof ApiError);
    assert.equal(error.status, 409);
    assert.equal(error.requestId, "r1");
    return true;
  });
  assert.equal(calls, 1);
});

test("passes AbortSignal to fetch", async () => {
  const controller = new AbortController();
  const client = new CloakOnBoardClient({
    baseUrl: "https://api.example.test",
    fetch: async (_url, init) => {
      assert.equal(init.signal, controller.signal);
      return new Response(JSON.stringify({ data: [] }), { status: 200 });
    },
  });
  await client.listApplications({ signal: controller.signal });
});

test("retries safe reads and serializes pagination", async () => {
  let calls = 0;
  let lastUrl = "";
  const client = new CloakOnBoardClient({
    baseUrl: "https://api.example.test",
    retries: 1,
    fetch: async (url) => {
      calls += 1;
      lastUrl = String(url);
      if (calls === 1) return new Response("unavailable", { status: 503 });
      return new Response(JSON.stringify({ data: [] }), { status: 200 });
    },
  });
  await client.listAuditLogs({ page: 2, pageSize: 25 });
  assert.equal(calls, 2);
  assert.equal(lastUrl, "https://api.example.test/audit-logs?page=2&page_size=25");
});
