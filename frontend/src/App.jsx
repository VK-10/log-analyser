import { useState } from "react";

function App() {
  const [logs, setLogs] = useState("");
  const [results, setResults] = useState([]);

  const handleClassify = async () => {
    const lines = logs.split("\n").filter(Boolean);

    const payload = lines.map((line) => ({
      source: "App1",
      log_message: line,
    }));

    const res = await fetch("http://localhost:8080/classify", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    const data = await res.json();
    setResults(data);
  };

  return (
    <div style={{ padding: "20px", fontFamily: "sans-serif" }}>
      <h2>Log Classifier</h2>

      <textarea
        rows="10"
        cols="60"
        placeholder="Paste logs here, one per line"
        value={logs}
        onChange={(e) => setLogs(e.target.value)}
      />

      <br />
      <button onClick={handleClassify} style={{ marginTop: "10px" }}>
        Classify Logs
      </button>

      <h3>Results</h3>
      <pre>{JSON.stringify(results, null, 2)}</pre>
    </div>
  );
}

export default App;
