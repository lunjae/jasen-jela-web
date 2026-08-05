import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { Spinner } from "../components/ui";
export function ProtectedRoute() {
  const { user, loading } = useAuth();
  const loc = useLocation();
  if (loading)
    return (
      <div className="page-center">
        <Spinner />
      </div>
    );
  if (!user && !import.meta.env.DEV)
    return <Navigate to="/admin/prijava" state={{ from: loc }} replace />;
  return <Outlet />;
}
