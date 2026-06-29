import { useEffect, useMemo, useRef, useState } from "react";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Link from "@tiptap/extension-link";
import Image from "@tiptap/extension-image";
import Placeholder from "@tiptap/extension-placeholder";
import Table from "@tiptap/extension-table";
import TableRow from "@tiptap/extension-table-row";
import TableHeader from "@tiptap/extension-table-header";
import TableCell from "@tiptap/extension-table-cell";
import {
  Bold,
  Italic,
  Strikethrough,
  Heading2,
  Heading3,
  Heading4,
  List,
  ListOrdered,
  Quote,
  Code2,
  Minus,
  Link2,
  Unlink,
  ImagePlus,
  Table2,
  Undo2,
  Redo2
} from "lucide-react";
import { fetchJSON } from "../lib/api";

const CONTENT_FALLBACK_REV = "html-fallback-20260628";

const isEmptyEditorDoc = (doc) => {
  if (!doc || doc.type !== "doc" || !Array.isArray(doc.content)) return true;
  if (doc.content.length === 0) return true;
  return doc.content.length === 1 && doc.content[0]?.type === "paragraph" && !doc.content[0]?.content?.length;
};

const resolveEditorContent = (value) => {
  if (CONTENT_FALLBACK_REV && value?.html && isEmptyEditorDoc(value?.json)) return value.html;
  return value?.json || value?.html || "";
};

function ToolButton({ title, active = false, disabled = false, onClick, children }) {
  return (
    <button
      type="button"
      title={title}
      aria-label={title}
      aria-pressed={active}
      disabled={disabled}
      onClick={onClick}
      className={`inline-flex h-9 w-9 shrink-0 items-center justify-center border text-primary transition disabled:cursor-not-allowed disabled:opacity-35 ${
        active ? "border-primary bg-primary text-white" : "border-primary/15 bg-white hover:bg-primary/5"
      }`}
    >
      {children}
    </button>
  );
}

export default function RichTextEditor({ value, locale, onChange, onUploadImage }) {
  const fileInputRef = useRef(null);
  const [linkPickerOpen, setLinkPickerOpen] = useState(false);
  const [linkSearch, setLinkSearch] = useState("");
  const [linkTargets, setLinkTargets] = useState([]);
  const [loadingLinks, setLoadingLinks] = useState(false);
  const editor = useEditor({
    extensions: [
      StarterKit.configure({ heading: { levels: [2, 3, 4] } }),
      Link.configure({ openOnClick: false, autolink: true, linkOnPaste: true }),
      Image.configure({ inline: false, allowBase64: false }),
      Placeholder.configure({ placeholder: "Write the article content..." }),
      Table.configure({ resizable: true }),
      TableRow,
      TableHeader,
      TableCell
    ],
    content: resolveEditorContent(value),
    editorProps: {
      attributes: {
        class: "blog-editor-content",
        "data-content-fallback-rev": CONTENT_FALLBACK_REV,
        dir: locale === "fa" || locale === "ar" ? "rtl" : "ltr"
      }
    },
    onUpdate: ({ editor: currentEditor }) => {
      onChange({ html: currentEditor.getHTML(), json: currentEditor.getJSON() });
    }
  });

  useEffect(() => {
    if (!editor) return;
    editor.setOptions({
      editorProps: {
        attributes: {
          class: "blog-editor-content",
          "data-content-fallback-rev": CONTENT_FALLBACK_REV,
          dir: locale === "fa" || locale === "ar" ? "rtl" : "ltr"
        }
      }
    });
  }, [editor, locale]);

  useEffect(() => {
    if (!editor) return;
    const nextHTML = value?.html || "";
    if (editor.getHTML() !== nextHTML) {
      editor.commands.setContent(resolveEditorContent(value), false);
    }
  }, [editor, value?.html, value?.json]);

  useEffect(() => {
    if (!linkPickerOpen || linkTargets.length) return;
    let active = true;
    setLoadingLinks(true);
    Promise.all([
      fetchJSON("/api/products?limit=500&offset=0").catch(() => ({ data: [] })),
      fetchJSON("/api/categories").catch(() => ({ data: [] })),
      fetchJSON("/api/projects").catch(() => ({ data: [] })),
      fetchJSON(`/api/blogs?locale=${locale}`).catch(() => ({ data: [] }))
    ]).then(([products, categories, projects, blogs]) => {
      if (!active) return;
      const localized = (item, field) => item?.[`${field}_${locale}`] || item?.[`${field}_en`] || item?.[`${field}_fa`] || item?.[field] || "";
      setLinkTargets([
        { type: "Page", label: "Products", href: `/${locale}/products` },
        { type: "Page", label: "Projects", href: "/projects" },
        { type: "Page", label: "Articles", href: `/${locale}/blogs` },
        { type: "Page", label: "About", href: "/about" },
        ...(products.data || []).filter((item) => item.slug).map((item) => ({ type: "Product", label: localized(item, "title") || item.slug, href: `/${locale}/products/${item.slug}` })),
        ...(categories.data || []).filter((item) => item.slug).map((item) => ({ type: "Category", label: localized(item, "title") || item.slug, href: `/${locale}/products/${item.slug}` })),
        ...(projects.data || []).map((item) => ({ type: "Project", label: localized(item, "title") || `Project ${item.id}`, href: `/projects/${item.id}` })),
        ...(blogs.data || []).filter((item) => item.slug).map((item) => ({ type: "Article", label: item.title, href: `/${locale}/blogs/${item.slug}` }))
      ]);
    }).finally(() => {
      if (active) setLoadingLinks(false);
    });
    return () => { active = false; };
  }, [linkPickerOpen, linkTargets.length, locale]);

  const visibleLinks = useMemo(() => {
    const query = linkSearch.trim().toLocaleLowerCase(locale);
    if (!query) return linkTargets.slice(0, 80);
    return linkTargets.filter((item) => `${item.type} ${item.label} ${item.href}`.toLocaleLowerCase(locale).includes(query)).slice(0, 80);
  }, [linkSearch, linkTargets, locale]);

  if (!editor) return <div className="h-64 animate-pulse border border-primary/10 bg-primary/5" />;

  const setCustomLink = () => {
    const previous = editor.getAttributes("link").href || "";
    const href = window.prompt("Link URL", previous);
    if (href === null) return;
    if (!href.trim()) {
      editor.chain().focus().extendMarkRange("link").unsetLink().run();
      return;
    }
    editor.chain().focus().extendMarkRange("link").setLink({ href: href.trim() }).run();
  };

  const applyInternalLink = (target) => {
    const selectedText = editor.state.doc.textBetween(editor.state.selection.from, editor.state.selection.to, " ").trim();
    if (selectedText) {
      editor.chain().focus().extendMarkRange("link").setLink({ href: target.href }).run();
    } else {
      editor.chain().focus().insertContent(`<a href="${target.href}">${target.label}</a>`).run();
    }
    setLinkPickerOpen(false);
    setLinkSearch("");
  };

  const uploadInlineImage = async (file) => {
    if (!file) return;
    const src = await onUploadImage(file);
    if (!src) return;
    const alt = window.prompt("Image alt text", "") || "";
    editor.chain().focus().setImage({ src, alt, title: alt }).run();
  };

  return (
    <div className="overflow-hidden border border-primary/20 bg-white">
      <div className="flex flex-wrap gap-1 border-b border-primary/10 bg-primary/[0.025] p-2">
        <ToolButton title="Bold" active={editor.isActive("bold")} onClick={() => editor.chain().focus().toggleBold().run()}><Bold size={17} /></ToolButton>
        <ToolButton title="Italic" active={editor.isActive("italic")} onClick={() => editor.chain().focus().toggleItalic().run()}><Italic size={17} /></ToolButton>
        <ToolButton title="Strike" active={editor.isActive("strike")} onClick={() => editor.chain().focus().toggleStrike().run()}><Strikethrough size={17} /></ToolButton>
        <ToolButton title="Heading 2" active={editor.isActive("heading", { level: 2 })} onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}><Heading2 size={18} /></ToolButton>
        <ToolButton title="Heading 3" active={editor.isActive("heading", { level: 3 })} onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}><Heading3 size={18} /></ToolButton>
        <ToolButton title="Heading 4" active={editor.isActive("heading", { level: 4 })} onClick={() => editor.chain().focus().toggleHeading({ level: 4 }).run()}><Heading4 size={18} /></ToolButton>
        <ToolButton title="Bullet list" active={editor.isActive("bulletList")} onClick={() => editor.chain().focus().toggleBulletList().run()}><List size={18} /></ToolButton>
        <ToolButton title="Numbered list" active={editor.isActive("orderedList")} onClick={() => editor.chain().focus().toggleOrderedList().run()}><ListOrdered size={18} /></ToolButton>
        <ToolButton title="Quote" active={editor.isActive("blockquote")} onClick={() => editor.chain().focus().toggleBlockquote().run()}><Quote size={18} /></ToolButton>
        <ToolButton title="Code block" active={editor.isActive("codeBlock")} onClick={() => editor.chain().focus().toggleCodeBlock().run()}><Code2 size={18} /></ToolButton>
        <ToolButton title="Horizontal rule" onClick={() => editor.chain().focus().setHorizontalRule().run()}><Minus size={18} /></ToolButton>
        <ToolButton title="Choose internal link" active={editor.isActive("link")} onClick={() => setLinkPickerOpen((open) => !open)}><Link2 size={18} /></ToolButton>
        <ToolButton title="Remove link" disabled={!editor.isActive("link")} onClick={() => editor.chain().focus().unsetLink().run()}><Unlink size={18} /></ToolButton>
        <ToolButton title="Upload image" onClick={() => fileInputRef.current?.click()}><ImagePlus size={18} /></ToolButton>
        <ToolButton title="Insert table" onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()}><Table2 size={18} /></ToolButton>
        <ToolButton title="Undo" disabled={!editor.can().undo()} onClick={() => editor.chain().focus().undo().run()}><Undo2 size={18} /></ToolButton>
        <ToolButton title="Redo" disabled={!editor.can().redo()} onClick={() => editor.chain().focus().redo().run()}><Redo2 size={18} /></ToolButton>
        <input ref={fileInputRef} type="file" accept="image/png,image/jpeg,image/webp" className="hidden" onChange={(event) => uploadInlineImage(event.target.files?.[0])} />
      </div>
      {linkPickerOpen && (
        <div className="border-b border-primary/10 bg-white p-3">
          <div className="flex gap-2">
            <input autoFocus value={linkSearch} onChange={(event) => setLinkSearch(event.target.value)} placeholder="Search products, categories, projects and articles" className="min-w-0 flex-1 border border-primary/20 px-3 py-2 text-sm outline-none focus:border-accent" />
            <button type="button" onClick={setCustomLink} className="border border-primary/20 px-3 py-2 text-xs font-semibold">Custom URL</button>
            <button type="button" title="Close" onClick={() => setLinkPickerOpen(false)} className="inline-flex h-9 w-9 items-center justify-center border border-primary/20">×</button>
          </div>
          <div className="mt-2 max-h-56 overflow-y-auto border border-primary/10">
            {loadingLinks ? <p className="p-3 text-xs text-primary/50">Loading links...</p> : visibleLinks.length === 0 ? <p className="p-3 text-xs text-primary/50">No link found.</p> : visibleLinks.map((target) => (
              <button key={`${target.type}-${target.href}`} type="button" onClick={() => applyInternalLink(target)} className="flex w-full items-center justify-between gap-4 border-b border-primary/10 px-3 py-2 text-left text-sm last:border-b-0 hover:bg-primary/[0.035]">
                <span className="min-w-0 truncate font-medium">{target.label}</span>
                <span className="shrink-0 text-[10px] font-semibold uppercase text-primary/40">{target.type}</span>
              </button>
            ))}
          </div>
        </div>
      )}
      <EditorContent editor={editor} />
    </div>
  );
}
