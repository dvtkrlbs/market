import { Link, Outlet } from "@tanstack/react-router";

export default function App() {
  return (
    <>
      <header className="app-header">
        <nav aria-label="Customer navigation">
          <Link to="/" activeOptions={{ exact: true }}>
            Home
          </Link>
          <Link to="/stores">Stores</Link>
        </nav>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </>
  );
}
