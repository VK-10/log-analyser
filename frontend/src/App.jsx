import { useState, useRef } from "react";

const LABEL_META = {
  USER_ACTION:          { color: "#4ade80", bg: "rgba(74,222,128,0.10)", icon: "◈" },
  SYSTEM_NOTIFICATION:  { color: "#60a5fa", bg: "rgba(96,165,250,0.10)", icon: "◉" },
  ERROR:                { color: "#f87171", bg: "rgba(248,113,113,0.10)", icon: "⬡" },
  WARNING:              { color: "#fbbf24", bg: "rgba(251,191,36,0.10)",  icon: "△" },
  UNCLASSIFIED:         { color: "#6b7280", bg: "rgba(107,114,128,0.10)", icon: "◌" },
};

const SOURCE_COLOR = {
  regex:        "#a78bfa",
  classifier:   "#34d399",
  llm:          "#fb923c",
  orchestrator: "#6b7280",
};

function getLabelMeta(labelID) {
  return LABEL_META[labelID] || LABEL_META.UNCLASSIFIED;
}

function ConfidenceBar({ value }) {
  const pct = Math.round((value || 0) * 100);
  const color =
    pct >= 80 ? "#4ade80" :
    pct >= 50 ? "#fbbf24" : "#f87171";
  return (
    <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
      <div style={{
        width: "80px", height: "4px",
        background: "rgba(255,255,255,0.08)",
        borderRadius: "2px", overflow: "hidden",
      }}>
        <div style={{
          width: `${pct}%`, height: "100%",
          background: color,
          transition: "width 0.4s ease",
        }} />
      </div>
      <span style={{ fontSize: "11px", color, fontVariantNumeric: "tabular-nums", minWidth: "32px" }}>
        {pct}%
      </span>
    </div>
  );
}

function ResultRow({ result, logLine, index }) {
  const meta = getLabelMeta(result?.label_id);
  const srcColor = SOURCE_COLOR[result?.source] || "#6b7280";

  return (
    <div style={{
      display: "grid",
      gridTemplateColumns: "28px 1fr 160px 90px 110px",
      alignItems: "center",
      gap: "0 16px",
      padding: "10px 16px",
      borderBottom: "1px solid rgba(255,255,255,0.05)",
      transition: "background 0.15s",
      background: "transparent",
    }}
    onMouseEnter={e => e.currentTarget.style.background = "rgba(255,255,255,0.03)"}
    onMouseLeave={e => e.currentTarget.style.background = "transparent"}
    >
      {/* index */}
      <span style={{ color: "rgba(255,255,255,0.2)", fontSize: "11px", fontVariantNumeric: "tabular-nums" }}>
        {String(index + 1).padStart(2, "0")}
      </span>

      {/* log message */}
      <span style={{
        fontSize: "12px", color: "rgba(255,255,255,0.6)",
        fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
        overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap",
      }} title={logLine}>
        {logLine}
      </span>

      {/* label */}
      <span style={{
        display: "inline-flex", alignItems: "center", gap: "6px",
        padding: "3px 10px",
        background: meta.bg,
        border: `1px solid ${meta.color}33`,
        borderRadius: "4px",
        color: meta.color,
        fontSize: "11px", fontWeight: 600, letterSpacing: "0.04em",
        whiteSpace: "nowrap",
      }}>
        <span>{meta.icon}</span>
        {result?.label || "—"}
      </span>

      {/* source */}
      <span style={{
        fontSize: "11px", color: srcColor, fontWeight: 500,
        letterSpacing: "0.05em", textTransform: "uppercase",
      }}>
        {result?.source || "—"}
      </span>

      {/* confidence bar */}
      <ConfidenceBar value={result?.confidence} />
    </div>
  );
}

function StatBadge({ label, value, color }) {
  return (
    <div style={{
      display: "flex", flexDirection: "column", alignItems: "center",
      gap: "2px", padding: "10px 20px",
      background: "rgba(255,255,255,0.04)",
      border: "1px solid rgba(255,255,255,0.08)",
      borderRadius: "8px",
    }}>
      <span style={{ fontSize: "20px", fontWeight: 700, color: color || "#fff", letterSpacing: "-0.02em" }}>
        {value}
      </span>
      <span style={{ fontSize: "10px", color: "rgba(255,255,255,0.35)", letterSpacing: "0.08em", textTransform: "uppercase" }}>
        {label}
      </span>
    </div>
  );
}

export default function App() {
  const [logs, setLogs] = useState("");
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [elapsed, setElapsed] = useState(null);
  const textareaRef = useRef();

  const lines = logs.split("\n").filter(Boolean);

  const labelCounts = results.reduce((acc, r) => {
    if (r?.label_id) acc[r.label_id] = (acc[r.label_id] || 0) + 1;
    return acc;
  }, {});

  const handleClassify = async () => {
    if (!lines.length) return;
    setLoading(true);
    setError(null);
    setResults([]);
    const t0 = performance.now();

    try {
      const payload = lines.map((line) => ({ source: "ui", log_message: line }));
      const res = await fetch("http://localhost:8080/classify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setResults(data || []);
      setElapsed(((performance.now() - t0) / 1000).toFixed(3));
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleClear = () => {
    setLogs("");
    setResults([]);
    setError(null);
    setElapsed(null);
  };

  return (
  <div style={{
    height: "100vh",           // full height
    width: "100vw",            // full width
    background: "#0d0d0f",
    color: "#e5e7eb",
    fontFamily: "'JetBrains Mono', 'Fira Code', 'Courier New', monospace",
    padding: "40px 32px",
  }}>

      {/* Header */}
      <div style={{ marginBottom: "32px" }}>
        <div style={{ display: "flex", alignItems: "baseline", gap: "12px", marginBottom: "6px" }}>
          <span style={{ fontSize: "11px", color: "#4ade80", letterSpacing: "0.2em", textTransform: "uppercase" }}>
            ▸ SYS.LOG
          </span>
          <span style={{ fontSize: "11px", color: "rgba(255,255,255,0.2)" }}>v1.0</span>
        </div>
        <h1 style={{
          fontSize: "28px", fontWeight: 800, margin: 0,
          letterSpacing: "-0.03em", color: "#fff",
        }}>
          Log Classifier
        </h1>
        <p style={{ fontSize: "12px", color: "rgba(255,255,255,0.3)", margin: "6px 0 0", fontFamily: "sans-serif" }}>
          Paste server logs below — one entry per line
        </p>
      </div>

      {/* Input panel */}
      <div style={{
        border: "1px solid rgba(255,255,255,0.1)",
        borderRadius: "10px",
        overflow: "hidden",
        marginBottom: "24px",
      }}>
        <div style={{
          display: "flex", alignItems: "center", justifyContent: "space-between",
          padding: "8px 14px",
          background: "rgba(255,255,255,0.04)",
          borderBottom: "1px solid rgba(255,255,255,0.06)",
        }}>
          <span style={{ fontSize: "11px", color: "rgba(255,255,255,0.3)", letterSpacing: "0.05em" }}>
            INPUT.LOG — {lines.length} line{lines.length !== 1 ? "s" : ""}
          </span>
          <button onClick={handleClear} style={{
            background: "none", border: "none", color: "rgba(255,255,255,0.25)",
            fontSize: "11px", cursor: "pointer", padding: "2px 6px", letterSpacing: "0.05em",
          }}
          onMouseEnter={e => e.currentTarget.style.color = "rgba(255,255,255,0.6)"}
          onMouseLeave={e => e.currentTarget.style.color = "rgba(255,255,255,0.25)"}
          >
            CLEAR ✕
          </button>
        </div>
        <textarea
          ref={textareaRef}
          rows={8}
          placeholder={`User admin logged in\nBackup started at 2024-01-15 03:00:00\nSystem updated to version 2.4.1\nFile report.pdf uploaded successfully by user john`}
          value={logs}
          onChange={(e) => setLogs(e.target.value)}
          style={{
            width: "100%", boxSizing: "border-box",
            background: "#111113", color: "rgba(255,255,255,0.7)",
            border: "none", outline: "none", resize: "vertical",
            fontFamily: "'JetBrains Mono', monospace", fontSize: "12.5px",
            lineHeight: "1.7", padding: "16px",
            caretColor: "#4ade80",
          }}
        />
      </div>

      {/* Classify button */}
      <button
        onClick={handleClassify}
        disabled={loading || !lines.length}
        style={{
          padding: "11px 28px",
          background: loading || !lines.length ? "rgba(74,222,128,0.1)" : "#4ade80",
          color: loading || !lines.length ? "rgba(74,222,128,0.4)" : "#0d0d0f",
          border: `1px solid ${loading || !lines.length ? "rgba(74,222,128,0.2)" : "#4ade80"}`,
          borderRadius: "8px",
          fontSize: "12px", fontWeight: 700, letterSpacing: "0.1em",
          textTransform: "uppercase", cursor: loading || !lines.length ? "not-allowed" : "pointer",
          transition: "all 0.2s",
          fontFamily: "inherit",
          marginBottom: "32px",
        }}
      >
        {loading ? "◌  Classifying…" : "▸  Run Classifier"}
      </button>

      {/* Error */}
      {error && (
        <div style={{
          marginBottom: "20px", padding: "12px 16px",
          background: "rgba(248,113,113,0.08)", border: "1px solid rgba(248,113,113,0.25)",
          borderRadius: "8px", fontSize: "12px", color: "#f87171",
        }}>
          ⚠ {error}
        </div>
      )}

      {/* Stats row */}
      {results.length > 0 && (
        <div style={{ display: "flex", gap: "12px", flexWrap: "wrap", marginBottom: "24px" }}>
          <StatBadge label="Total" value={results.length} color="#fff" />
          {Object.entries(labelCounts).map(([id, count]) => (
            <StatBadge
              key={id}
              label={id.replace(/_/g, " ")}
              value={count}
              color={getLabelMeta(id).color}
            />
          ))}
          {elapsed && (
            <StatBadge label="Latency" value={`${elapsed}s`} color="#60a5fa" />
          )}
        </div>
      )}

      {/* Results table */}
      {results.length > 0 && (
        <div style={{
          border: "1px solid rgba(255,255,255,0.08)",
          borderRadius: "10px",
          overflow: "hidden",
        }}>
          {/* Table header */}
          <div style={{
            display: "grid",
            gridTemplateColumns: "28px 1fr 160px 90px 110px",
            gap: "0 16px",
            padding: "8px 16px",
            background: "rgba(255,255,255,0.04)",
            borderBottom: "1px solid rgba(255,255,255,0.08)",
          }}>
            {["#", "MESSAGE", "LABEL", "SOURCE", "CONFIDENCE"].map((h) => (
              <span key={h} style={{
                fontSize: "10px", color: "rgba(255,255,255,0.25)",
                letterSpacing: "0.1em", fontWeight: 600,
              }}>{h}</span>
            ))}
          </div>

          {/* Rows */}
          {results.map((result, i) => (
            <ResultRow
              key={i}
              result={result}
              logLine={lines[i] || "—"}
              index={i}
            />
          ))}
        </div>
      )}

      {/* Legend */}
      {results.length > 0 && (
        <div style={{
          display: "flex", gap: "20px", flexWrap: "wrap",
          marginTop: "20px", padding: "12px 0",
          borderTop: "1px solid rgba(255,255,255,0.06)",
        }}>
          <span style={{ fontSize: "10px", color: "rgba(255,255,255,0.2)", letterSpacing: "0.05em" }}>SOURCE:</span>
          {Object.entries(SOURCE_COLOR).map(([src, color]) => (
            <span key={src} style={{ fontSize: "10px", color, letterSpacing: "0.08em", textTransform: "uppercase" }}>
              ● {src}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}