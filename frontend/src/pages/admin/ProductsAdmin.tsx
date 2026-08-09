import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useState } from "react";
import { adminApi } from "../../api/catalog";
import { Modal, Spinner, State } from "../../components/ui";
import type { Product } from "../../types";
import { money } from "../../utils/format";
export function ProductListPage() {
  const qc = useQueryClient();
  const [remove, setRemove] = useState<Product | null>(null);
  const p = useQuery({
    queryKey: ["admin-products"],
    queryFn: () => adminApi.products("pageSize=100"),
  });
  const del = useMutation({
    mutationFn: adminApi.deleteProduct,
    onSuccess: () => {
      setRemove(null);
      qc.invalidateQueries({ queryKey: ["admin-products"] });
    },
  });
  return (
    <>
      <div className="admin-heading">
        <div>
          <p className="eyebrow">Katalog</p>
          <h1>Proizvodi</h1>
        </div>
        <Link className="button" to="/admin/proizvodi/novi">
          Novi proizvod
        </Link>
      </div>
      {p.isLoading ? (
        <Spinner />
      ) : p.isError ? (
        <State title="Greška pri učitavanju" />
      ) : (
        <div className="admin-panel table-wrap">
          <table>
            <thead>
              <tr>
                <th>Proizvod</th>
                <th>Materijal</th>
                <th>Cena</th>
                <th>Status</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {p.data?.items.map((x) => (
                <tr key={x.id}>
                  <td>
                    <strong>{x.name}</strong>
                    <small>{x.slug}</small>
                  </td>
                  <td>{x.material}</td>
                  <td>
                    {x.price !== undefined
                      ? money(x.price, x.currency)
                      : "Na upit"}
                  </td>
                  <td>
                    <span
                      className={`badge ${x.published ? "published" : "draft"}`}
                    >
                      {x.published ? "Objavljen" : "Skica"}
                    </span>
                  </td>
                  <td className="row-actions">
                    <Link to={`/admin/proizvodi/${x.id}`}>Uredi</Link>
                    <button onClick={() => setRemove(x)}>Obriši</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!p.data?.items.length && <State title="Još nema proizvoda" />}
        </div>
      )}
      {remove && (
        <Modal title="Obrišite proizvod?" onClose={() => setRemove(null)}>
          <p>Proizvod „{remove.name}” biće trajno obrisan.</p>
          <div className="actions">
            <button onClick={() => setRemove(null)}>Odustani</button>
            <button
              className="button danger"
              onClick={() => del.mutate(remove.id)}
              disabled={del.isPending}
            >
              Obriši
            </button>
          </div>
        </Modal>
      )}
    </>
  );
}
