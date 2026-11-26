import { t as translate } from "../i18n";
import { createContext, useContext, useState, useEffect, useCallback } from "react";
import type { ReactNode } from "react";
import { api, setSelectedSection } from "../api/client";
import type { SectionDirectory } from "../api/types";
import { ErrorBanner } from "../components/ErrorBanner";

const Sections = createContext<{
  current: SectionDirectory["sections"][number] | undefined;
  directory: SectionDirectory;
  choose: (id: number) => void;
  reload: (preferredId?: number) => Promise<void>;
} | null>(null);

export function SectionProvider({ children }: { children: ReactNode }) {
  const [directory, setDirectory] = useState<SectionDirectory | null>(null);
  const [id, setId] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [attempt, setAttempt] = useState(0);
  const choose = useCallback((next: number | null) => {
    setSelectedSection(next);
    if (next === null) localStorage.removeItem("prodapp.section");
    else localStorage.setItem("prodapp.section", String(next));
    setId(next);
  }, []);
  const applyDirectory = useCallback(
    (result: SectionDirectory, preferredId?: number) => {
      const sections = result.sections ?? [];
      const savedId = preferredId ?? Number(localStorage.getItem("prodapp.section"));
      const match = sections.find((section) => section.id === savedId);
      if (preferredId !== undefined && !match) throw new Error(translate("Не удалось загрузить секции"));
      choose((match ?? sections[0])?.id ?? null);
      setDirectory({ ...result, sections });
    },
    [choose],
  );
  const reload = useCallback(
    async (preferredId?: number) => applyDirectory(await api.sections(), preferredId),
    [applyDirectory],
  );
  useEffect(() => {
    let active = true;
    setError("");
    api
      .sections()
      .then((result) => {
        if (active) applyDirectory(result);
      })
      .catch(() => {
        if (active) setError(translate("Не удалось загрузить секции"));
      });
    return () => {
      active = false;
    };
  }, [applyDirectory, attempt]);
  if (error) return <ErrorBanner message={error} onRetry={() => setAttempt((value) => value + 1)} />;
  if (!directory) return <p role="status">{translate("Загружаем секции…")}</p>;
  const current = directory.sections.find((s) => s.id === id);
  return (
    <Sections.Provider value={{ current, directory, choose, reload }}>
      {children}
    </Sections.Provider>
  );
}
export function useSectionDirectory() {
  const value = useContext(Sections);
  if (!value) throw new Error("SectionProvider is required");
  return value;
}

export function useSection() {
  const value = useSectionDirectory();
  if (!value.current) throw new Error("A selected section is required");
  return { ...value, current: value.current };
}
