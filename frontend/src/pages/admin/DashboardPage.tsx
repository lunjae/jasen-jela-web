import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { adminApi } from "../../api/catalog";
import { Spinner } from "../../components/ui";
export function DashboardPage() {
  const p = useQuery({
    queryKey: ["admin-products"],
    queryFn: () => adminApi.products("pageSize=5"),
  });
  const i = useQuery({
    queryKey: ["admin-inquiries"],
    queryFn: () => adminApi.inquiries("pageSize=5"),
  });
  return (
    <>
      <div className="admin-heading">
        <div>
          <p className="eyebrow">Administracija</p>
          <h1>Pregled</h1>
        </div>
      </div>
      {p.isLoading || i.isLoading ? (
        <Spinner />
      ) : (
        <>
          <div className="stats">
            <article>
              <span>Proizvodi</span>
              <strong>{p.data?.total || 0}</strong>
              <Link to="/admin/proizvodi">Upravljaj →</Link>
            </article>
            <article>
              <span>Upiti</span>
              <strong>{i.data?.total || 0}</strong>
              <Link to="/admin/upiti">Pregledaj →</Link>
            </article>
            <article>
              <span>Novi upiti</span>
              <strong>
                {i.data?.items.filter((x) => x.status === "new").length || 0}
              </strong>
              <Link to="/admin/upiti?status=new">Otvori →</Link>
            </article>
          </div>
          <section className="admin-panel">
            <div className="panel-heading">
              <h2>Najnoviji upiti</h2>
              <Link to="/admin/upiti">Svi upiti</Link>
            </div>
            {i.data?.items.map((x) => (
              <Link
                className="inquiry-row"
                key={x.id}
                to={`/admin/upiti/${x.id}`}
              >
                <span className={`badge ${x.status}`}>{x.status}</span>
                <strong>{x.fullName}</strong>
                <span>{x.email}</span>
              </Link>
            ))}
          </section>
        </>
      )}
    </>
  );
}
