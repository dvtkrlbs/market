import { Heading } from "@astryxdesign/core/Heading";
import { Link, Outlet } from "@tanstack/react-router";

export default function App() {
  return (
    <div className="admin-layout">
      <aside className="sidebar">
        <Heading level={2}>Market Admin</Heading>
        <nav aria-label="Admin navigation">
          <Link to="/" activeOptions={{ exact: true }}>
            Overview
          </Link>
          <Link to="/stores">Stores</Link>
        </nav>
      </aside>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
