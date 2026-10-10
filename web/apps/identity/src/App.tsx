import { Heading } from "@astryxdesign/core/Heading";
import * as stylex from "@stylexjs/stylex";
import { Outlet } from "@tanstack/react-router";

const styles = stylex.create({
  page: {
    minHeight: "100dvh",
    display: "grid",
    placeItems: "center",
    padding: "var(--spacing-6)",
  },
  panel: {
    width: "100%",
    maxWidth: 420,
    display: "flex",
    flexDirection: "column",
    gap: "var(--spacing-6)",
  },
});

export default function App() {
  return (
    <main {...stylex.props(styles.page)}>
      <section aria-label="Market identity" {...stylex.props(styles.panel)}>
        <Heading level={1}>Market identity</Heading>
        <Outlet />
      </section>
    </main>
  );
}
