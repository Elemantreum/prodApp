export interface Training {
  venue: string;
  id: number;
  description: string;
  details?: string;
  date: string;
  maxQuantity: number;
  duration: number;
  isActive: boolean;
  createdAt: string;
  bookedCount: number;
  recurringTrainingId?: number | null;
}

export interface RecurringTraining {
  venue: string;
  id: number;
  description: string;
  details?: string;
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
  venue?: string;
  description: string;
  details?: string;
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
  profileLocked: boolean;
  isAdmin: boolean;
  isRegistered: boolean;
  createdAt: string;
  updatedAt: string;
  identities: Identity[];
}

export interface Section {
  id: number;
  name: string;
  logo: string;
  admin: boolean;
}

export interface SectionDirectory {
  sections: (Section & { venues: string[] })[];
  systemAdmin: boolean;
}

export type InvitedSection = Pick<Section, "id" | "name" | "logo">;

export interface SectionInviteInput {
  expiresAt: string;
  maxUses: number;
}

export interface SectionInvite extends SectionInviteInput {
  id: number;
  uses: number;
  revokedAt: string | null;
  code?: string;
}

export interface HistoryItem {
  id: number;
  userId: number;
  name: string;
  trainingId: number;
  description: string;
  date: string;
  status: BookingStatus;
  section: string;
}
