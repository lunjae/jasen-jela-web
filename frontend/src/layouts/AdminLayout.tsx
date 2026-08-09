import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
export function AdminLayout() {
  const { user, logout } = useAuth();
  const nav = useNavigate();
  return (
    <div className="admin-shell">
      <aside>
        <NavLink to="/admin" className="admin-brand">
          Jasen Jela <small>Administracija</small>
        </NavLink>
        <nav>
          <NavLink end to="/admin">
            Pregled
          </NavLink>
          <NavLink to="/admin/proizvodi">Proizvodi</NavLink>
          <NavLink to="/admin/kategorije">Kategorije</NavLink>
          <NavLink to="/admin/upiti">Upiti</NavLink>
        </nav>
        <div className="admin-user">
          <small>{user?.email}</small>
          <button
            onClick={async () => {
              await logout();
              nav("/admin/prijava");
            }}
          >
            Odjavi se
          </button>
        </div>
      </aside>
      <section className="admin-content">
        <Outlet />
      </section>
    </div>
  );
}
