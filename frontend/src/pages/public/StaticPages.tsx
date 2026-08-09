import { Link } from "react-router-dom";
export function AboutPage() {
  return (
    <>
      <section className="page-hero">
        <div className="container">
          <p className="eyebrow">O nama</p>
          <h1>Tradicija koja obavezuje</h1>
          <p>Naš rad počiva na zanatskom znanju, odgovornosti i poštovanju.</p>
        </div>
      </section>
      <section className="section">
        <div className="container prose">
          <h2>Jasen Jela</h2>
          <p>
            Godinama gradimo poverenje kvalitetnom izradom i pouzdanom
            saradnjom. Razumemo osetljivost trenutka i svakom zahtevu pristupamo
            diskretno, pažljivo i profesionalno.
          </p>
          <p>
            Naši proizvodi izrađuju se od proverenih materijala, uz kontrolu
            svake faze proizvodnje. Partnerima nudimo jasnu komunikaciju,
            stabilan kvalitet i poštovanje dogovorenih rokova.
          </p>
          <Link className="button" to="/kontakt">
            Razgovarajte sa nama
          </Link>
        </div>
      </section>
    </>
  );
}
export function ContactPage() {
  return (
    <>
      <section className="page-hero">
        <div className="container">
          <p className="eyebrow">Kontakt</p>
          <h1>Tu smo za sva pitanja</h1>
        </div>
      </section>
      <section className="section">
        <div className="container contact-grid">
          <div>
            <h2>Kontakt podaci</h2>
            <p>
              Obratite nam se za informacije o modelima, dostupnosti i uslovima
              saradnje.
            </p>
            <address>
              <strong>Telefon</strong>
              <a href="tel:+381000000000">+381 00 000 000</a>
              <strong>Email</strong>
              <a href="mailto:info@jasen-jela.rs">info@jasen-jela.rs</a>
              <strong>Adresa</strong>
              <span>Srbija</span>
            </address>
          </div>
          <div className="contact-card">
            <h2>Pošaljite upit</h2>
            <p>Odgovorićemo u najkraćem mogućem roku.</p>
            <Link className="button full" to="/upit">
              Otvorite formular
            </Link>
          </div>
        </div>
      </section>
    </>
  );
}
export function NotFoundPage() {
  return (
    <div className="not-found">
      <span>404</span>
      <h1>Stranica nije pronađena</h1>
      <p>Tražena stranica ne postoji ili je premeštena.</p>
      <Link className="button" to="/">
        Nazad na početnu
      </Link>
    </div>
  );
}
