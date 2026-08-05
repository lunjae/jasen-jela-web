import { useQuery } from "@tanstack/react-query";
import { useSearchParams } from "react-router-dom";
import { catalogApi } from "../../api/catalog";
import { ProductCard } from "../../components/ProductCard";
import { Spinner, State } from "../../components/ui";
export function CatalogPage() {
  const [params, setParams] = useSearchParams();
  const query = params.toString();
  const products = useQuery({
    queryKey: ["products", query],
    queryFn: () => catalogApi.products(query),
  });
  const categories = useQuery({
    queryKey: ["categories"],
    queryFn: catalogApi.categories,
  });
  function set(k: string, v: string) {
    const n = new URLSearchParams(params);
    if (v) n.set(k, v);
    else n.delete(k);
    n.delete("page");
    setParams(n);
  }
  return (
    <section className="section catalog">
      <div className="container">
        <div className="page-title">
          <p className="eyebrow">Katalog</p>
          <h1>Naši proizvodi</h1>
          <p>
            Pregledajte ponudu i pronađite model koji odgovara vašim potrebama.
          </p>
        </div>
        <div className="filters">
          <label>
            Pretraga
            <input
              value={params.get("search") || ""}
              onChange={(e) => set("search", e.target.value)}
              placeholder="Naziv proizvoda"
            />
          </label>
          <label>
            Kategorija
            <select
              value={params.get("category") || ""}
              onChange={(e) => set("category", e.target.value)}
            >
              <option value="">Sve kategorije</option>
              {categories.data?.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Materijal
            <input
              value={params.get("material") || ""}
              onChange={(e) => set("material", e.target.value)}
              placeholder="npr. hrast"
            />
          </label>
          <label>
            Boja
            <input
              value={params.get("color") || ""}
              onChange={(e) => set("color", e.target.value)}
              placeholder="npr. orah"
            />
          </label>
          <label>
            Sortiranje
            <select
              value={params.get("sort") || "newest"}
              onChange={(e) => set("sort", e.target.value)}
            >
              <option value="newest">Najnovije</option>
              <option value="name_asc">Naziv A–Š</option>
              <option value="name_desc">Naziv Š–A</option>
              <option value="price_asc">Cena rastuće</option>
              <option value="price_desc">Cena opadajuće</option>
            </select>
          </label>
        </div>
        {products.isLoading ? (
          <div className="page-center">
            <Spinner />
          </div>
        ) : products.isError ? (
          <State title="Nije moguće učitati proizvode">
            Proverite vezu i pokušajte ponovo.
          </State>
        ) : !products.data?.items.length ? (
          <State title="Nema rezultata">Promenite kriterijume pretrage.</State>
        ) : (
          <>
            <p className="result-count">
              Prikazano {products.data.items.length} od {products.data.total}{" "}
              proizvoda
            </p>
            <div className="cards">
              {products.data.items.map((p) => (
                <ProductCard key={p.id} product={p} />
              ))}
            </div>
            <div className="pagination">
              <button
                disabled={products.data.page <= 1}
                onClick={() => set("page", String(products.data!.page - 1))}
              >
                Prethodna
              </button>
              <span>
                Strana {products.data.page} od {products.data.totalPages || 1}
              </span>
              <button
                disabled={products.data.page >= products.data.totalPages}
                onClick={() => set("page", String(products.data!.page + 1))}
              >
                Sledeća
              </button>
            </div>
          </>
        )}
      </div>
    </section>
  );
}
