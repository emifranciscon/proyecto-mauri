import { NavLink, Outlet } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const linkStyle = ({ isActive }: { isActive: boolean }) => ({
  padding: "0.45rem 0.75rem",
  borderRadius: 8,
  color: "var(--text)",
  textDecoration: "none",
  background: isActive ? "var(--surface2)" : "transparent",
  border: `1px solid ${isActive ? "var(--border)" : "transparent"}`,
});

export function Layout() {
  const { user, logout } = useAuth();
  return (
    <div style={{ minHeight: "100vh", display: "flex", flexDirection: "column" }}>
      <header
        style={{
          borderBottom: "1px solid var(--border)",
          background: "var(--surface)",
          padding: "0.75rem 1.25rem",
          display: "flex",
          alignItems: "center",
          gap: "1rem",
          flexWrap: "wrap",
        }}
      >
        <span
          style={{
            fontFamily: "Fraunces, serif",
            fontWeight: 600,
            fontSize: "1.15rem",
            marginRight: "0.5rem",
          }}
        >
          Sistema de control de inventario. Para Mauri.
        </span>
        <nav style={{ display: "flex", gap: "0.35rem", flex: 1 }}>
          <NavLink to="/asientos" style={linkStyle}>
            Asientos
          </NavLink>
          <NavLink to="/tanques" style={linkStyle}>
            Tanques
          </NavLink>
          <NavLink to="/balanzas" style={linkStyle}>
            Balanzas
          </NavLink>
        </nav>
        <span style={{ color: "var(--muted)", fontSize: "0.9rem" }}>
          {user?.email}
        </span>
        <button type="button" className="btn btn-ghost" onClick={logout}>
          Salir
        </button>
      </header>
      <main style={{ padding: "1.25rem", maxWidth: 1200, margin: "0 auto", width: "100%" }}>
        <Outlet />
      </main>
    </div>
  );
}
