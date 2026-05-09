import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { fetchJSON } from "../lib/api";
import { useTranslation } from "../lib/i18n";
import { getAbsoluteUrl, getCanonicalUrl } from "../lib/seo";

const LOGIN_MODE = "login";
const SIGNUP_MODE = "signup";
const API_BASE = import.meta.env.VITE_API_BASE_URL || "";
const LOGIN_DESCRIPTION_IMAGE = `${API_BASE}/images/login/login_description.png`;

const loginSeoContent = {
  fa: {
    login: {
      title: "ورود | سنگ حسن",
      description:
        "برای مشاهده پروفایل، مدیریت همکاری‌ها و پیگیری درخواست‌ها وارد حساب کاربری سنگ حسن شوید."
    },
    signup: {
      title: "ثبت‌نام | سنگ حسن",
      description:
        "برای شروع همکاری حرفه‌ای در شبکه تامین و تولید سنگ حسن، حساب کاربری جدید ایجاد کنید."
    }
  },
  en: {
    login: {
      title: "Login | SangeHassan",
      description:
        "Sign in to your SangeHassan account to access your profile, requests, and collaboration workflow."
    },
    signup: {
      title: "Sign Up | SangeHassan",
      description:
        "Create a new SangeHassan account to start professional collaboration in our natural stone network."
    }
  },
  ar: {
    login: {
      title: "تسجيل الدخول | سانج حسن",
      description:
        "سجّل الدخول إلى حساب سانج حسن للوصول إلى الملف الشخصي وطلباتك ومسار التعاون."
    },
    signup: {
      title: "إنشاء حساب | سانج حسن",
      description:
        "أنشئ حسابًا جديدًا في سانج حسن لبدء تعاون مهني ضمن شبكة توريد وإنتاج الحجر الطبيعي."
    }
  }
};

function HighlightIcon({ index }) {
  if (index === 0) {
    return (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="feature-icon"
        aria-hidden
      >
        <path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0" />
        <circle cx="12" cy="12" r="3" />
      </svg>
    );
  }

  if (index === 1) {
    return (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="feature-icon"
        aria-hidden
      >
        <path d="M6 22a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.704.706l3.588 3.588A2.4 2.4 0 0 1 20 8v12a2 2 0 0 1-2 2z" />
        <path d="M14 2v5a1 1 0 0 0 1 1h5" />
        <path d="M9 15h6" />
        <path d="M12 18v-6" />
      </svg>
    );
  }

  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className="feature-icon"
      aria-hidden
    >
      <path d="M14 18V6a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v11a1 1 0 0 0 1 1h2" />
      <path d="M15 18H9" />
      <path d="M19 18h2a1 1 0 0 0 1-1v-3.65a1 1 0 0 0-.22-.624l-3.48-4.35A1 1 0 0 0 17.52 8H14" />
      <circle cx="17" cy="18" r="2" />
      <circle cx="7" cy="18" r="2" />
    </svg>
  );
}

export default function Login() {
  const { t, lang } = useTranslation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const initialMode = searchParams.get("mode") === SIGNUP_MODE ? SIGNUP_MODE : LOGIN_MODE;

  const [mode, setMode] = useState(initialMode);
  const [loginPhone, setLoginPhone] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [signupPhone, setSignupPhone] = useState("");
  const [signupPassword, setSignupPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const checkSession = async () => {
      try {
        await fetchJSON("/api/v1/me");
        navigate("/profile", { replace: true });
      } catch (_) {
        // ignore if not logged in
      }
    };
    checkSession();
  }, [navigate]);

  useEffect(() => {
    const searchMode = searchParams.get("mode") === SIGNUP_MODE ? SIGNUP_MODE : LOGIN_MODE;
    setMode((currentMode) => (currentMode === searchMode ? currentMode : searchMode));
  }, [searchParams]);

  const updateMode = (nextMode) => {
    if (nextMode === mode) return;
    setMode(nextMode);
    setError("");

    const nextParams = new URLSearchParams(searchParams);
    if (nextMode === SIGNUP_MODE) {
      nextParams.set("mode", SIGNUP_MODE);
    } else {
      nextParams.delete("mode");
    }
    setSearchParams(nextParams, { replace: true });
  };

  const hydrateSessionUser = async () => {
    try {
      const meRes = await fetchJSON("/api/v1/me");
      const me = meRes?.data || meRes;
      sessionStorage.setItem("sh_me", JSON.stringify(me));
      window.dispatchEvent(new CustomEvent("sh:me-updated", { detail: me }));
    } catch (_) {
      // ignore profile refresh failure here, auth is already successful
    }
  };

  const handleLoginSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError("");

    try {
      await fetchJSON("/api/v1/auth/login", {
        method: "POST",
        body: JSON.stringify({
          phone: loginPhone,
          password: loginPassword
        })
      });

      await hydrateSessionUser();

      const redirectTo = sessionStorage.getItem("sh_after_login") || "/profile";
      sessionStorage.removeItem("sh_after_login");
      navigate(redirectTo, { replace: true });
    } catch (err) {
      setError(err?.message || t("messages.error"));
    } finally {
      setSubmitting(false);
    }
  };

  const handleSignupSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError("");

    try {
      await fetchJSON("/api/v1/auth/signup", {
        method: "POST",
        body: JSON.stringify({
          password: signupPassword,
          phone: signupPhone
        })
      });

      await hydrateSessionUser();
      navigate("/profile", { replace: true });
    } catch (err) {
      setError(err?.message || t("messages.error"));
    } finally {
      setSubmitting(false);
    }
  };

  const isSignup = mode === SIGNUP_MODE;
  const highlights = isSignup ? t("auth.signupHighlights") : t("auth.loginHighlights");
  const highlightItems = Array.isArray(highlights) ? highlights.slice(0, 3) : [];

  useEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return;

    const localized = loginSeoContent[lang] || loginSeoContent.fa;
    const seo = isSignup ? localized.signup : localized.login;
    const pageUrl = getCanonicalUrl(`/login${isSignup ? "?mode=signup" : ""}`);
    const canonicalUrl = getCanonicalUrl("/login");
    const loginImageUrl = getAbsoluteUrl(LOGIN_DESCRIPTION_IMAGE);
    const previousTitle = document.title;
    const cleanups = [];

    const upsertMeta = (selector, createAttrs, value) => {
      let el = document.head.querySelector(selector);
      const created = !el;

      if (!el) {
        el = document.createElement("meta");
        Object.entries(createAttrs).forEach(([key, attrValue]) => {
          el.setAttribute(key, attrValue);
        });
        document.head.appendChild(el);
      }

      const prevContent = el.getAttribute("content");
      el.setAttribute("content", value);

      cleanups.push(() => {
        if (created) {
          el.remove();
          return;
        }
        if (prevContent === null) {
          el.removeAttribute("content");
          return;
        }
        el.setAttribute("content", prevContent);
      });
    };

    const upsertCanonical = (href) => {
      let el = document.head.querySelector('link[rel="canonical"]');
      const created = !el;

      if (!el) {
        el = document.createElement("link");
        el.setAttribute("rel", "canonical");
        document.head.appendChild(el);
      }

      const prevHref = el.getAttribute("href");
      el.setAttribute("href", href);

      cleanups.push(() => {
        if (created) {
          el.remove();
          return;
        }
        if (prevHref === null) {
          el.removeAttribute("href");
          return;
        }
        el.setAttribute("href", prevHref);
      });
    };

    document.title = seo.title;
    upsertCanonical(canonicalUrl);
    upsertMeta('meta[name="description"]', { name: "description" }, seo.description);
    upsertMeta('meta[name="robots"]', { name: "robots" }, "noindex,nofollow,noarchive");
    upsertMeta('meta[property="og:type"]', { property: "og:type" }, "website");
    upsertMeta('meta[property="og:title"]', { property: "og:title" }, seo.title);
    upsertMeta('meta[property="og:description"]', { property: "og:description" }, seo.description);
    upsertMeta('meta[property="og:url"]', { property: "og:url" }, pageUrl);
    upsertMeta('meta[property="og:image"]', { property: "og:image" }, loginImageUrl);
    upsertMeta('meta[name="twitter:card"]', { name: "twitter:card" }, "summary_large_image");
    upsertMeta('meta[name="twitter:title"]', { name: "twitter:title" }, seo.title);
    upsertMeta('meta[name="twitter:description"]', { name: "twitter:description" }, seo.description);
    upsertMeta('meta[name="twitter:image"]', { name: "twitter:image" }, loginImageUrl);

    return () => {
      document.title = previousTitle;
      cleanups.reverse().forEach((fn) => fn());
    };
  }, [lang, isSignup]);

  return (
    <section className="section-shell ">
      <div className="min-h-[calc(100vh-9rem)] overflow-hidden">
        <div className="relative">
          <div className="relative h-[24rem] w-full sm:h-[28rem] md:h-[32rem] lg:h-[30rem]">
            <img
              src={LOGIN_DESCRIPTION_IMAGE}
              alt={t("auth.sidePreviewPlaceholder")}
              className="h-full w-full object-cover object-top"
              loading="eager"
              fetchpriority="high"
              decoding="async"
            />
            <div className="pointer-events-none absolute inset-0 bg-gradient-to-b from-transparent via-sand/35 to-sand/85" />

            <div className="absolute inset-x-0 bottom-0 z-10 px-6 pb-6 sm:px-8 sm:pb-8 lg:px-12 lg:pb-10">
              <div className="relative mx-auto max-w-4xl text-center">
                <div className="pointer-events-none absolute inset-x-4 -bottom-4 -top-4 bg-sand/90 blur-2xl" />

                <div className="relative">
                  <h1 className="font-display text-3xl text-primary">
                    {isSignup ? t("auth.signupTitle") : t("auth.loginTitle")}
                  </h1>
                  <div className="mx-auto mt-[5%] grid max-w-4xl grid-cols-3">
                    {highlightItems.map((item, index) => (
                      <div key={`${item}-${index}`} className="feature-item">
                        <span className="feature-icon-shell">
                          <HighlightIcon index={index} />
                        </span>
                        <span className="feature-title">{item}</span>
                      </div>
                    ))}
                  </div>
                  {/* <p className="mx-auto mt-3 max-w-3xl text-xs leading-6 text-primary/70 sm:text-sm">
                    {isSignup ? t("auth.signupHighlightsFootnote") : t("auth.loginHighlightsFootnote")}
                  </p> */}
                </div>
              </div>
            </div>
          </div>

          <div className="flex justify-center p-6 sm:p-8">
            <div className="w-full max-w-md">
              <form className="space-y-4" onSubmit={isSignup ? handleSignupSubmit : handleLoginSubmit}>
                <label className="block text-xs font-semibold text-primary/70">
                  {t("auth.phone")}
                  <input
                    type="tel"
                    className="mt-2 w-full rounded-full border border-primary/20 bg-white px-4 py-2 text-xs sm:px-5 sm:py-2.5 sm:text-sm"
                    placeholder={t("auth.phone")}
                    value={isSignup ? signupPhone : loginPhone}
                    onChange={(e) => (isSignup ? setSignupPhone(e.target.value) : setLoginPhone(e.target.value))}
                    required
                  />
                </label>

                <label className="block text-xs font-semibold text-primary/70">
                  {t("auth.password")}
                  <input
                    type="password"
                    className="mt-2 w-full rounded-full border border-primary/20 bg-white px-4 py-2 text-xs sm:px-5 sm:py-2.5 sm:text-sm"
                    placeholder={t("auth.password")}
                    value={isSignup ? signupPassword : loginPassword}
                    onChange={(e) => (isSignup ? setSignupPassword(e.target.value) : setLoginPassword(e.target.value))}
                    required
                  />
                </label>

                {error && <p className="text-sm text-red-600">{error}</p>}

                <button
                  type="submit"
                  disabled={submitting}
                  className="w-full rounded-full bg-primary px-4 py-2 text-xs font-semibold text-sand disabled:opacity-60 sm:px-5 sm:py-2.5 sm:text-sm"
                >
                  {submitting
                    ? t("messages.loading")
                    : isSignup
                      ? t("auth.signupButton")
                      : t("auth.loginButton")}
                </button>
              </form>

              <p className="mt-4 text-center text-sm text-primary/70">
                {isSignup ? t("auth.switchToLoginPrompt") : t("auth.switchToSignupPrompt")}{" "}
                <button
                  type="button"
                  className="font-semibold text-primary transition hover:text-accent"
                  onClick={() => updateMode(isSignup ? LOGIN_MODE : SIGNUP_MODE)}
                >
                  {isSignup ? t("auth.switchToLoginAction") : t("auth.switchToSignupAction")}
                </button>
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
