// Mirrors backend/internal/{training,booking}/model.go exactly — keep in
// sync by hand, there's no shared codegen between the two sides.

export interface Training {
  id: number;
  description: string;
  /** ISO 8601 timestamp */
  date: string;
  maxQuantity: number;
  /** Go time.Duration, marshalled as nanoseconds */
  duration: number;
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
  username: string;
  bookingDate: string;
}

export interface CreateTrainingInput {
  description: string;
  /** ISO 8601 timestamp */
  date: string;
  maxQuantity: number;
  /** nanoseconds */
  duration: number;
}

export interface AuthResponse {
  access_token: string;
}

export interface LinkCode {
  code: string;
  /** ISO 8601 timestamp */
  expiresAt: string;
}

export interface User {
  id: number;
  username: string;
  firstName: string;
  lastName: string;
  isApproved: boolean;
  isAdmin: boolean;
  createdAt: string;
  updatedAt: string;
}
