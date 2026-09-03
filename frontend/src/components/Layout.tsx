import type { ReactNode } from "react";
import { TabBar } from "./TabBar";

export function Layout({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="page">
      <header className="page__header">
        <h1>{title}</h1>
      </header>
      <main className="page__content">{children}</main>
      <TabBar />
    </div>
  );
}
