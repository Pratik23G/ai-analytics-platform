import { useEffect, useState } from "react";

export default function App() {
  const [stats, setStats] = useState(null);
  const [predict, setPredict] = useState(null);
  const [err, setErr] = useState("");

  async function loadStats() {
    try {
      setErr("");
      const res = await fetch("http://localhost:8080/stats", {
        headers: { "X-API-Key": "demo-key" },
      });
      setStats(await res.json());
    } catch (e) {
      setErr(String(e));
    }
  }

  async function runPredict() {
    try {
      setErr("");
      const res = await fetch("http://localhost:8080/predict", {
        headers: { "X-API-Key": "demo-key" },
      });
      setPredict(await res.json());
      await loadStats();
    } catch (e) {
      setErr(String(e));
    }
  }

  useEffect(() => {
    loadStats();
    const t = setInterval(loadStats, 1000);
    return () => clearInterval(t);
  }, []);

  return (
    <div style={{ fontFamily: "system-ui", padding: 24, maxWidth: 900, margin: "0 auto" }}>
      <h1>AI Analytics Dashboard</h1>
      <p>Live stats from Redis via Go API Gateway</p>

      {err && <pre style={{ color: "crimson" }}>{err}</pre>}

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 12 }}>
        <Card title="Events Total" value={stats?.events_total ?? "-"} />
        <Card title="Positive Total" value={stats?.positive_total ?? "-"} />
        <Card title="Avg Score" value={stats?.score_avg?.toFixed?.(2) ?? "-"} />
      </div>

      <div style={{ marginTop: 20 }}>
        <button onClick={runPredict} style={{ padding: "10px 14px", cursor: "pointer" }}>
          Run /predict
        </button>
        {predict && (
          <pre style={{ marginTop: 12, background: "#111", color: "#0f0", padding: 12 }}>
            {JSON.stringify(predict, null, 2)}
          </pre>
        )}
      </div>
    </div>
  );
}

function Card({ title, value }) {
  return (
    <div style={{ border: "1px solid #ddd", borderRadius: 12, padding: 16 }}>
      <div style={{ fontSize: 12, opacity: 0.7 }}>{title}</div>
      <div style={{ fontSize: 28, fontWeight: 700 }}>{value}</div>
    </div>
  );
}
