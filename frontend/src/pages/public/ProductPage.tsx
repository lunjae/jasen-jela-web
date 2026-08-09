import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { catalogApi } from "../../api/catalog";
import { ImageGallery } from "../../components/ImageGallery";
import { ProductCard } from "../../components/ProductCard";
import { Spinner, State } from "../../components/ui";
import { money } from "../../utils/format";
export function ProductPage() {
  const { slug = "" } = useParams();
  const p = useQuery({
    queryKey: ["product", slug],
    queryFn: () => catalogApi.product(slug),
  });
  const related = useQuery({
    queryKey: ["related", p.data?.categoryId],
    queryFn: () =>
      catalogApi.products(`category=${p.data!.categoryId}&pageSize=3`),
    enabled: !!p.data,
  });
  if (p.isLoading)
    return (
      <div className="page-center">
        <Spinner />
      </div>
    );
  if (p.isError || !p.data)
    return (
      <State title="Proizvod nije pronađen">
        Proverite adresu ili se vratite u katalog.
      </State>
    );
  const x = p.data;
  return (
    <section className="section">
      <div className="container">
        <Link className="back" to="/proizvodi">
          ← Nazad na katalog
        </Link>
        <div className="product-detail">
          <ImageGallery images={x.images} name={x.name} />
          <div>
            <p className="eyebrow">
              {x.material} · {x.color}
            </p>
            <h1>{x.name}</h1>
            <p className="lead">{x.shortDescription}</p>
            <strong className="price">
              {x.price !== undefined
                ? money(x.price, x.currency)
                : "Cena na upit"}
            </strong>
            <dl>
              <div>
                <dt>Materijal</dt>
                <dd>{x.material}</dd>
              </div>
              <div>
                <dt>Boja</dt>
                <dd>{x.color}</dd>
              </div>
              {x.dimensions && (
                <div>
                  <dt>Dimenzije</dt>
                  <dd>
                    {[
                      x.dimensions.length,
                      x.dimensions.width,
                      x.dimensions.height,
                    ]
                      .filter(Boolean)
                      .join(" × ")}{" "}
                    cm
                  </dd>
                </div>
              )}
              <div>
                <dt>Dostupnost</dt>
                <dd>{x.published ? "Dostupno za upit" : "Nije dostupno"}</dd>
              </div>
            </dl>
            <Link className="button full" to={`/upit?productId=${x.id}`}>
              Pošaljite upit za ovaj proizvod
            </Link>
          </div>
        </div>
        <div className="description">
          <h2>Opis proizvoda</h2>
          <p>{x.description}</p>
        </div>
        {related.data?.items.filter((y) => y.id !== x.id).length ? (
          <div>
            <div className="section-heading">
              <h2>Srodni proizvodi</h2>
            </div>
            <div className="cards">
              {related.data.items
                .filter((y) => y.id !== x.id)
                .slice(0, 3)
                .map((y) => (
                  <ProductCard key={y.id} product={y} />
                ))}
            </div>
          </div>
        ) : null}
      </div>
    </section>
  );
}
