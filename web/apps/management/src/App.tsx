import { Link, Outlet } from "@tanstack/react-router";

export default function App() {
  return (
    <>
      <header className="app-header">
        <nav aria-label="Management navigation">
          <Link to="/" activeOptions={{ exact: true }}>
            Home
          </Link>
          <Link to="/apply">Apply for a store</Link>
        </nav>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </>
  );
}
