import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import type { BookingStatus, Section, Training } from "../api/types";
import { useProfile } from "../auth/Profile";
import { useSection } from "../auth/Sections";
import { t } from "../i18n";
import { isoToLocalDateTimeInput } from "../utils/date";

export interface PersonalSession {
  training: Training;
  section: Section;
  status: BookingStatus | "waiting";
  bookingId?: number;
}

export function useOpenSession() {
  const { choose } = useSection();
  const navigate = useNavigate();
  return ({ section, training }: PersonalSession) => {
    choose(section.id);
    const day = isoToLocalDateTimeInput(training.date).slice(0, 10);
    navigate(`/?day=${day}&session=${training.id}`);
  };
}

export function usePersonalSessions(all: boolean, attendance = false) {
  const { current, directory } = useSection();
  const { profile } = useProfile();
  const [sessions, setSessions] = useState<PersonalSession[] | null>(null);
  const [error, setError] = useState("");
  const [version, setVersion] = useState(0);

  useEffect(() => {
    let active = true;
    setError("");
    setSessions(null);
    async function load(section: Section) {
      const [bookings, trainings, waiting] = await Promise.all([
        api.getBookingsByUserId(profile.id, section.id),
        api.getTrainings(section.id),
        attendance ? [] : api.getWaitlist(section.id),
      ]);
      const byId = new Map(trainings.map((training) => [training.id, training]));
      const rows: PersonalSession[] = [];
      const visible = bookings.filter((booking) =>
        attendance
          ? booking.status === "completed" || booking.status === "not_completed"
          : booking.status === "active",
      );
      const needed = new Set([...visible.map((booking) => booking.trainingId), ...waiting]);
      await Promise.all(
        [...needed]
          .filter((id) => !byId.has(id))
          .map(async (id) => {
            byId.set(id, await api.getTraining(id, section.id));
          }),
      );
      for (const booking of visible) {
        rows.push({
          section,
          training: byId.get(booking.trainingId)!,
          status: booking.status,
          bookingId: booking.id,
        });
      }
      for (const id of waiting) {
        const training = byId.get(id);
        if (training) rows.push({ section, training, status: "waiting" });
      }
      return rows;
    }
    Promise.all((all ? directory.sections : [current]).map(load))
      .then((groups) => {
        const rows = groups.flat().sort((a, b) => Date.parse(a.training.date) - Date.parse(b.training.date));
        if (active) setSessions(attendance ? rows.reverse() : rows);
      })
      .catch((reason) => {
        if (active) setError(reason instanceof Error ? reason.message : t("Не удалось загрузить записи"));
      });
    return () => {
      active = false;
    };
  }, [all, attendance, current, directory, profile.id, version]);
  return { sessions, error, reload: () => setVersion((value) => value + 1) };
}
