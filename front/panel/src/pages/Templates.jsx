import { useEffect, useState } from "react";
import { useTranslation } from "../lib/i18n";
import { API_BASE, fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const emptyForm = {
  name: "",
  image_url: "",
  is_active: true
};

export default function Templates() {
  const { t } = useTranslation();
  const [templates, setTemplates] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadTemplates = async () => {
    try {
      const response = await fetchJSON("/api/admin/templates");
      setTemplates(response.data || []);
      setError("");
    } catch (err) {
      setTemplates([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTemplates();
  }, []);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError("");
    try {
      if (editingId) {
        await fetchJSON(`/api/admin/templates/${editingId}`, {
          method: "PUT",
          body: JSON.stringify(form)
        });
      } else {
        await fetchJSON("/api/admin/templates", {
          method: "POST",
          body: JSON.stringify(form)
        });
      }
      setForm(emptyForm);
      setEditingId(null);
      loadTemplates();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  const handleEdit = (template) => {
    setEditingId(template.id);
    setForm({
      name: template.name || "",
      image_url: template.image_url || "",
      is_active: template.is_active
    });
  };

  const handleDelete = async (id) => {
    try {
      await fetchJSON(`/api/admin/templates/${id}`, { method: "DELETE" });
      loadTemplates();
    } catch (err) {
      setError(t("messages.error"));
    }
  };

  return (
    <div className="grid gap-6 lg:grid-cols-[1.1fr_1.9fr]">
      <form className="panel-card space-y-4" onSubmit={handleSubmit}>
        <h2 className="font-display text-xl">{t("templates.title")}</h2>

        <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
          {t("form.templateName")}
          <input
            type="text"
            className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
            value={form.name}
            onChange={(event) => setForm({ ...form, name: event.target.value })}
            required
          />
        </label>
        <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
          {t("form.templateImageUrl")}
          <div className="mt-2 flex items-center gap-2">
            <input
              type="file"
              accept="image/*"
              className="w-full text-sm text-primary/70 file:mr-4 file:rounded-full file:border-0 file:bg-primary/10 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20"
              onChange={async (e) => {
                const file = e.target.files?.[0];
                if (!file) return;

                const formData = new FormData();
                formData.append("file", file);

                try {
                  setLoading(true);
                  const res = await fetch(`${API_BASE}/api/admin/upload/template`, {
                    method: "POST",
                    body: formData,
                    credentials: "include",
                  });
                  if (!res.ok) throw new Error("Upload failed");
                  const data = await res.json();
                  setForm({ ...form, image_url: data?.data?.image_url || "" });
                } catch (err) {
                  setError(t("messages.error"));
                } finally {
                  setLoading(false);
                }
              }}
            />
            {form.image_url && (
              <div className="h-10 w-10 overflow-hidden rounded-lg border border-primary/20 bg-primary/5">
                <img src={resolveImageUrl(form.image_url)} alt="Preview" className="h-full w-full object-cover" />
              </div>
            )}
          </div>
        </label>
        <label className="flex items-center gap-3 text-xs font-semibold uppercase tracking-wide text-primary/70">
          <input
            type="checkbox"
            checked={form.is_active}
            onChange={(event) => setForm({ ...form, is_active: event.target.checked })}
            className="h-4 w-4 rounded border-primary/20"
          />
          {t("form.active")}
        </label>

        {error && <p className="text-sm text-red-500">{error}</p>}

        <div className="flex items-center gap-3">
          <button
            type="submit"
            className="rounded-full bg-primary px-5 py-2 text-xs font-semibold text-sand"
          >
            {editingId ? t("actions.update") : t("actions.create")}
          </button>
          {editingId && (
            <button
              type="button"
              className="rounded-full border border-primary/20 px-5 py-2 text-xs font-semibold text-primary/70"
              onClick={() => {
                setEditingId(null);
                setForm(emptyForm);
              }}
            >
              {t("actions.cancel")}
            </button>
          )}
        </div>
      </form>

      <div className="panel-card">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="font-display text-xl">{t("templates.title")}</h3>
        </div>

        {loading ? (
          <p className="text-sm text-primary/70">{t("messages.loading")}</p>
        ) : templates.length === 0 ? (
          <p className="text-sm text-primary/70">{t("templates.empty")}</p>
        ) : (
          <div className="space-y-3">
            {templates.map((template) => (
              <div
                key={template.id}
                className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-primary/10 bg-white/80 px-4 py-3"
              >
                <div className="flex items-center gap-4">
                  <div className="h-16 w-16 overflow-hidden rounded-xl border border-primary/10 bg-primary/5">
                    {template.image_url && (
                      <img
                        src={resolveImageUrl(template.image_url)}
                        alt={template.name}
                        className="h-full w-full object-contain"
                      />
                    )}
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-primary">{template.name}</p>
                    <p className="text-xs text-primary/40">{template.image_url}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {template.is_active ? (
                    <span className="rounded-full bg-accent/20 px-3 py-1 text-xs font-semibold text-accent">
                      {t("form.active")}
                    </span>
                  ) : (
                    <span className="rounded-full bg-primary/10 px-3 py-1 text-xs font-semibold text-primary/60">
                      {t("form.inactive")}
                    </span>
                  )}
                  <button
                    type="button"
                    onClick={() => handleEdit(template)}
                    className="rounded-full border border-primary/20 px-3 py-1 text-xs font-semibold text-primary/70"
                  >
                    {t("actions.edit")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(template.id)}
                    className="rounded-full border border-red-200 px-3 py-1 text-xs font-semibold text-red-500"
                  >
                    {t("actions.delete")}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
