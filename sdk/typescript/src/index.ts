export type TokenProvider = () => string | undefined | Promise<string | undefined>;

export interface ClientOptions {
  baseUrl: string;
  token?: TokenProvider;
  fetch?: typeof globalThis.fetch;
  retries?: number;
}

export interface RequestOptions {
  signal?: AbortSignal;
  requestId?: string;
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
    public readonly requestId?: string,
    public readonly body?: unknown,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export interface Page<T> {
  data: T[];
  nextCursor?: string;
}

export interface Application {
  id: string;
  name: string;
  slug: string;
  status: string;
  [key: string]: unknown;
}

export interface Environment {
  id: string;
  name: string;
  slug: string;
  promotion_order: number;
  protected: boolean;
  enabled: boolean;
}

export interface RealmConnection {
  id: string;
  environment_id: string;
  name: string;
  base_url: string;
  realm: string;
  enabled: boolean;
  secret_set: boolean;
}

export interface Deployment {
  id: string;
  application_id: string;
  environment_id: string;
  snapshot_id: string;
  status: string;
  drift_status: string;
}

type Json = Record<string, unknown>;

export class CloakOnBoardClient {
  private readonly fetcher: typeof globalThis.fetch;
  private readonly baseUrl: string;
  private readonly retries: number;

  constructor(private readonly options: ClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.fetcher = options.fetch ?? globalThis.fetch;
    this.retries = Math.max(0, options.retries ?? 1);
  }

  currentUser(options?: RequestOptions) { return this.request<unknown>("GET", "/auth/me", undefined, options); }
  listApplications(options?: RequestOptions) { return this.request<Application[]>("GET", "/applications", undefined, options); }
  getApplication(id: string, options?: RequestOptions) { return this.request<Application>("GET", `/applications/${encodeURIComponent(id)}`, undefined, options); }
  createApplication(input: Json, options?: RequestOptions) { return this.request<Application>("POST", "/applications", input, options); }
  updateApplication(id: string, input: Json, options?: RequestOptions) { return this.request<Application>("PUT", `/applications/${encodeURIComponent(id)}`, input, options); }
  deleteApplication(id: string, options?: RequestOptions) { return this.request<void>("DELETE", `/applications/${encodeURIComponent(id)}`, undefined, options); }
  provisionApplication(id: string, options?: RequestOptions) { return this.request<unknown>("POST", `/applications/${encodeURIComponent(id)}/provision`, {}, options); }
  importApplication(input: Json, options?: RequestOptions) { return this.request<Application>("POST", "/applications/import", input, options); }
  listKeycloakClients(search = "", options?: RequestOptions) {
    return this.request<unknown[]>("GET", `/keycloak/clients${search ? `?search=${encodeURIComponent(search)}` : ""}`, undefined, options);
  }
  listClientScopes(id: string, options?: RequestOptions) { return this.request<unknown>("GET", `/applications/${encodeURIComponent(id)}/client-scopes`, undefined, options); }
  assignClientScope(id: string, scopeId: string, type: "default" | "optional", options?: RequestOptions) {
    return this.request<unknown>("PUT", `/applications/${encodeURIComponent(id)}/client-scopes/${encodeURIComponent(scopeId)}`, { type }, options);
  }
  removeClientScope(id: string, scopeId: string, type: "default" | "optional", options?: RequestOptions) {
    return this.request<void>("DELETE", `/applications/${encodeURIComponent(id)}/client-scopes/${encodeURIComponent(scopeId)}?type=${type}`, undefined, options);
  }
  listProtocolMappers(id: string, options?: RequestOptions) { return this.request<unknown[]>("GET", `/applications/${encodeURIComponent(id)}/protocol-mappers`, undefined, options); }
  createProtocolMapper(id: string, input: Json, options?: RequestOptions) { return this.request<unknown>("POST", `/applications/${encodeURIComponent(id)}/protocol-mappers`, input, options); }
  updateProtocolMapper(id: string, mapperId: string, input: Json, options?: RequestOptions) {
    return this.request<unknown>("PUT", `/applications/${encodeURIComponent(id)}/protocol-mappers/${encodeURIComponent(mapperId)}`, input, options);
  }
  deleteProtocolMapper(id: string, mapperId: string, options?: RequestOptions) {
    return this.request<void>("DELETE", `/applications/${encodeURIComponent(id)}/protocol-mappers/${encodeURIComponent(mapperId)}`, undefined, options);
  }

  listTemplates(options?: RequestOptions) { return this.request<unknown[]>("GET", "/templates", undefined, options); }
  getTemplate(id: string, options?: RequestOptions) { return this.request<unknown>("GET", `/templates/${encodeURIComponent(id)}`, undefined, options); }
  seedTemplates(options?: RequestOptions) { return this.request<unknown>("POST", "/templates/seed", {}, options); }
  getSettings(options?: RequestOptions) { return this.request<unknown>("GET", "/settings", undefined, options); }
  updateSettings(input: Json, options?: RequestOptions) { return this.request<unknown>("PUT", "/settings", input, options); }

  listEnvironments(options?: RequestOptions) { return this.request<Environment[]>("GET", "/environments", undefined, options); }
  createEnvironment(input: Json, options?: RequestOptions) { return this.request<Environment>("POST", "/environments", input, options); }
  updateEnvironment(id: string, input: Json, options?: RequestOptions) { return this.request<Environment>("PUT", `/environments/${encodeURIComponent(id)}`, input, options); }
  deleteEnvironment(id: string, options?: RequestOptions) { return this.request<void>("DELETE", `/environments/${encodeURIComponent(id)}`, undefined, options); }
  listRealmConnections(options?: RequestOptions) { return this.request<RealmConnection[]>("GET", "/realm-connections", undefined, options); }
  createRealmConnection(input: Json, options?: RequestOptions) { return this.request<RealmConnection>("POST", "/realm-connections", input, options); }
  updateRealmConnection(id: string, input: Json, options?: RequestOptions) { return this.request<RealmConnection>("PUT", `/realm-connections/${encodeURIComponent(id)}`, input, options); }
  testRealmConnection(id: string, options?: RequestOptions) { return this.request<RealmConnection>("POST", `/realm-connections/${encodeURIComponent(id)}/test`, {}, options); }
  disableRealmConnection(id: string, options?: RequestOptions) { return this.request<void>("DELETE", `/realm-connections/${encodeURIComponent(id)}`, undefined, options); }

  listDeployments(applicationId?: string, options?: RequestOptions) {
    const query = applicationId ? `?application_id=${encodeURIComponent(applicationId)}` : "";
    return this.request<Deployment[]>("GET", `/deployments${query}`, undefined, options);
  }
  listSnapshots(applicationId: string, options?: RequestOptions) { return this.request<unknown[]>("GET", `/applications/${encodeURIComponent(applicationId)}/snapshots`, undefined, options); }
  createSnapshot(applicationId: string, options?: RequestOptions) { return this.request<unknown>("POST", `/applications/${encodeURIComponent(applicationId)}/snapshots`, {}, options); }
  deployApplication(applicationId: string, input: Json, options?: RequestOptions) { return this.request<Deployment>("POST", `/applications/${encodeURIComponent(applicationId)}/deployments`, input, options); }
  promoteApplication(applicationId: string, input: Json, options?: RequestOptions) { return this.request<Deployment>("POST", `/applications/${encodeURIComponent(applicationId)}/promotions`, input, options); }
  rollbackDeployment(id: string, options?: RequestOptions) { return this.request<Deployment>("POST", `/deployments/${encodeURIComponent(id)}/rollback`, {}, options); }
  checkDrift(deploymentId: string, options?: RequestOptions) { return this.request<unknown>("POST", `/deployments/${encodeURIComponent(deploymentId)}/drift-checks`, {}, options); }
  listDriftRuns(deploymentId?: string, options?: RequestOptions) {
    const query = deploymentId ? `?deployment_id=${encodeURIComponent(deploymentId)}` : "";
    return this.request<unknown[]>("GET", `/drift-runs${query}`, undefined, options);
  }
  requestDriftReconciliation(id: string, options?: RequestOptions) { return this.request<unknown>("POST", `/deployments/${encodeURIComponent(id)}/reconcile`, {}, options); }
  requestSecretRotation(id: string, options?: RequestOptions) { return this.request<unknown>("POST", `/deployments/${encodeURIComponent(id)}/rotate-secret`, {}, options); }
  consumeSecretDelivery(id: string, options?: RequestOptions) { return this.request<{ secret: string }>("POST", `/secret-deliveries/${encodeURIComponent(id)}/consume`, {}, options); }

  submitApproval(applicationId: string, input: Json, options?: RequestOptions) { return this.request<unknown>("POST", `/applications/${encodeURIComponent(applicationId)}/approval-requests`, input, options); }
  listApprovals(options?: RequestOptions) { return this.request<unknown[]>("GET", "/approval-requests", undefined, options); }
  getApproval(id: string, options?: RequestOptions) { return this.request<unknown>("GET", `/approval-requests/${encodeURIComponent(id)}`, undefined, options); }
  approve(id: string, comment = "", options?: RequestOptions) { return this.request<unknown>("POST", `/approval-requests/${encodeURIComponent(id)}/approve`, { comment }, options); }
  reject(id: string, comment: string, options?: RequestOptions) { return this.request<unknown>("POST", `/approval-requests/${encodeURIComponent(id)}/reject`, { comment }, options); }
  cancelApproval(id: string, comment = "", options?: RequestOptions) { return this.request<unknown>("POST", `/approval-requests/${encodeURIComponent(id)}/cancel`, { comment }, options); }
  retryApproval(id: string, options?: RequestOptions) { return this.request<unknown>("POST", `/approval-requests/${encodeURIComponent(id)}/retry`, {}, options); }

  listJobs(options?: RequestOptions) { return this.request<unknown[]>("GET", "/jobs", undefined, options); }
  getJob(id: string, options?: RequestOptions) { return this.request<unknown>("GET", `/jobs/${encodeURIComponent(id)}`, undefined, options); }
  listNotifications(options?: RequestOptions) { return this.request<unknown[]>("GET", "/notifications", undefined, options); }
  unreadNotificationCount(options?: RequestOptions) { return this.request<{ count: number }>("GET", "/notifications/unread-count", undefined, options); }
  markNotificationRead(id: string, options?: RequestOptions) { return this.request<unknown>("PUT", `/notifications/${encodeURIComponent(id)}/read`, {}, options); }
  markAllNotificationsRead(options?: RequestOptions) { return this.request<unknown>("PUT", "/notifications/read-all", {}, options); }
  listAuditLogs(pagination: { page?: number; pageSize?: number } = {}, options?: RequestOptions) {
    const query = new URLSearchParams();
    if (pagination.page) query.set("page", String(pagination.page));
    if (pagination.pageSize) query.set("page_size", String(pagination.pageSize));
    return this.request<unknown[]>("GET", `/audit-logs${query.size ? `?${query}` : ""}`, undefined, options);
  }
  getAuditLog(id: string, options?: RequestOptions) { return this.request<unknown>("GET", `/audit-logs/${encodeURIComponent(id)}`, undefined, options); }

  private async request<T>(
    method: string,
    path: string,
    body?: Json,
    options: RequestOptions = {},
  ): Promise<T> {
    const token = await this.options.token?.();
    const headers: Record<string, string> = { Accept: "application/json" };
    if (body) headers["Content-Type"] = "application/json";
    if (token) headers.Authorization = `Bearer ${token}`;
    if (options.requestId) headers["X-Request-ID"] = options.requestId;

    let response: Response | undefined;
    for (let attempt = 0; attempt <= (method === "GET" ? this.retries : 0); attempt += 1) {
      try {
        response = await this.fetcher(this.baseUrl + path, {
          method,
          headers,
          body: body ? JSON.stringify(body) : undefined,
          signal: options.signal,
        });
        if (response.status < 500 || attempt === this.retries) break;
      } catch (error) {
        if (options.signal?.aborted || attempt === this.retries || method !== "GET") throw error;
      }
    }
    if (!response) throw new Error("Request failed before receiving a response");
    if (response.status === 204) return undefined as T;
    const payload = await response.json().catch(() => undefined) as { data?: T; error?: string; request_id?: string } | undefined;
    if (!response.ok) {
      throw new ApiError(response.status, payload?.error ?? "API request failed", payload?.request_id ?? response.headers.get("X-Request-ID") ?? undefined, payload);
    }
    return payload?.data as T;
  }
}
