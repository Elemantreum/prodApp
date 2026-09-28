import type {
  AuthResponse,
  Booking,
  BookingDetail,
  CreateTrainingInput,
  Identity,
  LinkCode,
  Participant,
  RecurringTraining,
  SinnerStats,
  Training,
  User,
} from "./types";

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

export function setAuthToken(token: string | null): void {
  authToken = token;
}

let unauthorizedHandler: (() => Promise<string | null>) | null = null;

export function setUnauthorizedHandler(handler: (() => Promise<string | null>) | null): void {
  unauthorizedHandler = handler;
}

async function parseErrorMessage(res: Response): Promise<string> {
  try {
    const body = (await res.clone().json()) as { error?: string };
    if (body?.error) return body.error;
  } catch {}
  return res.statusText || `HTTP ${res.status}`;
}

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
  authTelegram(initData: string): Promise<AuthResponse> {
    return request<AuthResponse>(
      "/auth/telegram",
      { method: "POST", headers: { Authorization: `tma ${initData}` } },
      { skipAuthRetry: true },
    );
  },

  authVk(rawQuery: string): Promise<AuthResponse> {
    return request<AuthResponse>(`/auth/vk?${rawQuery}`, { method: "POST" }, { skipAuthRetry: true });
  },

  generateLinkCode(): Promise<LinkCode> {
    return request<LinkCode>("/user/linkCode", { method: "POST" });
  },

  linkTelegram(initData: string, code: string): Promise<AuthResponse> {
    return request<AuthResponse>(
      "/auth/telegram/link",
      { method: "POST", headers: { Authorization: `tma ${initData}` }, body: JSON.stringify({ code }) },
      { skipAuthRetry: true },
    );
  },

  linkVk(rawQuery: string, code: string): Promise<AuthResponse> {
    const query = rawQuery ? `${rawQuery}&code=${encodeURIComponent(code)}` : `code=${encodeURIComponent(code)}`;
    return request<AuthResponse>(`/auth/vk/link?${query}`, { method: "POST" }, { skipAuthRetry: true });
  },

  refreshToken(refreshToken: string): Promise<AuthResponse> {
    return request<AuthResponse>(
      "/auth/refresh",
      { method: "POST", body: JSON.stringify({ refresh_token: refreshToken }) },
      { skipAuthRetry: true },
    );
  },

  logout(refreshToken: string): Promise<void> {
    return request<void>(
      "/auth/logout",
      { method: "POST", body: JSON.stringify({ refresh_token: refreshToken }) },
      { skipAuthRetry: true },
    );
  },

  getTrainings(): Promise<Training[]> {
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

  updateTraining(id: number, input: CreateTrainingInput & { isActive: boolean }): Promise<Training> {
    return request<Training>("/trainings/update", {
      method: "POST",
      body: JSON.stringify({ id, ...input }),
    });
  },

  deleteTraining(id: number): Promise<true> {
    return request<true>("/trainings/delete", {
      method: "POST",
      body: JSON.stringify({ id }),
    });
  },

  createRecurringTraining(input: {
    description: string;
    anchorDate: string;
    maxQuantity: number;
    duration: number;
    leadTime: number;
  }): Promise<RecurringTraining> {
    return request<RecurringTraining>("/trainings/recurring/create", {
      method: "POST",
      body: JSON.stringify(input),
    });
  },

  getRecurringTrainings(): Promise<RecurringTraining[]> {
    return request<RecurringTraining[] | null>("/trainings/recurring/get").then((r) => r ?? []);
  },

  updateRecurringTraining(
    id: number,
    input: { description: string; anchorDate: string; maxQuantity: number; duration: number; leadTime: number; isActive: boolean },
  ): Promise<RecurringTraining> {
    return request<RecurringTraining>("/trainings/recurring/update", {
      method: "POST",
      body: JSON.stringify({ id, ...input }),
    });
  },

  getBookingsByUserId(userId: number): Promise<Booking[]> {
    return request<Booking[] | null>(`/booking/getByUserId?userId=${userId}`).then((r) => r ?? []);
  },

  getTrainingParticipants(trainingId: number): Promise<Participant[]> {
    return request<Participant[] | null>(`/booking/getByTrainingId?trainingId=${trainingId}`).then((r) => r ?? []);
  },

  getAntiLeaderboard(): Promise<SinnerStats[]> {
    return request<SinnerStats[] | null>("/booking/antiLeaderboard").then((r) => r ?? []);
  },

  getTrainingBookings(trainingId: number): Promise<BookingDetail[]> {
    return request<BookingDetail[] | null>(`/booking/admin/getByTrainingId?trainingId=${trainingId}`).then(
      (r) => r ?? [],
    );
  },

  setAttendance(bookingId: number, attended: boolean): Promise<Booking> {
    return request<Booking>("/booking/setAttendance", {
      method: "POST",
      body: JSON.stringify({ bookingId, attended }),
    });
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

  getOwnProfile(): Promise<User> {
    return request<User>("/user/me");
  },

  getOwnIdentities(): Promise<Identity[]> {
    return request<Identity[] | null>("/user/identities").then((r) => r ?? []);
  },

  updateProfile(firstName: string, lastName: string, userId?: number): Promise<User> {
    return request<User>("/user/updateProfile", {
      method: "POST",
      body: JSON.stringify({ userId, firstName, lastName }),
    });
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
