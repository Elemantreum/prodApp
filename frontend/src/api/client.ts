import type {
  SectionDirectory,
  Section,
  SectionInvite,
  SectionInviteInput,
  InvitedSection,
  HistoryItem,
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

let selectedSection: number | null = null;
export function setSelectedSection(id: number | null) {
  selectedSection = id;
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
  {
    retrying = false,
    skipAuthRetry = false,
    includeSection = true,
  }: { retrying?: boolean; skipAuthRetry?: boolean; includeSection?: boolean } = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  if (includeSection && selectedSection !== null && !headers.has("X-Section-ID")) {
    headers.set("X-Section-ID", String(selectedSection));
  }
  if (authToken && !headers.has("Authorization")) headers.set("Authorization", `Bearer ${authToken}`);
  if (options.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (res.status === 401 && !retrying && !skipAuthRetry && unauthorizedHandler) {
    const fresh = await unauthorizedHandler();
    if (fresh) {
      authToken = fresh;
      headers.set("Authorization", `Bearer ${fresh}`);
      return request<T>(path, { ...options, headers }, { retrying: true, includeSection });
    }
  }

  if (!res.ok) {
    throw new ApiError(res.status, await parseErrorMessage(res));
  }

  if (res.status === 204) return undefined as T;

  const text = await res.text();
  return (text ? JSON.parse(text) : undefined) as T;
}

function sectionHeaders(sectionId = selectedSection): HeadersInit {
  return sectionId === null ? {} : { "X-Section-ID": String(sectionId) };
}

export const api = {
  deleteSection(id: number): Promise<true> {
    return request(`/sections?id=${id}`, { method: "DELETE" });
  },
  candidates(): Promise<{ id: number; firstName: string; lastName: string; admin: boolean }[]> {
    return request("/booking/admin/candidates");
  },
  deleteRule(id: number): Promise<true> {
    return request(`/trainings/recurring/${id}`, { method: "DELETE" });
  },
  sections(): Promise<SectionDirectory> {
    return request("/sections", {}, { includeSection: false });
  },
  sectionInvites(sectionId: number): Promise<SectionInvite[]> {
    return request<SectionInvite[] | null>("/sections/invites", { headers: sectionHeaders(sectionId) }).then(
      (invites) => invites ?? [],
    );
  },
  createSectionInvite(sectionId: number, input: SectionInviteInput): Promise<SectionInvite & { code: string }> {
    return request("/sections/invites", {
      method: "POST",
      headers: sectionHeaders(sectionId),
      body: JSON.stringify(input),
    });
  },
  revokeSectionInvite(sectionId: number, id: number): Promise<void> {
    return request(`/sections/invites/${id}`, { method: "DELETE", headers: sectionHeaders(sectionId) });
  },
  previewSectionInvite(code: string): Promise<InvitedSection> {
    return request(
      "/sections/invites/preview",
      { method: "POST", body: JSON.stringify({ code }) },
      { includeSection: false },
    );
  },
  redeemSectionInvite(code: string): Promise<InvitedSection> {
    return request(
      "/sections/invites/redeem",
      { method: "POST", body: JSON.stringify({ code }) },
      { includeSection: false },
    );
  },
  saveSection(id: number, name: string, logo: string): Promise<Section> {
    return request("/sections/save", { method: "POST", body: JSON.stringify({ id, name, logo }) });
  },
  addVenue(sectionId: number, name: string): Promise<string> {
    return request("/sections/venues", {
      method: "POST",
      headers: sectionHeaders(sectionId),
      body: JSON.stringify({ name }),
    });
  },
  deleteVenue(sectionId: number, name: string): Promise<void> {
    return request(`/sections/venues?${new URLSearchParams({ name })}`, {
      method: "DELETE",
      headers: sectionHeaders(sectionId),
    });
  },
  setMember(sectionId: number, userId: number, admin: boolean, remove = false): Promise<true> {
    return request("/sections/member", {
      method: "POST",
      body: JSON.stringify({ sectionId, userId, admin, remove }),
    });
  },
  websiteBot(): Promise<{ bot: string }> {
    return request("/auth/telegram/website", {}, { skipAuthRetry: true });
  },

  websiteLogin(fields: Record<string, string>): Promise<AuthResponse> {
    return request(
      "/auth/telegram/website",
      {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams(fields),
      },
      { skipAuthRetry: true },
    );
  },

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
    const query = rawQuery
      ? `${rawQuery}&code=${encodeURIComponent(code)}`
      : `code=${encodeURIComponent(code)}`;
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

  getTrainings(sectionId?: number): Promise<Training[]> {
    return request<Training[] | null>("/trainings/get", { headers: sectionHeaders(sectionId) }).then(
      (r) => r ?? [],
    );
  },

  getTraining(id: number, sectionId?: number): Promise<Training> {
    return request<Training>(`/trainings/get/${id}`, { headers: sectionHeaders(sectionId) });
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
    venue?: string;
    description: string;
    details?: string;
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
    input: {
      venue?: string;
      description: string;
      details?: string;
      anchorDate: string;
      maxQuantity: number;
      duration: number;
      leadTime: number;
      isActive: boolean;
    },
  ): Promise<RecurringTraining> {
    return request<RecurringTraining>("/trainings/recurring/update", {
      method: "POST",
      body: JSON.stringify({ id, ...input }),
    });
  },

  getBookingsByUserId(userId: number, sectionId?: number): Promise<Booking[]> {
    return request<Booking[] | null>(`/booking/getByUserId?userId=${userId}`, {
      headers: sectionHeaders(sectionId),
    }).then((r) => r ?? []);
  },

  getTrainingParticipants(trainingId: number): Promise<Participant[]> {
    return request<Participant[] | null>(`/booking/getByTrainingId?trainingId=${trainingId}`).then(
      (r) => r ?? [],
    );
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

  history(userId?: number): Promise<HistoryItem[]> {
    return request(userId ? `/user/history?userId=${userId}` : "/booking/admin/history");
  },
  sectionMembers(
    sectionId?: number,
  ): Promise<{ id: number; firstName: string; lastName: string; admin: boolean }[]> {
    return request("/booking/admin/members", { headers: sectionHeaders(sectionId) });
  },
  waitingMembers(id: number): Promise<{ id: number; firstName: string; lastName: string }[]> {
    return request(`/booking/admin/waitlist?trainingId=${id}`);
  },
  manageBooking(change: {
    action: string;
    trainingId: number;
    previousUserId?: number;
    userId?: number;
    bookingId?: number;
    targetId?: number;
  }): Promise<true> {
    return request("/booking/admin/manage", { method: "POST", body: JSON.stringify(change) });
  },
  getWaitlist(sectionId?: number): Promise<number[]> {
    return request<number[] | null>("/booking/waitlist", { headers: sectionHeaders(sectionId) }).then(
      (ids) => ids ?? [],
    );
  },

  setWaiting(trainingId: number, join: boolean, sectionId?: number): Promise<true> {
    return request<true>(`/booking/waitlist?trainingId=${trainingId}`, {
      method: join ? "PUT" : "DELETE",
      headers: sectionHeaders(sectionId),
    });
  },

  signUp(trainingId: number): Promise<Booking> {
    return request<Booking>("/booking/signUp", {
      method: "POST",
      body: JSON.stringify({ trainingId }),
    });
  },

  cancelBooking(bookingId: number, sectionId?: number): Promise<true> {
    return request<true>("/booking/cancel", {
      headers: sectionHeaders(sectionId),
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

  deleteUser(userId: number): Promise<true> {
    return request<true>("/user/delete", { method: "POST", body: JSON.stringify({ userId }) });
  },
};
