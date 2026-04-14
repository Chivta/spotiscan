import * as React from "react";
import { useState, useEffect } from "react";
import type { ArtistInsertSuggestion } from "../types/models";
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

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export default function ArtistSuggestions() {
  const { lang } = useLanguage();
  const tx = translations[lang];

  const [suggestions, setSuggestions] = useState<ArtistInsertSuggestion[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Create form
  const [createName, setCreateName] = useState("");
  const [createDesc, setCreateDesc] = useState("");
  const [createError, setCreateError] = useState<string | null>(null);
  const [createSuccess, setCreateSuccess] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // Edit state
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editDesc, setEditDesc] = useState("");
  const [editError, setEditError] = useState<string | null>(null);
  const [editSaving, setEditSaving] = useState(false);

  // Delete state
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [deleteError, setDeleteError] = useState<{ id: number; msg: string } | null>(null);

  const loadSuggestions = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("/api/suggestions/artist-insert", { credentials: "include" });
      if (!res.ok) throw new Error(tx.somethingWentWrong);
      const data: ArtistInsertSuggestion[] = await res.json();
      setSuggestions(data ?? []);
    } catch {
      setError(tx.somethingWentWrong);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { loadSuggestions(); }, []);

  const resolveErrorMessage = (code: string | undefined, fallback: string): string => {
    if (code === "ARTIST_EXISTS") return tx.artistExists;
    if (code === "SUGGESTION_APPROVED") return tx.suggestionApproved;
    return fallback;
  };

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
      setCreateError(resolveErrorMessage(e.code, e.message || tx.somethingWentWrong));
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

  const cancelEdit = () => {
    setEditingId(null);
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
      setDeleteError({ id, msg: resolveErrorMessage(e.code, e.message || tx.somethingWentWrong) });
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
      {/* Create form */}
      <div style={cardStyle}>
        <h2 style={{ margin: "0 0 6px", color: "#fff", fontSize: "1.1rem", fontWeight: 700 }}>
          {tx.suggestArtist}
        </h2>
        <p style={{ margin: "0 0 20px", color: "rgba(255,255,255,0.5)", fontSize: 13 }}>
          {tx.suggestionsHint}
        </p>
        <form onSubmit={handleCreate} style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          <div>
            <label style={{ display: "block", color: "rgba(255,255,255,0.6)", fontSize: 12, marginBottom: 6 }}>
              {tx.artistNameLabel}
            </label>
            <input
              style={inputStyle}
              placeholder={tx.artistNamePlaceholder}
              value={createName}
              onChange={e => setCreateName(e.target.value)}
              maxLength={200}
            />
          </div>
          <div>
            <label style={{ display: "block", color: "rgba(255,255,255,0.6)", fontSize: 12, marginBottom: 6 }}>
              {tx.descriptionLabel}
            </label>
            <textarea
              style={{ ...inputStyle, minHeight: 80, resize: "vertical" }}
              placeholder={tx.descriptionPlaceholder}
              value={createDesc}
              onChange={e => setCreateDesc(e.target.value)}
              maxLength={1000}
            />
            <div style={{ textAlign: "right", fontSize: 11, color: "rgba(255,255,255,0.3)", marginTop: 4 }}>
              {createDesc.length}/1000
            </div>
          </div>
          {createError && (
            <div style={{ color: "#ff7070", fontSize: 13 }}>{createError}</div>
          )}
          {createSuccess && (
            <div style={{ color: "#1DB954", fontSize: 13 }}>{tx.suggestionCreated}</div>
          )}
          <div>
            <button
              type="submit"
              style={{ ...primaryButtonStyle, opacity: submitting || !createName.trim() || !createDesc.trim() ? 0.5 : 1 }}
              disabled={submitting || !createName.trim() || !createDesc.trim()}
            >
              {submitting ? tx.submitting : tx.submitSuggestion}
            </button>
          </div>
        </form>
      </div>

      {/* List */}
      <div style={cardStyle}>
        <h2 style={{ margin: "0 0 16px", color: "#fff", fontSize: "1.1rem", fontWeight: 700 }}>
          {tx.yourSuggestions}
        </h2>

        {loading && (
          <div style={{ color: "rgba(255,255,255,0.4)", fontSize: 14 }}>...</div>
        )}
        {error && (
          <div style={{ color: "#ff7070", fontSize: 14 }}>{error}</div>
        )}
        {!loading && !error && suggestions.length === 0 && (
          <div style={{ color: "rgba(255,255,255,0.4)", fontSize: 14 }}>{tx.noSuggestions}</div>
        )}

        {!loading && suggestions.length > 0 && (
          <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            {suggestions.map(s => (
              <div
                key={s.ID}
                style={{
                  background: "rgba(255,255,255,0.04)",
                  border: "1px solid rgba(255,255,255,0.08)",
                  borderRadius: 12,
                  padding: "16px 20px",
                }}
              >
                {editingId === s.ID ? (
                  <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                    <input
                      style={inputStyle}
                      value={editName}
                      onChange={e => setEditName(e.target.value)}
                      maxLength={200}
                    />
                    <textarea
                      style={{ ...inputStyle, minHeight: 70, resize: "vertical" }}
                      value={editDesc}
                      onChange={e => setEditDesc(e.target.value)}
                      maxLength={1000}
                    />
                    <div style={{ textAlign: "right", fontSize: 11, color: "rgba(255,255,255,0.3)" }}>
                      {editDesc.length}/1000
                    </div>
                    {editError && <div style={{ color: "#ff7070", fontSize: 13 }}>{editError}</div>}
                    <div style={{ display: "flex", gap: 8 }}>
                      <button
                        style={{ ...primaryButtonStyle, opacity: editSaving ? 0.5 : 1 }}
                        disabled={editSaving}
                        onClick={() => handleUpdate(s.ID)}
                      >
                        {tx.saveChanges}
                      </button>
                      <button style={ghostButtonStyle} onClick={cancelEdit}>
                        {tx.cancelEdit}
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12 }}>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap", marginBottom: 6 }}>
                          <span style={{ color: "#fff", fontWeight: 600, fontSize: 15 }}>{s.ArtistName}</span>
                          <span style={{
                            fontSize: 11,
                            padding: "2px 8px",
                            borderRadius: 50,
                            background: s.Approved ? "rgba(29,185,84,0.15)" : "rgba(255,255,255,0.08)",
                            color: s.Approved ? "#1DB954" : "rgba(255,255,255,0.4)",
                            border: `1px solid ${s.Approved ? "rgba(29,185,84,0.3)" : "rgba(255,255,255,0.1)"}`,
                          }}>
                            {s.Approved ? tx.approved : tx.pending}
                          </span>
                        </div>
                        <p style={{ margin: "0 0 8px", color: "rgba(255,255,255,0.6)", fontSize: 13, lineHeight: 1.5 }}>
                          {s.Description}
                        </p>
                        <div style={{ color: "rgba(255,255,255,0.25)", fontSize: 11 }}>
                          {formatDate(s.CreatedAt)}
                        </div>
                      </div>
                      {!s.Approved && (
                        <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: 6, flexShrink: 0 }}>
                          <div style={{ display: "flex", gap: 6 }}>
                            <button style={ghostButtonStyle} onClick={() => startEdit(s)}>
                              {tx.editSuggestion}
                            </button>
                            <button
                              style={{ ...dangerButtonStyle, opacity: deletingId === s.ID ? 0.5 : 1 }}
                              disabled={deletingId === s.ID}
                              onClick={() => {
                                setDeleteError(null);
                                if (window.confirm(tx.confirmDelete)) handleDelete(s.ID);
                              }}
                            >
                              {tx.deleteSuggestion}
                            </button>
                          </div>
                          {deleteError?.id === s.ID && (
                            <div style={{ color: "#ff7070", fontSize: 11, textAlign: "right", maxWidth: 220 }}>
                              {deleteError.msg}
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
