import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { catalogApi } from "../../api/catalog";
import { ProductCard } from "../../components/ProductCard";
import hero from "../../assets/hero.png";
export function HomePage() {
  const { data } = useQuery({
    queryKey: ["featured"],
    queryFn: () => catalogApi.products("pageSize=3&sort=newest"),
  });
  const { data: cats } = useQuery({
    queryKey: ["categories"],
    queryFn: catalogApi.categories,
  });
  return (
    <>
      <section className="hero">
        <div className="container hero-grid">
          <div>
            <p className="eyebrow">Pouzdana izrada od 1998.</p>
            <h1>Dostojanstvo u svakom detalju</h1>
            <p className="lead">
              Pažljivo izrađeni proizvodi od biranih materijala, uz
              profesionalnu podršku onda kada je najpotrebnija.
            </p>
            <div className="actions">
              <Link className="button" to="/proizvodi">
                Pogledajte proizvode
              </Link>
              <Link className="text-link" to="/kontakt">
                Kontaktirajte nas →
              </Link>
            </div>
          </div>
          <div className="hero-art">
            <img src={hero} alt="Detalj zanatske obrade drveta" />
          </div>
        </div>
      </section>
      <section className="section">
        <div className="container intro">
          <p className="eyebrow">Naša ponuda</p>
          <h2>Kvalitet kojem možete verovati</h2>
          <p>
            Jasen Jela spaja dugogodišnje iskustvo, proverene materijale i
            odgovoran odnos prema svakoj porodici i partneru.
          </p>
        </div>
        <div className="container cards">
          {data?.items.map((p) => (
            <ProductCard key={p.id} product={p} />
          ))}
        </div>
        <div className="center">
          <Link className="button outline" to="/proizvodi">
            Kompletan katalog
          </Link>
        </div>
      </section>
      <section className="section muted">
        <div className="container">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Kategorije</p>
              <h2>Pronađite odgovarajući model</h2>
            </div>
          </div>
          <div className="category-grid">
            {cats?.map((c) => (
              <Link key={c.id} to={`/proizvodi?category=${c.id}`}>
                <h3>{c.name}</h3>
                <p>{c.description || "Pogledajte dostupne modele."}</p>
                <span>Pogledajte ponudu →</span>
              </Link>
            ))}
          </div>
        </div>
      </section>
      <section className="section">
        <div className="container values">
          <div>
            <p className="eyebrow">Zašto Jasen Jela</p>
            <h2>Posvećenost bez kompromisa</h2>
          </div>
          {[
            [
              "01",
              "Provereni materijali",
              "Brižljivo biramo drvo i završne materijale.",
            ],
            [
              "02",
              "Precizna izrada",
              "Svaki detalj prolazi pažljivu kontrolu kvaliteta.",
            ],
            [
              "03",
              "Pouzdana isporuka",
              "Dogovoreni rokovi i jasna komunikacija.",
            ],
          ].map((x) => (
            <article key={x[0]}>
              <span>{x[0]}</span>
              <h3>{x[1]}</h3>
              <p>{x[2]}</p>
            </article>
          ))}
        </div>
      </section>
      <section className="cta">
        <div className="container">
          <div>
            <p className="eyebrow">Tu smo za vas</p>
            <h2>Potrebna vam je dodatna informacija?</h2>
          </div>
          <Link className="button light-button" to="/upit">
            Pošaljite upit
          </Link>
        </div>
      </section>
    </>
  );
}
