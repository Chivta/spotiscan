import * as React from "react";
import { useState, useEffect } from "react";
import type { ArtistInsertSuggestion, ArtistDeleteSuggestion } from "../types/models";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";

const cardStyle: React.CSSProperties = {
  background: "rgba(255, 255, 255, 0.05)",
  backdropFilter: "blur(10px)",
  border: "1px solid rgba(255, 255, 255, 0.1)",
  borderRadius: 16,
  padding: 24,
};

const inputStyle: React.CSSProperties = {
  width: "100%",
  padding: "12px 16px",
  fontSize: 14,
  background: "rgba(255, 255, 255, 0.1)",
  border: "1px solid rgba(255, 255, 255, 0.2)",
  borderRadius: 8,
  color: "#fff",
  outline: "none",
  boxSizing: "border-box",
};

const primaryButtonStyle: React.CSSProperties = {
  padding: "10px 22px",
  background: "#1DB954",
  color: "#000",
  border: "none",
  borderRadius: 50,
  fontWeight: 600,
  fontSize: 13,
  cursor: "pointer",
};

const ghostButtonStyle: React.CSSProperties = {
  padding: "6px 14px",
  background: "transparent",
  border: "1px solid rgba(255,255,255,0.25)",
  borderRadius: 50,
  color: "rgba(255,255,255,0.7)",
  fontSize: 12,
  cursor: "pointer",
};

const dangerButtonStyle: React.CSSProperties = {
  ...ghostButtonStyle,
  borderColor: "rgba(255, 80, 80, 0.4)",
  color: "rgba(255, 120, 120, 0.9)",
};

type SuggestionState = "pending" | "approved" | "declined";

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
  const style = stateStyles[state];
  return (
    <span style={{
      fontSize: 11,
      padding: "2px 8px",
      borderRadius: 50,
      background: style.bg,
      color: style.color,
      border: `1px solid ${style.border}`,
    }}>
      {tx[state]}
    </span>
  );
}

function DeclineReasonBlock({ reason, label }: { reason: string; label: string }) {
  return (
    <div style={{ marginBottom: 8, padding: "6px 10px", background: "rgba(231,76,60,0.08)", border: "1px solid rgba(231,76,60,0.2)", borderRadius: 6 }}>
      <span style={{ fontSize: 11, color: "rgba(231,76,60,0.7)", fontWeight: 600, textTransform: "uppercase", letterSpacing: "0.05em" }}>{label}: </span>
      <span style={{ fontSize: 12, color: "rgba(255,255,255,0.55)" }}>{reason}</span>
    </div>
  );
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

function resolveErrorMessage(
  code: string | undefined,
  fallback: string,
  tx: { artistExists: string; suggestionNotPending: string; artistNotInDbForDelete: string },
  context?: "delete",
): string {
  if (code === "ARTIST_EXISTS") return tx.artistExists;
  if (code === "SUGGESTION_NOT_PENDING") return tx.suggestionNotPending;
  if (code === "NOT_FOUND" && context === "delete") return tx.artistNotInDbForDelete;
  return fallback;
}

// ─── Insert suggestions ───────────────────────────────────────────────────────

function InsertSuggestions() {
  const { lang } = useLanguage();
  const tx = translations[lang];

  const [suggestions, setSuggestions] = useState<ArtistInsertSuggestion[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const [createName, setCreateName] = useState("");
  const [createDesc, setCreateDesc] = useState("");
  const [createError, setCreateError] = useState<string | null>(null);
  const [createSuccess, setCreateSuccess] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editDesc, setEditDesc] = useState("");
  const [editError, setEditError] = useState<string | null>(null);
  const [editSaving, setEditSaving] = useState(false);

  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [deleteError, setDeleteError] = useState<{ id: number; msg: string } | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const res = await fetch("/api/suggestions/artist-insert", { credentials: "include" });
        if (!res.ok) throw new Error();
        const data: ArtistInsertSuggestion[] = await res.json();
        setSuggestions(data ?? []);
      } catch {
        setLoadError(tx.somethingWentWrong);
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateError(null);
    setCreateSuccess(false);
    if (!createName.trim() || !createDesc.trim()) return;
    setSubmitting(true);
    try {
      const res = await fetch("/api/suggestions/artist-insert", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ArtistName: createName.trim(), Description: createDesc.trim() }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        throw Object.assign(new Error(body?.error || tx.somethingWentWrong), { code: body?.code });
      }
      const created: ArtistInsertSuggestion = await res.json();
      setSuggestions(prev => [created, ...prev]);
      setCreateName("");
      setCreateDesc("");
      setCreateSuccess(true);
      setTimeout(() => setCreateSuccess(false), 3000);
    } catch (err: unknown) {
      const e = err as Error & { code?: string };
      setCreateError(resolveErrorMessage(e.code, e.message || tx.somethingWentWrong, tx));
    } finally {
      setSubmitting(false);
    }
  };

  const startEdit = (s: ArtistInsertSuggestion) => {
    setEditingId(s.ID);
    setEditName(s.ArtistName);
    setEditDesc(s.Description);
    setEditError(null);
  };

  const handleUpdate = async (id: number) => {
    setEditError(null);
    if (!editName.trim() || !editDesc.trim()) return;
    setEditSaving(true);
    try {
      const res = await fetch(`/api/suggestions/artist-insert/${id}`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ArtistName: editName.trim(), Description: editDesc.trim() }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        throw new Error(body?.error || tx.somethingWentWrong);
      }
      const updated: ArtistInsertSuggestion = await res.json();
      setSuggestions(prev => prev.map(s => s.ID === id ? updated : s));
      setEditingId(null);
    } catch (err: unknown) {
      setEditError(err instanceof Error ? err.message : tx.somethingWentWrong);
    } finally {
      setEditSaving(false);
    }
  };

  const handleDelete = async (id: number) => {
    setDeletingId(id);
    setDeleteError(null);
    try {
      const res = await fetch(`/api/suggestions/artist-insert/${id}`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        throw Object.assign(new Error(body?.error || tx.somethingWentWrong), { code: body?.code });
      }
      setSuggestions(prev => prev.filter(s => s.ID !== id));
    } catch (err: unknown) {
      const e = err as Error & { code?: string };
      setDeleteError({ id, msg: resolveErrorMessage(e.code, e.message || tx.somethingWentWrong, tx) });
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
      <div style={cardStyle}>
        <h2 style={{ margin: "0 0 6px", color: "#fff", fontSize: "1.1rem", fontWeight: 700 }}>{tx.suggestArtist}</h2>
        <p style={{ margin: "0 0 20px", color: "rgba(255,255,255,0.5)", fontSize: 13 }}>{tx.suggestionsHint}</p>
        <form onSubmit={handleCreate} style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          <div>
            <label style={{ display: "block", color: "rgba(255,255,255,0.6)", fontSize: 12, marginBottom: 6 }}>{tx.artistNameLabel}</label>
            <input style={inputStyle} placeholder={tx.artistNamePlaceholder} value={createName} onChange={e => setCreateName(e.target.value)} maxLength={200} />
          </div>
          <div>
            <label style={{ display: "block", color: "rgba(255,255,255,0.6)", fontSize: 12, marginBottom: 6 }}>{tx.descriptionLabel}</label>
            <textarea style={{ ...inputStyle, minHeight: 80, resize: "vertical" }} placeholder={tx.descriptionPlaceholder} value={createDesc} onChange={e => setCreateDesc(e.target.value)} maxLength={1000} />
            <div style={{ textAlign: "right", fontSize: 11, color: "rgba(255,255,255,0.3)", marginTop: 4 }}>{createDesc.length}/1000</div>
          </div>
          {createError && <div style={{ color: "#ff7070", fontSize: 13 }}>{createError}</div>}
          {createSuccess && <div style={{ color: "#1DB954", fontSize: 13 }}>{tx.suggestionCreated}</div>}
          <div>
            <button type="submit" style={{ ...primaryButtonStyle, opacity: submitting || !createName.trim() || !createDesc.trim() ? 0.5 : 1 }} disabled={submitting || !createName.trim() || !createDesc.trim()}>
              {submitting ? tx.submitting : tx.submitSuggestion}
            </button>
          </div>
        </form>
      </div>

      <div style={cardStyle}>
        <h2 style={{ margin: "0 0 16px", color: "#fff", fontSize: "1.1rem", fontWeight: 700 }}>{tx.yourSuggestions}</h2>
        {loading && <div style={{ color: "rgba(255,255,255,0.4)", fontSize: 14 }}>...</div>}
        {loadError && <div style={{ color: "#ff7070", fontSize: 14 }}>{loadError}</div>}
        {!loading && !loadError && suggestions.length === 0 && (
          <div style={{ color: "rgba(255,255,255,0.4)", fontSize: 14 }}>{tx.noSuggestions}</div>
        )}
        {!loading && suggestions.length > 0 && (
          <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            {suggestions.map(s => (
              <div key={s.ID} style={{ background: "rgba(255,255,255,0.04)", border: "1px solid rgba(255,255,255,0.08)", borderRadius: 12, padding: "16px 20px" }}>
                {editingId === s.ID ? (
                  <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                    <input style={inputStyle} value={editName} onChange={e => setEditName(e.target.value)} maxLength={200} />
                    <textarea style={{ ...inputStyle, minHeight: 70, resize: "vertical" }} value={editDesc} onChange={e => setEditDesc(e.target.value)} maxLength={1000} />
                    <div style={{ textAlign: "right", fontSize: 11, color: "rgba(255,255,255,0.3)" }}>{editDesc.length}/1000</div>
                    {editError && <div style={{ color: "#ff7070", fontSize: 13 }}>{editError}</div>}
                    <div style={{ display: "flex", gap: 8 }}>
                      <button style={{ ...primaryButtonStyle, opacity: editSaving ? 0.5 : 1 }} disabled={editSaving} onClick={() => handleUpdate(s.ID)}>{tx.saveChanges}</button>
                      <button style={ghostButtonStyle} onClick={() => { setEditingId(null); setEditError(null); }}>{tx.cancelEdit}</button>
                    </div>
                  </div>
                ) : (
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12 }}>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap", marginBottom: 6 }}>
                        <span style={{ color: "#fff", fontWeight: 600, fontSize: 15 }}>{s.ArtistName}</span>
                        <StatusBadge state={toState(s.State)} tx={tx} />
                      </div>
                      <p style={{ margin: "0 0 8px", color: "rgba(255,255,255,0.6)", fontSize: 13, lineHeight: 1.5 }}>{s.Description}</p>
                      {toState(s.State) === "declined" && s.DeclineReason && (
                        <DeclineReasonBlock reason={s.DeclineReason} label={tx.adminDeclineReason} />
                      )}
                      <div style={{ color: "rgba(255,255,255,0.25)", fontSize: 11 }}>{formatDate(s.CreatedAt)}</div>
                    </div>
                    {toState(s.State) === "pending" && (
                      <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: 6, flexShrink: 0 }}>
                        <div style={{ display: "flex", gap: 6 }}>
                          <button style={ghostButtonStyle} onClick={() => startEdit(s)}>{tx.editSuggestion}</button>
                          <button
                            style={{ ...dangerButtonStyle, opacity: deletingId === s.ID ? 0.5 : 1 }}
                            disabled={deletingId === s.ID}
                            onClick={() => { setDeleteError(null); if (window.confirm(tx.confirmDelete)) handleDelete(s.ID); }}
                          >
                            {tx.deleteSuggestion}
                          </button>
                        </div>
                        {deleteError?.id === s.ID && (
                          <div style={{ color: "#ff7070", fontSize: 11, textAlign: "right", maxWidth: 220 }}>{deleteError.msg}</div>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Delete suggestions ───────────────────────────────────────────────────────

interface DeleteSuggestionsProps {
  prefillName?: string;
}

function DeleteSuggestions({ prefillName }: DeleteSuggestionsProps) {
  const { lang } = useLanguage();
  const tx = translations[lang];

  const [suggestions, setSuggestions] = useState<ArtistDeleteSuggestion[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const [createName, setCreateName] = useState(prefillName ?? "");
  const [createDesc, setCreateDesc] = useState("");
  const [createError, setCreateError] = useState<string | null>(null);
  const [createSuccess, setCreateSuccess] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editDesc, setEditDesc] = useState("");
  const [editError, setEditError] = useState<string | null>(null);
  const [editSaving, setEditSaving] = useState(false);

  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [deleteError, setDeleteError] = useState<{ id: number; msg: string } | null>(null);

  // Update create form when prefill changes (e.g. user clicks another artist)
  useEffect(() => {
    if (prefillName !== undefined) {
      setCreateName(prefillName);
      setCreateError(null);
      setCreateSuccess(false);
    }
  }, [prefillName]);

  useEffect(() => {
    (async () => {
      try {
        const res = await fetch("/api/suggestions/artist-delete", { credentials: "include" });
        if (!res.ok) throw new Error();
        const data: ArtistDeleteSuggestion[] = await res.json();
        setSuggestions(data ?? []);
      } catch {
        setLoadError(tx.somethingWentWrong);
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateError(null);
    setCreateSuccess(false);
    if (!createName.trim() || !createDesc.trim()) return;
    setSubmitting(true);
    try {
      const res = await fetch("/api/suggestions/artist-delete", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ArtistName: createName.trim(), Description: createDesc.trim() }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        throw Object.assign(new Error(body?.error || tx.somethingWentWrong), { code: body?.code });
      }
      const created: ArtistDeleteSuggestion = await res.json();
      setSuggestions(prev => [created, ...prev]);
      setCreateName("");
      setCreateDesc("");
      setCreateSuccess(true);
      setTimeout(() => setCreateSuccess(false), 3000);
    } catch (err: unknown) {
      const e = err as Error & { code?: string };
      setCreateError(resolveErrorMessage(e.code, e.message || tx.somethingWentWrong, tx, "delete"));
    } finally {
      setSubmitting(false);
    }
  };

  const startEdit = (s: ArtistDeleteSuggestion) => {
    setEditingId(s.ID);
    setEditName(s.ArtistName);
    setEditDesc(s.Description);
    setEditError(null);
  };

  const handleUpdate = async (id: number) => {
    setEditError(null);
    if (!editName.trim() || !editDesc.trim()) return;
    setEditSaving(true);
    try {
      const res = await fetch(`/api/suggestions/artist-delete/${id}`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ArtistName: editName.trim(), Description: editDesc.trim() }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        throw new Error(body?.error || tx.somethingWentWrong);
      }
      const updated: ArtistDeleteSuggestion = await res.json();
      setSuggestions(prev => prev.map(s => s.ID === id ? updated : s));
      setEditingId(null);
    } catch (err: unknown) {
      setEditError(err instanceof Error ? err.message : tx.somethingWentWrong);
    } finally {
      setEditSaving(false);
    }
  };

  const handleDelete = async (id: number) => {
    setDeletingId(id);
    setDeleteError(null);
    try {
      const res = await fetch(`/api/suggestions/artist-delete/${id}`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        throw Object.assign(new Error(body?.error || tx.somethingWentWrong), { code: body?.code });
      }
      setSuggestions(prev => prev.filter(s => s.ID !== id));
    } catch (err: unknown) {
      const e = err as Error & { code?: string };
      setDeleteError({ id, msg: resolveErrorMessage(e.code, e.message || tx.somethingWentWrong, tx) });
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
      <div style={cardStyle}>
        <h2 style={{ margin: "0 0 6px", color: "#fff", fontSize: "1.1rem", fontWeight: 700 }}>{tx.suggestDeleteArtist}</h2>
        <p style={{ margin: "0 0 20px", color: "rgba(255,255,255,0.5)", fontSize: 13 }}>{tx.suggestDeleteHint}</p>
        <form onSubmit={handleCreate} style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          <div>
            <label style={{ display: "block", color: "rgba(255,255,255,0.6)", fontSize: 12, marginBottom: 6 }}>{tx.artistNameLabel}</label>
            <input style={inputStyle} placeholder={tx.artistNamePlaceholder} value={createName} onChange={e => setCreateName(e.target.value)} maxLength={200} />
          </div>
          <div>
            <label style={{ display: "block", color: "rgba(255,255,255,0.6)", fontSize: 12, marginBottom: 6 }}>{tx.descriptionLabel}</label>
            <textarea style={{ ...inputStyle, minHeight: 80, resize: "vertical" }} placeholder={tx.descriptionPlaceholder} value={createDesc} onChange={e => setCreateDesc(e.target.value)} maxLength={1000} />
            <div style={{ textAlign: "right", fontSize: 11, color: "rgba(255,255,255,0.3)", marginTop: 4 }}>{createDesc.length}/1000</div>
          </div>
          {createError && <div style={{ color: "#ff7070", fontSize: 13 }}>{createError}</div>}
          {createSuccess && <div style={{ color: "#1DB954", fontSize: 13 }}>{tx.suggestionCreated}</div>}
          <div>
            <button type="submit" style={{ ...primaryButtonStyle, opacity: submitting || !createName.trim() || !createDesc.trim() ? 0.5 : 1 }} disabled={submitting || !createName.trim() || !createDesc.trim()}>
              {submitting ? tx.submitting : tx.submitSuggestion}
            </button>
          </div>
        </form>
      </div>

      <div style={cardStyle}>
        <h2 style={{ margin: "0 0 16px", color: "#fff", fontSize: "1.1rem", fontWeight: 700 }}>{tx.yourDeleteSuggestions}</h2>
        {loading && <div style={{ color: "rgba(255,255,255,0.4)", fontSize: 14 }}>...</div>}
        {loadError && <div style={{ color: "#ff7070", fontSize: 14 }}>{loadError}</div>}
        {!loading && !loadError && suggestions.length === 0 && (
          <div style={{ color: "rgba(255,255,255,0.4)", fontSize: 14 }}>{tx.noDeleteSuggestions}</div>
        )}
        {!loading && suggestions.length > 0 && (
          <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            {suggestions.map(s => (
              <div key={s.ID} style={{ background: "rgba(255,255,255,0.04)", border: "1px solid rgba(255,255,255,0.08)", borderRadius: 12, padding: "16px 20px" }}>
                {editingId === s.ID ? (
                  <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                    <input style={inputStyle} value={editName} onChange={e => setEditName(e.target.value)} maxLength={200} placeholder={tx.artistNamePlaceholder} />
                    <textarea style={{ ...inputStyle, minHeight: 70, resize: "vertical" }} value={editDesc} onChange={e => setEditDesc(e.target.value)} maxLength={1000} />
                    <div style={{ textAlign: "right", fontSize: 11, color: "rgba(255,255,255,0.3)" }}>{editDesc.length}/1000</div>
                    {editError && <div style={{ color: "#ff7070", fontSize: 13 }}>{editError}</div>}
                    <div style={{ display: "flex", gap: 8 }}>
                      <button style={{ ...primaryButtonStyle, opacity: editSaving ? 0.5 : 1 }} disabled={editSaving} onClick={() => handleUpdate(s.ID)}>{tx.saveChanges}</button>
                      <button style={ghostButtonStyle} onClick={() => { setEditingId(null); setEditError(null); }}>{tx.cancelEdit}</button>
                    </div>
                  </div>
                ) : (
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12 }}>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap", marginBottom: 6 }}>
                        <span style={{ color: "#fff", fontWeight: 600, fontSize: 15 }}>{s.ArtistName}</span>
                        <StatusBadge state={toState(s.State)} tx={tx} />
                      </div>
                      <p style={{ margin: "0 0 8px", color: "rgba(255,255,255,0.6)", fontSize: 13, lineHeight: 1.5 }}>{s.Description}</p>
                      {toState(s.State) === "declined" && s.DeclineReason && (
                        <DeclineReasonBlock reason={s.DeclineReason} label={tx.adminDeclineReason} />
                      )}
                      <div style={{ color: "rgba(255,255,255,0.25)", fontSize: 11 }}>{formatDate(s.CreatedAt)}</div>
                    </div>
                    {toState(s.State) === "pending" && (
                      <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: 6, flexShrink: 0 }}>
                        <div style={{ display: "flex", gap: 6 }}>
                          <button style={ghostButtonStyle} onClick={() => startEdit(s)}>{tx.editSuggestion}</button>
                          <button
                            style={{ ...dangerButtonStyle, opacity: deletingId === s.ID ? 0.5 : 1 }}
                            disabled={deletingId === s.ID}
                            onClick={() => { setDeleteError(null); if (window.confirm(tx.confirmDelete)) handleDelete(s.ID); }}
                          >
                            {tx.deleteSuggestion}
                          </button>
                        </div>
                        {deleteError?.id === s.ID && (
                          <div style={{ color: "#ff7070", fontSize: 11, textAlign: "right", maxWidth: 220 }}>{deleteError.msg}</div>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Tab container ────────────────────────────────────────────────────────────

type SuggestionTab = "insert" | "delete";

interface ArtistSuggestionsProps {
  initialTab?: SuggestionTab;
  deletePrefillName?: string;
}

export default function ArtistSuggestions({ initialTab = "insert", deletePrefillName }: ArtistSuggestionsProps) {
  const { lang } = useLanguage();
  const tx = translations[lang];
  const [tab, setTab] = useState<SuggestionTab>(initialTab);

  // When a new prefill arrives (from scanner), always switch to delete tab
  useEffect(() => {
    if (deletePrefillName !== undefined) setTab("delete");
  }, [deletePrefillName]);

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
      <div style={{
        display: "flex",
        gap: 4,
        background: "rgba(255,255,255,0.05)",
        border: "1px solid rgba(255,255,255,0.1)",
        borderRadius: 50,
        padding: 4,
        width: "fit-content",
      }}>
        {(["insert", "delete"] as SuggestionTab[]).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            style={{
              padding: "7px 18px",
              borderRadius: 50,
              border: "none",
              fontSize: 12,
              fontWeight: 600,
              cursor: "pointer",
              transition: "all 0.15s ease",
              background: tab === t ? "rgba(255,255,255,0.12)" : "transparent",
              color: tab === t ? "#fff" : "rgba(255,255,255,0.45)",
            }}
          >
            {t === "insert" ? tx.suggestArtist : tx.suggestDeleteArtist}
          </button>
        ))}
      </div>

      {tab === "insert" && <InsertSuggestions />}
      {tab === "delete" && <DeleteSuggestions prefillName={deletePrefillName} />}
    </div>
  );
}
