import { createContext, useContext } from "react";
import type { User } from "../api/types";

export const ProfileContext = createContext<{ profile: User; setProfile: (profile: User) => void } | null>(
  null,
);

export function useProfile() {
  const value = useContext(ProfileContext);
  if (!value) throw new Error("ProfileContext is required");
  return value;
}
