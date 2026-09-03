import type { AuthResponse, Booking, CreateTrainingInput, LinkCode, Participant, Training, User } from "./types";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

export class ApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

let authToken: string | null = null;

/** Called by AuthContext once the login flow resolves (or on logout, with null). */
export function setAuthToken(token: string | null): void {
  authToken = token;
}

/**
 * Registered by AuthContext. When a request comes back 401 (expired/invalid
 * token), the client asks this to mint a fresh one from the still-valid
 * platform launch data (Telegram initData / VK launch params) and retries
 * the request exactly once — the user never sees a "please log in again".
 */
let unauthorizedHandler: (() => Promise<string | null>) | null = null;

export function setUnauthorizedHandler(handler: (() => Promise<string | null>) | null): void {
  unauthorizedHandler = handler;
}

async function parseErrorMessage(res: Response): Promise<string> {
  try {
    const body = (await res.clone().json()) as { error?: string };
    if (body?.error) return body.error;
  } catch {
    // body wasn't JSON — fall through to statusText
  }
  return res.statusText || `HTTP ${res.status}`;
}

/**
 * `skipAuthRetry` is set for the two /auth/* calls themselves: a 401 there
 * means "the platform signature is invalid", not "the token expired", so
 * feeding it back into unauthorizedHandler (which just calls the very same
 * login function again) would recurse forever, hammering the backend with
 * an unbounded retry loop instead of surfacing a real error.
 */
async function request<T>(
  path: string,
  options: RequestInit = {},
  { retrying = false, skipAuthRetry = false }: { retrying?: boolean; skipAuthRetry?: boolean } = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  if (authToken) headers.set("Authorization", `Bearer ${authToken}`);
  if (options.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (res.status === 401 && !retrying && !skipAuthRetry && unauthorizedHandler) {
    const fresh = await unauthorizedHandler();
    if (fresh) {
      authToken = fresh;
      return request<T>(path, options, { retrying: true });
    }
  }

  if (!res.ok) {
    throw new ApiError(res.status, await parseErrorMessage(res));
  }

  if (res.status === 204) return undefined as T;

  const text = await res.text();
  return (text ? JSON.parse(text) : undefined) as T;
}

export const api = {
  /** initData is the raw, signed string Telegram gives the mini app. */
  authTelegram(initData: string): Promise<AuthResponse> {
    return request<AuthResponse>(
      "/auth/telegram",
      { method: "POST", headers: { Authorization: `tma ${initData}` } },
      { skipAuthRetry: true },
    );
  },

  /** rawQuery is the VK launch params querystring (with `sign`), unmodified. */
  authVk(rawQuery: string): Promise<AuthResponse> {
    return request<AuthResponse>(`/auth/vk?${rawQuery}`, { method: "POST" }, { skipAuthRetry: true });
  },

  /** Mints a one-time code (5 min TTL) for the *currently* logged-in account,
   *  to be redeemed via linkTelegram/linkVk from a session on the other
   *  platform — that's how the same person ends up as one account instead
   *  of two separate ones. */
  generateLinkCode(): Promise<LinkCode> {
    return request<LinkCode>("/user/linkCode", { method: "POST" });
  },

  /** Redeems a link code using this Telegram session's initData. Returns a
   *  token for the account the code was generated on — NOT a new one. */
  linkTelegram(initData: string, code: string): Promise<AuthResponse> {
    return request<AuthResponse>(
      "/auth/telegram/link",
      { method: "POST", headers: { Authorization: `tma ${initData}` }, body: JSON.stringify({ code }) },
      { skipAuthRetry: true },
    );
  },

  /** Redeems a link code using this VK session's launch params. */
  linkVk(rawQuery: string, code: string): Promise<AuthResponse> {
    const query = rawQuery ? `${rawQuery}&code=${encodeURIComponent(code)}` : `code=${encodeURIComponent(code)}`;
    return request<AuthResponse>(`/auth/vk/link?${query}`, { method: "POST" }, { skipAuthRetry: true });
  },

  getTrainings(): Promise<Training[]> {
    // Go marshals a nil slice as JSON `null`, not `[]` — always normalize to
    // an array here so callers never have to null-check before .filter/.map.
    return request<Training[] | null>("/trainings/get").then((r) => r ?? []);
  },

  getTraining(id: number): Promise<Training> {
    return request<Training>(`/trainings/get/${id}`);
  },

  createTraining(input: CreateTrainingInput): Promise<Training> {
    return request<Training>("/trainings/create", {
      method: "POST",
      body: JSON.stringify(input),
    });
  },

  getBookingsByUserId(userId: number): Promise<Booking[]> {
    return request<Booking[] | null>(`/booking/getByUserId?userId=${userId}`).then((r) => r ?? []);
  },

  getTrainingParticipants(trainingId: number): Promise<Participant[]> {
    return request<Participant[] | null>(`/booking/getByTrainingId?trainingId=${trainingId}`).then((r) => r ?? []);
  },

  signUp(trainingId: number): Promise<Booking> {
    return request<Booking>("/booking/signUp", {
      method: "POST",
      body: JSON.stringify({ trainingId }),
    });
  },

  cancelBooking(bookingId: number): Promise<true> {
    return request<true>("/booking/cancel", {
      method: "POST",
      body: JSON.stringify({ bookingId }),
    });
  },

  getUsers(): Promise<User[]> {
    return request<User[] | null>("/user/get").then((r) => r ?? []);
  },

  approveUser(userId: number): Promise<User> {
    return request<User>("/user/approve", { method: "POST", body: JSON.stringify({ userId }) });
  },

  unapproveUser(userId: number): Promise<User> {
    return request<User>("/user/unapprove", { method: "POST", body: JSON.stringify({ userId }) });
  },

  makeAdmin(userId: number): Promise<User> {
    return request<User>("/user/makeAdmin", { method: "POST", body: JSON.stringify({ userId }) });
  },

  removeAdmin(userId: number): Promise<User> {
    return request<User>("/user/removeAdmin", { method: "POST", body: JSON.stringify({ userId }) });
  },

  deleteUser(userId: number): Promise<true> {
    return request<true>("/user/delete", { method: "POST", body: JSON.stringify({ userId }) });
  },
};
