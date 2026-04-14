import * as React from "react";
import { useState, useEffect } from "react";
import Aurora from "../components/react-bits/Aurora";
import type { ArtistInsertSuggestion, ArtistDeleteSuggestion } from "../types/models";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";

// ─── Shared helpers ───────────────────────────────────────────────────────────

type SuggestionState = "pending" | "approved" | "declined";
type SuggestionType = "insert" | "delete";
type FilterState = "all" | SuggestionState;

const VALID_STATES = new Set<SuggestionState>(["pending", "approved", "declined"]);
function toState(raw: string): SuggestionState {
  return VALID_STATES.has(raw as SuggestionState) ? (raw as SuggestionState) : "pending";
}

const stateStyles: Record<SuggestionState, { bg: string; color: string; border: string }> = {
  pending:  { bg: "rgba(255,255,255,0.08)", color: "rgba(255,255,255,0.4)", border: "rgba(255,255,255,0.1)" },
  approved: { bg: "rgba(29,185,84,0.15)",   color: "#1DB954",               border: "rgba(29,185,84,0.3)"   },
  declined: { bg: "rgba(231,76,60,0.12)",   color: "#e74c3c",               border: "rgba(231,76,60,0.3)"   },
};

function StatusBadge({ state, tx }: { state: SuggestionState; tx: { approved: string; pending: string; declined: string } }) {
  const s = stateStyles[state];
  return (
    <span style={{ fontSize: 11, padding: "2px 8px", borderRadius: 50, background: s.bg, color: s.color, border: `1px solid ${s.border}` }}>
      {tx[state]}
    </span>
  );
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

const cardStyle: React.CSSProperties = {
  background: "rgba(255, 255, 255, 0.05)",
  backdropFilter: "blur(10px)",
  border: "1px solid rgba(255, 255, 255, 0.1)",
  borderRadius: 16,
  padding: 24,
};

const inputStyle: React.CSSProperties = {
  width: "100%",
  padding: "10px 14px",
  fontSize: 13,
  background: "rgba(255, 255, 255, 0.1)",
  border: "1px solid rgba(255, 255, 255, 0.2)",
  borderRadius: 8,
  color: "#fff",
  outline: "none",
  boxSizing: "border-box",
  resize: "vertical" as const,
};

const pillTabStyle = (active: boolean): React.CSSProperties => ({
  padding: "6px 16px",
  borderRadius: 50,
  border: "none",
  fontSize: 12,
  fontWeight: 600,
  cursor: "pointer",
  background: active ? "rgba(255,255,255,0.12)" : "transparent",
  color: active ? "#fff" : "rgba(255,255,255,0.45)",
  transition: "all 0.15s ease",
});

// ─── Suggestion card ──────────────────────────────────────────────────────────

type AnySuggestion = (ArtistInsertSuggestion | ArtistDeleteSuggestion) & { ArtistName: string };

interface SuggestionCardProps {
  suggestion: AnySuggestion;
  type: SuggestionType;
  onApprove: (id: number) => Promise<void>;
  onDecline: (id: number, reason: string) => Promise<void>;
}

function SuggestionCard({ suggestion: s, type, onApprove, onDecline }: SuggestionCardProps) {
  const { lang } = useLanguage();
  const tx = translations[lang];
  const state = toState(s.State);

  const [actioning, setActioning] = useState<"approve" | "decline" | null>(null);
  const [showDeclineForm, setShowDeclineForm] = useState(false);
  const [reason, setReason] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handleApprove = async () => {
    setError(null);
    setActioning("approve");
    try {
      await onApprove(s.ID);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : tx.somethingWentWrong);
    } finally {
      setActioning(null);
    }
  };

  const handleDecline = async () => {
    if (!reason.trim()) return;
    setError(null);
    setActioning("decline");
    try {
      await onDecline(s.ID, reason.trim());
      setShowDeclineForm(false);
      setReason("");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : tx.somethingWentWrong);
    } finally {
      setActioning(null);
    }
  };

  return (
    <div style={{
      background: "rgba(255,255,255,0.04)",
      border: "1px solid rgba(255,255,255,0.08)",
      borderRadius: 12,
      padding: "16px 20px",
      display: "flex",
      flexDirection: "column",
      gap: 10,
    }}>
      {/* Header row */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12 }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap", marginBottom: 4 }}>
            <span style={{ color: "#fff", fontWeight: 600, fontSize: 15 }}>{s.ArtistName}</span>
            <StatusBadge state={state} tx={tx} />
            <span style={{ fontSize: 11, color: "rgba(255,255,255,0.3)" }}>
              {tx.adminCreatorId} #{s.CreatorID}
            </span>
          </div>
          <p style={{ margin: 0, color: "rgba(255,255,255,0.6)", fontSize: 13, lineHeight: 1.5 }}>{s.Description}</p>
          {state === "declined" && s.DeclineReason && (
            <div style={{ marginTop: 6, padding: "8px 12px", background: "rgba(231,76,60,0.08)", border: "1px solid rgba(231,76,60,0.2)", borderRadius: 8 }}>
              <span style={{ fontSize: 11, color: "rgba(231,76,60,0.7)", fontWeight: 600, textTransform: "uppercase", letterSpacing: "0.05em" }}>{tx.adminDeclineReason}: </span>
              <span style={{ fontSize: 12, color: "rgba(255,255,255,0.55)" }}>{s.DeclineReason}</span>
            </div>
          )}
        </div>
        <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: 4, flexShrink: 0 }}>
          <span style={{ fontSize: 11, color: "rgba(255,255,255,0.25)" }}>{formatDate(s.CreatedAt)}</span>
          {state === "pending" && (
            <div style={{ display: "flex", gap: 6, marginTop: 4 }}>
              <button
                onClick={handleApprove}
                disabled={actioning !== null}
                style={{
                  padding: "6px 14px",
                  background: "rgba(29,185,84,0.2)",
                  border: "1px solid rgba(29,185,84,0.4)",
                  borderRadius: 50,
                  color: "#1DB954",
                  fontSize: 12,
                  fontWeight: 600,
                  cursor: actioning !== null ? "not-allowed" : "pointer",
                  opacity: actioning !== null ? 0.5 : 1,
                  transition: "all 0.15s ease",
                }}
                onMouseEnter={e => { if (!actioning) e.currentTarget.style.background = "rgba(29,185,84,0.32)"; }}
                onMouseLeave={e => { e.currentTarget.style.background = "rgba(29,185,84,0.2)"; }}
              >
                {actioning === "approve" ? tx.adminApproving : tx.adminApprove}
              </button>
              <button
                onClick={() => { setShowDeclineForm(v => !v); setError(null); }}
                disabled={actioning !== null}
                style={{
                  padding: "6px 14px",
                  background: showDeclineForm ? "rgba(231,76,60,0.2)" : "rgba(255,255,255,0.06)",
                  border: `1px solid ${showDeclineForm ? "rgba(231,76,60,0.4)" : "rgba(255,255,255,0.15)"}`,
                  borderRadius: 50,
                  color: showDeclineForm ? "#e74c3c" : "rgba(255,255,255,0.6)",
                  fontSize: 12,
                  fontWeight: 600,
                  cursor: actioning !== null ? "not-allowed" : "pointer",
                  opacity: actioning !== null ? 0.5 : 1,
                  transition: "all 0.15s ease",
                }}
              >
                {tx.adminDecline}
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Inline decline form */}
      {showDeclineForm && state === "pending" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 8, paddingTop: 4, borderTop: "1px solid rgba(255,255,255,0.07)" }}>
          <label style={{ fontSize: 12, color: "rgba(255,255,255,0.5)" }}>{tx.adminDeclineReason}</label>
          <textarea
            style={{ ...inputStyle, minHeight: 64 }}
            placeholder={tx.adminDeclineReasonPlaceholder}
            value={reason}
            onChange={e => setReason(e.target.value)}
            maxLength={1000}
          />
          <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
            <div style={{ fontSize: 11, color: "rgba(255,255,255,0.25)" }}>{reason.length}/1000</div>
            <div style={{ display: "flex", gap: 6 }}>
              <button
                onClick={() => { setShowDeclineForm(false); setReason(""); setError(null); }}
                style={{ padding: "6px 14px", background: "transparent", border: "1px solid rgba(255,255,255,0.2)", borderRadius: 50, color: "rgba(255,255,255,0.6)", fontSize: 12, cursor: "pointer" }}
              >
                {tx.cancelEdit}
              </button>
              <button
                onClick={handleDecline}
                disabled={!reason.trim() || actioning !== null}
                style={{
                  padding: "6px 14px",
                  background: "rgba(231,76,60,0.2)",
                  border: "1px solid rgba(231,76,60,0.4)",
                  borderRadius: 50,
                  color: "#e74c3c",
                  fontSize: 12,
                  fontWeight: 600,
                  cursor: !reason.trim() || actioning !== null ? "not-allowed" : "pointer",
                  opacity: !reason.trim() || actioning !== null ? 0.5 : 1,
                }}
              >
                {actioning === "decline" ? tx.adminDeclining : tx.adminSubmitDecline}
              </button>
            </div>
          </div>
        </div>
      )}

      {error && <div style={{ fontSize: 12, color: "#ff7070" }}>{error}</div>}
    </div>
  );
}

// ─── Suggestion list panel ────────────────────────────────────────────────────

interface SuggestionPanelProps {
  type: SuggestionType;
}

function SuggestionPanel({ type }: SuggestionPanelProps) {
  const { lang } = useLanguage();
  const tx = translations[lang];

  const endpoint = type === "insert"
    ? "/api/admin/suggestions/artist-insert"
    : "/api/admin/suggestions/artist-delete";

  type S = AnySuggestion;
  const [items, setItems] = useState<S[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterState>("pending");

  useEffect(() => {
    (async () => {
      setLoading(true);
      setLoadError(null);
      try {
        const res = await fetch(endpoint, { credentials: "include" });
        if (!res.ok) throw new Error();
        const data: S[] = await res.json();
        setItems(data ?? []);
      } catch {
        setLoadError(tx.somethingWentWrong);
      } finally {
        setLoading(false);
      }
    })();
  }, [type]);

  const counts: Record<FilterState, number> = {
    all: items.length,
    pending: items.filter(i => toState(i.State) === "pending").length,
    approved: items.filter(i => toState(i.State) === "approved").length,
    declined: items.filter(i => toState(i.State) === "declined").length,
  };

  const visible = filter === "all" ? items : items.filter(i => toState(i.State) === filter);

  const handleApprove = async (id: number) => {
    const res = await fetch(`${endpoint}/${id}/approve`, { method: "POST", credentials: "include" });
    if (!res.ok) {
      const body = await res.json().catch(() => null);
      throw new Error(body?.error || tx.somethingWentWrong);
    }
    const updated: S = await res.json();
    setItems(prev => prev.map(i => i.ID === id ? updated : i));
  };

  const handleDecline = async (id: number, reason: string) => {
    const res = await fetch(`${endpoint}/${id}/decline`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ DeclineReason: reason }),
    });
    if (!res.ok) {
      const body = await res.json().catch(() => null);
      throw new Error(body?.error || tx.somethingWentWrong);
    }
    const updated: S = await res.json();
    setItems(prev => prev.map(i => i.ID === id ? updated : i));
  };

  const filters: { key: FilterState; label: string }[] = [
    { key: "pending",  label: `${tx.adminFilterPending} (${counts.pending})`  },
    { key: "approved", label: `${tx.adminFilterApproved} (${counts.approved})` },
    { key: "declined", label: `${tx.adminFilterDeclined} (${counts.declined})` },
    { key: "all",      label: `${tx.adminFilterAll} (${counts.all})`           },
  ];

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      {/* Filter tabs */}
      <div style={{
        display: "flex",
        gap: 4,
        background: "rgba(255,255,255,0.05)",
        border: "1px solid rgba(255,255,255,0.08)",
        borderRadius: 50,
        padding: 4,
        width: "fit-content",
        flexWrap: "wrap",
      }}>
        {filters.map(f => (
          <button key={f.key} onClick={() => setFilter(f.key)} style={pillTabStyle(filter === f.key)}>
            {f.label}
          </button>
        ))}
      </div>

      {loading && <div style={{ color: "rgba(255,255,255,0.4)", fontSize: 14 }}>...</div>}
      {loadError && <div style={{ color: "#ff7070", fontSize: 14 }}>{loadError}</div>}
      {!loading && !loadError && visible.length === 0 && (
        <div style={{ color: "rgba(255,255,255,0.4)", fontSize: 14 }}>{tx.adminNoSuggestions}</div>
      )}
      {!loading && visible.length > 0 && (
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {visible.map(s => (
            <SuggestionCard
              key={s.ID}
              suggestion={s}
              type={type}
              onApprove={handleApprove}
              onDecline={handleDecline}
            />
          ))}
        </div>
      )}
    </div>
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function AdminPage() {
  const { lang } = useLanguage();
  const tx = translations[lang];
  const [tab, setTab] = useState<SuggestionType>("insert");

  return (
    <div style={{ position: "relative", minHeight: "100vh", width: "100%", overflow: "hidden" }}>
      <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, zIndex: 0, pointerEvents: "none" }}>
        <Aurora colorStops={["#0D4F1C", "#1DB954", "#90EE90"]} blend={0.5} amplitude={1.0} speed={0.5} />
      </div>
      <div style={{ position: "relative", zIndex: 1, width: "100%", padding: "88px 24px 40px", boxSizing: "border-box" }}>
        <div style={{ maxWidth: 860, margin: "0 auto", display: "flex", flexDirection: "column", gap: 24 }}>
          {/* Type tabs */}
          <div style={cardStyle}>
            <div style={{ display: "flex", gap: 4, background: "rgba(255,255,255,0.05)", border: "1px solid rgba(255,255,255,0.1)", borderRadius: 50, padding: 4, width: "fit-content", marginBottom: 24 }}>
              {(["insert", "delete"] as SuggestionType[]).map(t => (
                <button
                  key={t}
                  onClick={() => setTab(t)}
                  style={{
                    padding: "8px 20px",
                    borderRadius: 50,
                    border: "none",
                    fontSize: 13,
                    fontWeight: 600,
                    cursor: "pointer",
                    transition: "all 0.15s ease",
                    background: tab === t ? "#1DB954" : "transparent",
                    color: tab === t ? "#000" : "rgba(255,255,255,0.55)",
                  }}
                >
                  {t === "insert" ? tx.adminInsertSuggestions : tx.adminDeleteSuggestions}
                </button>
              ))}
            </div>
            <SuggestionPanel key={tab} type={tab} />
          </div>
        </div>
      </div>
    </div>
  );
}
