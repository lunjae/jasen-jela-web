import { Link } from "react-router-dom";
import type { Product } from "../types";
import { money } from "../utils/format";
export function ProductCard({ product }: { product: Product }) {
  const image = product.images.find((x) => x.isPrimary) || product.images[0];
  return (
    <article className="product-card">
      <Link to={`/proizvodi/${product.slug}`} className="product-image">
        {image?.url ? (
          <img src={image.url} alt={image.alt || product.name} />
        ) : (
          <span>Jasen Jela</span>
        )}
      </Link>
      <div className="product-card-body">
        <p className="eyebrow">
          {product.material} · {product.color}
        </p>
        <h3>
          <Link to={`/proizvodi/${product.slug}`}>{product.name}</Link>
        </h3>
        <p>{product.shortDescription}</p>
        <div className="card-footer">
          <strong>
            {product.price !== undefined
              ? money(product.price, product.currency)
              : "Cena na upit"}
          </strong>
          <Link
            to={`/proizvodi/${product.slug}`}
            aria-label={`Detalji za ${product.name}`}
          >
            Detalji →
          </Link>
        </div>
      </div>
    </article>
  );
}
