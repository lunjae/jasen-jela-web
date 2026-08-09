import { useState } from "react";
import type { ProductImage } from "../types";
export function ImageGallery({
  images,
  name,
}: {
  images: ProductImage[];
  name: string;
}) {
  const sorted = [...images].sort((a, b) => a.order - b.order);
  const [active, setActive] = useState(
    sorted.find((x) => x.isPrimary) || sorted[0],
  );
  if (!active) return <div className="gallery-empty">Jasen Jela</div>;
  return (
    <div className="gallery">
      <img className="gallery-main" src={active.url} alt={active.alt || name} />
      {sorted.length > 1 && (
        <div className="thumbnails">
          {sorted.map((x) => (
            <button
              key={x.publicId}
              onClick={() => setActive(x)}
              aria-label={`Prikaži ${x.alt || name}`}
              className={x.publicId === active.publicId ? "active" : ""}
            >
              <img src={x.url} alt="" />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
