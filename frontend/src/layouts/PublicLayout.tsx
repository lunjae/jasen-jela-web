import { useState } from "react";
import { Link, NavLink, Outlet } from "react-router-dom";
export function PublicLayout() {
  const [open, setOpen] = useState(false);
  return (
    <>
      <header className="site-header">
        <div className="container nav">
          <Link className="brand" to="/">
            <span>JJ</span>
            <div>
              <strong>Jasen Jela</strong>
              <small>Tradicija. Kvalitet. Poverenje.</small>
            </div>
          </Link>
          <button
            className="menu"
            onClick={() => setOpen(!open)}
            aria-expanded={open}
            aria-label="Otvori navigaciju"
          >
            ☰
          </button>
          <nav className={open ? "open" : ""} aria-label="Glavna navigacija">
            {[
              ["/", "Početna"],
              ["/proizvodi", "Proizvodi"],
              ["/o-nama", "O nama"],
              ["/kontakt", "Kontakt"],
            ].map(([to, x]) => (
              <NavLink key={to} to={to} onClick={() => setOpen(false)}>
                {x}
              </NavLink>
            ))}
            <Link className="button small" to="/upit">
              Pošaljite upit
            </Link>
          </nav>
        </div>
      </header>
      <main>
        <Outlet />
      </main>
      <footer>
        <div className="container footer-grid">
          <div>
            <div className="brand light">
              <span>JJ</span>
              <strong>Jasen Jela</strong>
            </div>
            <p>Dostojanstvena izrada, pouzdan kvalitet i pažljiva usluga.</p>
          </div>
          <div>
            <strong>Navigacija</strong>
            <Link to="/proizvodi">Proizvodi</Link>
            <Link to="/o-nama">O nama</Link>
            <Link to="/kontakt">Kontakt</Link>
          </div>
          <div>
            <strong>Kontakt</strong>
            <a href="tel:+381000000000">+381 00 000 000</a>
            <a href="mailto:info@jasen-jela.rs">info@jasen-jela.rs</a>
            <span>Srbija</span>
          </div>
        </div>
        <div className="container copyright">
          © {new Date().getFullYear()} Jasen Jela. Sva prava zadržana.{" "}
          <Link to="/admin/prijava">Administracija</Link>
        </div>
      </footer>
    </>
  );
}
