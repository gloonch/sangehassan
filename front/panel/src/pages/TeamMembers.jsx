import { useEffect, useMemo, useState } from "react";
import { ImagePlus, Linkedin, Pencil, Trash2, Upload } from "lucide-react";
import { useTranslation } from "../lib/i18n";
import { API_BASE, fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";

const emptyForm = {
  name_en: "",
  name_fa: "",
  name_ar: "",
  role_en: "",
  role_fa: "",
  role_ar: "",
  bio_en: "",
  bio_fa: "",
  bio_ar: "",
  photo_url: "",
  linkedin_url: "",
  order_index: 0,
  is_active: true
};

const pickLocalized = (item, field, lang) => {
  if (!item) return "";
  return item[`${field}_${lang}`] || item[`${field}_en`] || item[`${field}_fa`] || item[`${field}_ar`] || "";
};

export default function TeamMembers() {
  const { t, lang } = useTranslation();
  const [members, setMembers] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");
  const [formOpen, setFormOpen] = useState(true);

  const loadMembers = async () => {
    try {
      const response = await fetchJSON("/api/admin/team-members");
      setMembers(response.data || []);
      setError("");
    } catch (_) {
      setMembers([]);
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadMembers();
  }, []);

  const previewMember = useMemo(
    () => ({
      name: pickLocalized(form, "name", lang) || t("panelTeam.previewName"),
      role: pickLocalized(form, "role", lang) || t("panelTeam.previewRole"),
      bio: pickLocalized(form, "bio", lang) || t("panelTeam.previewBio"),
      photo: form.photo_url ? resolveImageUrl(form.photo_url) : ""
    }),
    [form, lang, t]
  );

  const handlePhotoUpload = async (file) => {
    if (!file) return;
    setUploading(true);
    setError("");
    try {
      const formData = new FormData();
      formData.append("file", file);
      const response = await fetch(`${API_BASE}/api/admin/upload/team`, {
        method: "POST",
        body: formData,
        credentials: "include"
      });
      if (!response.ok) {
        throw new Error("Upload failed");
      }
      const data = await response.json();
      setForm((prev) => ({ ...prev, photo_url: data?.data?.image_url || "" }));
    } catch (_) {
      setError(t("panelTeam.uploadFailed"));
    } finally {
      setUploading(false);
    }
  };

  const validateForm = () => {
    const hasName = [form.name_en, form.name_fa, form.name_ar].some((value) => value.trim());
    const hasRole = [form.role_en, form.role_fa, form.role_ar].some((value) => value.trim());
    if (!hasName) {
      setError(t("panelTeam.nameRequired"));
      return false;
    }
    if (!hasRole) {
      setError(t("panelTeam.roleRequired"));
      return false;
    }
    return true;
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (!validateForm()) return;

    const payload = {
      ...form,
      order_index: Number(form.order_index) || 0
    };

    setError("");
    try {
      if (editingId) {
        await fetchJSON(`/api/admin/team-members/${editingId}`, {
          method: "PUT",
          body: JSON.stringify(payload)
        });
      } else {
        await fetchJSON("/api/admin/team-members", {
          method: "POST",
          body: JSON.stringify(payload)
        });
      }
      setForm(emptyForm);
      setEditingId(null);
      loadMembers();
    } catch (_) {
      setError(t("messages.error"));
    }
  };

  const handleEdit = async (member) => {
    setEditingId(member.id);
    setLoading(true);
    setError("");
    try {
      const response = await fetchJSON(`/api/admin/team-members/${member.id}`);
      const item = response.data || member;
      setForm({
        name_en: item.name_en || "",
        name_fa: item.name_fa || "",
        name_ar: item.name_ar || "",
        role_en: item.role_en || "",
        role_fa: item.role_fa || "",
        role_ar: item.role_ar || "",
        bio_en: item.bio_en || "",
        bio_fa: item.bio_fa || "",
        bio_ar: item.bio_ar || "",
        photo_url: item.photo_url || "",
        linkedin_url: item.linkedin_url || "",
        order_index: item.order_index || 0,
        is_active: Boolean(item.is_active)
      });
      setFormOpen(true);
    } catch (_) {
      setError(t("messages.error"));
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    try {
      await fetchJSON(`/api/admin/team-members/${id}`, { method: "DELETE" });
      loadMembers();
    } catch (_) {
      setError(t("messages.error"));
    }
  };

  const resetForm = () => {
    setForm(emptyForm);
    setEditingId(null);
    setError("");
  };

  return (
    <div className="space-y-6">
      <section className="panel-card">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="font-display text-xl">{t("panelTeam.title")}</h2>
            <p className="mt-1 text-sm text-primary/55">{t("panelTeam.subtitle")}</p>
          </div>
          <button
            type="button"
            onClick={() => setFormOpen((prev) => !prev)}
            className="rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
          >
            {formOpen ? t("actions.hideForm") : t("actions.showForm")}
          </button>
        </div>

        {formOpen && (
          <form className="mt-6 grid gap-6 lg:grid-cols-[minmax(0,1.45fr)_minmax(20rem,0.75fr)]" onSubmit={handleSubmit}>
            <div className="space-y-5">
              <div className="grid gap-4 md:grid-cols-3">
                {["en", "fa", "ar"].map((locale) => (
                  <label key={`name-${locale}`} className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                    {t(`panelTeam.name${locale.toUpperCase()}`)}
                    <input
                      type="text"
                      className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                      value={form[`name_${locale}`]}
                      onChange={(event) => setForm((prev) => ({ ...prev, [`name_${locale}`]: event.target.value }))}
                    />
                  </label>
                ))}
              </div>

              <div className="grid gap-4 md:grid-cols-3">
                {["en", "fa", "ar"].map((locale) => (
                  <label key={`role-${locale}`} className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                    {t(`panelTeam.role${locale.toUpperCase()}`)}
                    <input
                      type="text"
                      className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                      value={form[`role_${locale}`]}
                      onChange={(event) => setForm((prev) => ({ ...prev, [`role_${locale}`]: event.target.value }))}
                    />
                  </label>
                ))}
              </div>

              <div className="grid gap-4 md:grid-cols-3">
                {["en", "fa", "ar"].map((locale) => (
                  <label key={`bio-${locale}`} className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                    {t(`panelTeam.bio${locale.toUpperCase()}`)}
                    <textarea
                      rows={5}
                      className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                      value={form[`bio_${locale}`]}
                      onChange={(event) => setForm((prev) => ({ ...prev, [`bio_${locale}`]: event.target.value }))}
                    />
                  </label>
                ))}
              </div>

              <div className="grid gap-4 md:grid-cols-[minmax(0,1.2fr)_minmax(0,1.2fr)_minmax(0,0.7fr)]">
                <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("panelTeam.photo")}
                  <input
                    type="file"
                    accept="image/*"
                    className="mt-2 w-full text-sm text-primary/70 file:mr-4 file:rounded-full file:border-0 file:bg-primary/10 file:px-4 file:py-2 file:text-xs file:font-semibold file:text-primary hover:file:bg-primary/20"
                    disabled={uploading}
                    onChange={(event) => handlePhotoUpload(event.target.files?.[0])}
                  />
                  <span className="mt-2 inline-flex items-center gap-2 text-xs text-primary/45">
                    {uploading ? <Upload className="h-3.5 w-3.5 animate-pulse" /> : <ImagePlus className="h-3.5 w-3.5" />}
                    {uploading ? t("panelTeam.uploading") : t("panelTeam.photoHint")}
                  </span>
                </label>

                <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("panelTeam.linkedin")}
                  <input
                    type="url"
                    className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                    value={form.linkedin_url}
                    onChange={(event) => setForm((prev) => ({ ...prev, linkedin_url: event.target.value }))}
                    placeholder="https://linkedin.com/in/..."
                  />
                </label>

                <label className="block text-xs font-semibold uppercase tracking-wide text-primary/70">
                  {t("form.orderIndex")}
                  <input
                    type="number"
                    className="mt-2 w-full rounded-xl border border-primary/20 bg-white px-4 py-3 text-sm"
                    value={form.order_index}
                    onChange={(event) => setForm((prev) => ({ ...prev, order_index: event.target.value }))}
                  />
                </label>
              </div>

              <div className="flex flex-wrap items-center gap-4">
                <label className="flex items-center gap-3 rounded-2xl border border-primary/10 bg-white px-4 py-3 text-sm">
                  <input
                    type="checkbox"
                    className="h-4 w-4"
                    checked={form.is_active}
                    onChange={(event) => setForm((prev) => ({ ...prev, is_active: event.target.checked }))}
                  />
                  <span className="font-semibold text-primary/80">{t("form.isActive")}</span>
                </label>

                {form.photo_url ? (
                  <button
                    type="button"
                    className="inline-flex items-center gap-2 rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
                    onClick={() => setForm((prev) => ({ ...prev, photo_url: "" }))}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    {t("panelTeam.removePhoto")}
                  </button>
                ) : null}
              </div>

              {error && <p className="text-sm text-red-500">{error}</p>}

              <div className="flex flex-wrap gap-3">
                <button
                  type="submit"
                  className="rounded-full bg-primary px-6 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-sand"
                >
                  {editingId ? t("actions.update") : t("actions.create")}
                </button>
                {editingId && (
                  <button
                    type="button"
                    className="rounded-full border border-primary/20 px-6 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-primary"
                    onClick={resetForm}
                  >
                    {t("actions.cancel")}
                  </button>
                )}
              </div>
            </div>

            <aside className="rounded-2xl bg-primary p-4 text-sand shadow-[0_24px_70px_rgba(8,58,79,0.28)]">
              <p className="mb-4 text-xs font-semibold uppercase tracking-[0.24em] text-sand/65">{t("panelTeam.preview")}</p>
              <div className="relative min-h-[28rem] overflow-hidden rounded-lg bg-white/[0.08]">
                {previewMember.photo ? (
                  <img src={previewMember.photo} alt="" className="absolute inset-0 h-full w-full object-cover" />
                ) : (
                  <div className="absolute inset-0 bg-[linear-gradient(145deg,rgba(229,225,221,0.24),rgba(8,58,79,0.2)_36%,rgba(0,0,0,0.42))]" />
                )}
                <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(255,255,255,0.07),rgba(0,0,0,0.08)_42%,rgba(0,0,0,0.72))]" />
                <div className="absolute inset-x-0 bottom-0 flex min-h-[40%] flex-col justify-end bg-[linear-gradient(180deg,rgba(3,24,33,0),rgba(3,24,33,0.72)_18%,rgba(3,24,33,0.95))] px-5 pb-6 pt-16 text-center">
                  <h3 className="font-display text-xl leading-8 text-white">{previewMember.name}</h3>
                  <p className="mt-1 text-[11px] font-semibold uppercase tracking-[0.2em] text-accent/88">{previewMember.role}</p>
                  <p className="mt-4 text-sm leading-7 text-white/72">{previewMember.bio}</p>
                  {form.linkedin_url ? (
                    <span className="mt-5 inline-flex items-center justify-center gap-2 text-sm font-semibold text-white">
                      <Linkedin className="h-4 w-4" />
                      LinkedIn
                    </span>
                  ) : null}
                </div>
              </div>
            </aside>
          </form>
        )}
      </section>

      <section className="panel-card max-h-[720px] overflow-y-auto">
        {loading && <p className="text-sm text-primary/70">{t("messages.loading")}</p>}
        {!loading && members.length === 0 && <p className="text-sm text-primary/70">{t("panelTeam.empty")}</p>}

        <div className="grid gap-3">
          {members.map((member) => {
            const name = pickLocalized(member, "name", lang);
            const role = pickLocalized(member, "role", lang);
            return (
              <div key={member.id} className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-primary/10 bg-white px-5 py-4">
                <div className="flex min-w-0 items-center gap-4">
                  <div className="h-16 w-12 flex-shrink-0 overflow-hidden rounded-lg bg-primary/10">
                    {member.photo_url ? (
                      <img src={resolveImageUrl(member.photo_url)} alt="" className="h-full w-full object-cover" />
                    ) : (
                      <div className="h-full w-full bg-[linear-gradient(145deg,rgba(8,58,79,0.24),rgba(165,141,102,0.22))]" />
                    )}
                  </div>
                  <div className="min-w-0">
                    <p className="truncate text-sm font-semibold text-primary">{name}</p>
                    <p className="truncate text-xs text-primary/55">{role}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2 text-xs text-primary/70">
                  <span>{t("form.orderIndex")}: {member.order_index}</span>
                  <span>{member.is_active ? t("form.active") : t("form.inactive")}</span>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    className="inline-flex items-center gap-2 rounded-full border border-primary/20 px-4 py-2 text-xs font-semibold text-primary/70"
                    onClick={() => handleEdit(member)}
                  >
                    <Pencil className="h-3.5 w-3.5" />
                    {t("actions.edit")}
                  </button>
                  <button
                    type="button"
                    className="inline-flex items-center gap-2 rounded-full border border-red-200 px-4 py-2 text-xs font-semibold text-red-500"
                    onClick={() => handleDelete(member.id)}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    {t("actions.delete")}
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </section>
    </div>
  );
}
