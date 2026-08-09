import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { adminApi } from "../../api/catalog";
import { Spinner, State } from "../../components/ui";
import { date } from "../../utils/format";
const labels = {
  new: "Nov",
  read: "Pročitan",
  contacted: "Kontaktiran",
  closed: "Zatvoren",
};
export function InquiryListPage() {
  const [p, setP] = useSearchParams();
  const status = p.get("status") || "";
  const q = useQuery({
    queryKey: ["admin-inquiries", status],
    queryFn: () => adminApi.inquiries(status ? `status=${status}` : ""),
  });
  return (
    <>
      <div className="admin-heading">
        <div>
          <p className="eyebrow">Komunikacija</p>
          <h1>Upiti</h1>
        </div>
        <select
          value={status}
          onChange={(e) =>
            setP(e.target.value ? { status: e.target.value } : {})
          }
        >
          <option value="">Svi statusi</option>
          {Object.entries(labels).map(([k, v]) => (
            <option key={k} value={k}>
              {v}
            </option>
          ))}
        </select>
      </div>
      {q.isLoading ? (
        <Spinner />
      ) : q.isError ? (
        <State title="Upiti nisu dostupni" />
      ) : (
        <div className="admin-panel table-wrap">
          <table>
            <thead>
              <tr>
                <th>Pošiljalac</th>
                <th>Kontakt</th>
                <th>Datum</th>
                <th>Status</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {q.data?.items.map((i) => (
                <tr key={i.id}>
                  <td>
                    <strong>{i.fullName}</strong>
                  </td>
                  <td>
                    {i.email}
                    <small>{i.phone}</small>
                  </td>
                  <td>{date(i.createdAt)}</td>
                  <td>
                    <span className={`badge ${i.status}`}>
                      {labels[i.status]}
                    </span>
                  </td>
                  <td>
                    <Link to={`/admin/upiti/${i.id}`}>Otvori</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!q.data?.items.length && <State title="Nema upita" />}
        </div>
      )}
    </>
  );
}
export function InquiryDetailPage() {
  const { id = "" } = useParams();
  const qc = useQueryClient();
  const q = useQuery({
    queryKey: ["admin-inquiry", id],
    queryFn: () => adminApi.inquiry(id),
  });
  const status = useMutation({
    mutationFn: (s: string) => adminApi.status(id, s),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-inquiry", id] });
      qc.invalidateQueries({ queryKey: ["admin-inquiries"] });
    },
  });
  if (q.isLoading) return <Spinner />;
  if (!q.data) return <State title="Upit nije pronađen" />;
  const i = q.data;
  return (
    <>
      <div className="admin-heading">
        <div>
          <Link className="back" to="/admin/upiti">
            ← Svi upiti
          </Link>
          <h1>{i.fullName}</h1>
        </div>
        <select
          value={i.status}
          onChange={(e) => status.mutate(e.target.value)}
        >
          {Object.entries(labels).map(([k, v]) => (
            <option key={k} value={k}>
              {v}
            </option>
          ))}
        </select>
      </div>
      <div className="inquiry-detail">
        <section className="admin-panel">
          <h2>Poruka</h2>
          <p className="message">{i.message}</p>
        </section>
        <aside className="admin-panel">
          <h2>Podaci</h2>
          <dl>
            <dt>Email</dt>
            <dd>
              <a href={`mailto:${i.email}`}>{i.email}</a>
            </dd>
            <dt>Telefon</dt>
            <dd>
              <a href={`tel:${i.phone}`}>{i.phone}</a>
            </dd>
            <dt>Poslato</dt>
            <dd>{date(i.createdAt)}</dd>
            {i.productId && (
              <>
                <dt>Proizvod ID</dt>
                <dd>{i.productId}</dd>
              </>
            )}
          </dl>
        </aside>
      </div>
    </>
  );
}
