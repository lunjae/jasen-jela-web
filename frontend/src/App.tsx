import { useEffect, useState } from "react";
import "./index.css";
import { getHealth } from "./services/api";

function App() {
  const [apiMessage, setApiMessage] = useState("Proveravanje backenda...");
  const [error, setError] = useState("");

  useEffect(() => {
    async function checkBackend() {
      try {
        const response = await getHealth();
        setApiMessage(response.message);
      } catch {
        setError("Nije moguće povezati se sa backendom.");
      }
    }

    checkBackend();
  }, []);

  return (
    <main>
      <h1>Jasen Jela</h1>
      <p>Web katalog</p>

      <h2>Status sistema</h2>

      {error ? (
        <p>{error}</p>
      ) : (
        <p>Backend: {apiMessage}</p>
      )}
    </main>
  );
}

export default App;