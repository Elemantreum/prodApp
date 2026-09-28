export interface Training {
  id: number;
  description: string;
  date: string;
  maxQuantity: number;
  duration: number;
  isActive: boolean;
  createdAt: string;
  bookedCount: number;
  recurringTrainingId?: number | null;
}

export interface RecurringTraining {
  id: number;
  description: string;
  anchorDate: string;
  maxQuantity: number;
  duration: number;
  leadTime: number;
  isActive: boolean;
  createdAt: string;
}

export type BookingStatus = "active" | "cancelled" | "completed" | "not_completed";

export interface Booking {
  id: number;
  userId: number;
  trainingId: number;
  status: BookingStatus;
  bookingDate: string;
  unbookingDate: string | null;
}

export interface Participant {
  userId: number;
  firstName: string;
  lastName: string;
  bookingDate: string;
  identities: Identity[];
}

export interface BookingDetail {
  bookingId: number;
  userId: number;
  firstName: string;
  lastName: string;
  status: BookingStatus;
  bookingDate: string;
  identities: Identity[];
}

export interface SinnerStats {
  userId: number;
  firstName: string;
  lastName: string;
  lateCancellations: number;
  noShows: number;
  penaltyPoints: number;
}

export interface CreateTrainingInput {
  description: string;
  date: string;
  maxQuantity: number;
  duration: number;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
}

export interface LinkCode {
  code: string;
  expiresAt: string;
}

export interface Identity {
  provider: string;
  externalId: string;
  username: string;
}

export interface User {
  id: number;
  firstName: string;
  lastName: string;
  isApproved: boolean;
  isAdmin: boolean;
  isRegistered: boolean;
  createdAt: string;
  updatedAt: string;
  identities: Identity[];
}
