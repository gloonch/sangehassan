import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { gsap } from "gsap";
import { ChevronDown } from "lucide-react";
import { useTranslation } from "../lib/i18n";
import { fetchJSON } from "../lib/api";
import { resolveImageUrl } from "../lib/assets";
import { getAbsoluteUrl, getCanonicalUrl, getSiteOrigin } from "../lib/seo";
import blocksOverlayImage from "@shared/assets/landing_page/landingpage_blocks_overlay.webp";
import marketComplexityIllustration from "@shared/assets/landing_icons/market_complexity_icon_transparent.webp";
import networkSupplyIllustration from "@shared/assets/landing_icons/network_supply_icon_transparent.webp";
import trustQualityIllustration from "@shared/assets/landing_icons/trust_quality_icon_transparent.webp";
import blockImage01 from "@shared/assets/landing_page/blocks/block-slide-01.webp";
import blockImage02 from "@shared/assets/landing_page/blocks/block-slide-02.webp";
import blockImage03 from "@shared/assets/landing_page/blocks/block-slide-03.webp";
import blockImage04 from "@shared/assets/landing_page/blocks/block-slide-04.webp";
import blockImage05 from "@shared/assets/landing_page/blocks/block-slide-05.webp";
import blockImage06 from "@shared/assets/landing_page/blocks/block-slide-06.webp";
import blockImage07 from "@shared/assets/landing_page/blocks/block-slide-07.webp";
import blockImage08 from "@shared/assets/landing_page/blocks/block-slide-08.webp";
import finishesImage01 from "@shared/assets/landing_page/products/finish-slide-01.webp";
import finishesImage02 from "@shared/assets/landing_page/products/finish-slide-02.webp";
import finishesImage03 from "@shared/assets/landing_page/products/finish-slide-03.webp";
import finishesImage04 from "@shared/assets/landing_page/products/finish-slide-04.webp";
import productImage01 from "@shared/assets/landing_page/products/product-slide-01.webp";
import productImage02 from "@shared/assets/landing_page/products/product-slide-02.webp";
import productImage03 from "@shared/assets/landing_page/products/product-slide-03.webp";

const productSlides = [
  { src: finishesImage01, width: 736, height: 981 },
  { src: finishesImage02, width: 736, height: 1508 },
  { src: finishesImage03, width: 736, height: 1308 },
  { src: finishesImage04, width: 735, height: 825 },
  { src: productImage01, width: 900, height: 862 },
  { src: productImage02, width: 900, height: 864 },
  { src: productImage03, width: 900, height: 850 }
];
const blockSlides = [
  { src: blockImage01, width: 900, height: 895 },
  { src: blockImage02, width: 900, height: 894 },
  { src: blockImage03, width: 725, height: 906 },
  { src: blockImage04, width: 900, height: 1123 },
  { src: blockImage05, width: 900, height: 1123 },
  { src: blockImage06, width: 900, height: 1104 },
  { src: blockImage07, width: 900, height: 1106 },
  { src: blockImage08, width: 900, height: 1104 }
];
const fallbackSlide = { src: blocksOverlayImage, width: 1068, height: 845 };

const shuffleSlides = (slides) => {
  const shuffled = [...slides];
  for (let index = shuffled.length - 1; index > 0; index -= 1) {
    const swapIndex = Math.floor(Math.random() * (index + 1));
    [shuffled[index], shuffled[swapIndex]] = [shuffled[swapIndex], shuffled[index]];
  }
  return shuffled;
};

const initialSlideDecks = () => ({
  products: productSlides.length ? productSlides : [fallbackSlide],
  blocks: blockSlides.length ? blockSlides : [fallbackSlide]
});

const getLocalStorageValue = (key) => {
  if (typeof window === "undefined") return null;
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
};

const setLocalStorageValue = (key, value) => {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(key, value);
  } catch {
    /* Storage can be unavailable in private or restricted browser modes. */
  }
};

const shuffleForNewVisit = (slides, storageKey) => {
  const shuffled = shuffleSlides(slides);
  if (typeof window === "undefined" || shuffled.length < 2) return shuffled;

  const previousFirstSlide = getLocalStorageValue(storageKey);
  if (previousFirstSlide && shuffled[0]?.src === previousFirstSlide) {
    const swapIndex = 1 + Math.floor(Math.random() * (shuffled.length - 1));
    [shuffled[0], shuffled[swapIndex]] = [shuffled[swapIndex], shuffled[0]];
  }

  setLocalStorageValue(storageKey, shuffled[0]?.src || "");
  return shuffled;
};

const createSlideDecks = () => ({
  products: shuffleForNewVisit(productSlides.length ? productSlides : [fallbackSlide], "sh-home-products-first-slide"),
  blocks: shuffleForNewVisit(blockSlides.length ? blockSlides : [fallbackSlide], "sh-home-blocks-first-slide")
});

const slideTransitionMs = 1800;

const getNextSlideIndex = (activeSlide, slides) => {
  if (!slides.length || slides.length < 2) return null;
  return (activeSlide + 1) % slides.length;
};

const isSlideDebugEnabled = () => {
  if (typeof window === "undefined") return false;
  return window.location.search.includes("debugSlides=1") || getLocalStorageValue("sh-debug-slides") === "1";
};

const getSlideAssetName = (src = "") => src.split("/").pop() || src;

const logSlideDebug = (event, payload = {}) => {
  if (!isSlideDebugEnabled()) return;
  console.info("[home-slides]", event, {
    at: Math.round(performance.now()),
    ...payload
  });
};

const getSlideDebugHandlers = (panel, role, slideIndex, image) => {
  if (!isSlideDebugEnabled()) return {};
  const asset = getSlideAssetName(image.src);
  const readStyles = (target) => {
    const styles = window.getComputedStyle(target);
    return {
      opacity: styles.opacity,
      filter: styles.filter,
      transitionDuration: styles.transitionDuration
    };
  };

  return {
    onLoad: (event) => {
      logSlideDebug("image-load", {
        panel,
        role,
        slideIndex,
        asset,
        ...readStyles(event.currentTarget)
      });
    },
    onTransitionStart: (event) => {
      if (event.propertyName !== "opacity") return;
      logSlideDebug("transition-start", {
        panel,
        role,
        slideIndex,
        asset,
        ...readStyles(event.currentTarget)
      });
    },
    onTransitionEnd: (event) => {
      if (event.propertyName !== "opacity") return;
      logSlideDebug("transition-end", {
        panel,
        role,
        slideIndex,
        asset,
        ...readStyles(event.currentTarget)
      });
    }
  };
};

const homeSeoContent = {
  fa: {
    title: "سنگ حسن | تامین و تولید سنگ طبیعی",
    description:
      "شبکه تامین و تولید سنگ طبیعی سنگ حسن؛ از کوپ و بلوک تا محصولات فرآوری‌شده برای پروژه‌های حرفه‌ای، B2B و صادرات.",
    locale: "fa_IR"
  },
  en: {
    title: "SangeHassan | Natural Stone Supply & Production",
    description:
      "Integrated natural stone supply and production network, from quarry blocks to finished stone products for professional projects, B2B, and export.",
    locale: "en_US"
  },
  ar: {
    title: "سانج حسن | توريد وإنتاج الحجر الطبيعي",
    description:
      "شبكة متكاملة لتوريد وإنتاج الحجر الطبيعي من البلوك الخام حتى المنتجات المعالجة للمشاريع المهنية والتعاون B2B والتصدير.",
    locale: "ar_SA"
  }
};

const pickField = (section, field, lang) => {
  if (!section) return "";
  const key = `${field}_${lang}`;
  return section[key] || section[`${field}_en`] || section[`${field}_fa`] || section[`${field}_ar`] || "";
};

const splitLines = (text) =>
  text
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);

const whySangehassanIllustrations = {
  market: {
    src: marketComplexityIllustration,
    width: 802,
    height: 648
  },
  network: {
    src: networkSupplyIllustration,
    width: 941,
    height: 942
  },
  trust: {
    src: trustQualityIllustration,
    width: 1059,
    height: 726
  }
};

const whySangehassanContent = {
  fa: {
    eyebrow: "چرا سنگ حسن؟",
    title: "بازار سنگ نباید اینقدر سخت باشد",
    intro:
      "خرید سنگ ساختمانی هنوز برای بسیاری از پروژه‌ها مسیری پیچیده است؛ تماس‌های متعدد، بازدیدهای پی‌درپی، قیمت‌های متفاوت، اطلاعات ناقص و انتخاب‌هایی که اغلب به موجودی چند انبار محدود می‌شوند.",
    support:
      "مشکل بازار سنگ، کمبود سنگ خوب نیست. مشکل، پراکندگی اطلاعات، فاصله با منبع واقعی و سختی دسترسی به انتخاب درست است.",
    quote: "انتخاب شما نباید به موجودی یک انبار محدود شود؛ باید به ظرفیت یک شبکه متصل باشد.",
    cta: "مشاهده محصولات",
    cards: [
      {
        key: "market",
        title: "بازار پراکنده، تصمیم سخت",
        imageAlt: "مسیرهای پراکنده بازار سنگ و انتخاب سخت",
        summary:
          "خریدار معمولاً بین چند فروشنده، چند قیمت و روایت متفاوت تصمیم می‌گیرد؛ سنگ حسن این فاصله را با نمایش شفاف‌تر محصولات، پروژه‌ها و منابع تأمین کمتر می‌کند.",
        paragraphs: [
          "در بازار سنتی سنگ، خریدار معمولاً باید بین چند فروشنده، چند قیمت و چند روایت متفاوت تصمیم بگیرد. بسیاری از گزینه‌های بهتر اصلاً دیده نمی‌شوند، چون در مسیر جست‌وجوی مشتری قرار نمی‌گیرند.",
          "سنگ حسن تلاش می‌کند این فاصله را کمتر کند؛ با نمایش شفاف‌تر محصولات، پروژه‌ها، منابع تأمین و مسیر انتخاب."
        ]
      },
      {
        key: "network",
        title: "انتخاب محدود به یک انبار نیست",
        imageAlt: "شبکه‌ای از منابع تأمین سنگ متصل به یک هاب مرکزی",
        summary:
          "انتخاب شما نباید به موجودی یک انبار محدود شود؛ سنگ حسن به شبکه‌ای از تولیدکنندگان، معادن، کارخانه‌ها و منابع تأمین وصل است تا گزینه مناسب‌تر پروژه پیدا شود.",
        paragraphs: [
          "سنگ حسن یک انبار سنتی نیست. این یک تصمیم آگاهانه است.",
          "ما نمی‌خواهیم انتخاب شما فقط به موجودی خودمان محدود شود. به‌جای انبارمحور بودن، به شبکه‌ای از تولیدکنندگان، معادن، کارخانه‌ها و منابع تأمین متصل هستیم تا برای هر پروژه، گزینه مناسب‌تر پیدا شود.",
          "هدف ما فروش آن چیزی نیست که صرفاً موجود است؛ هدف ما پیدا کردن سنگی است که برای پروژه شما درست‌تر است."
        ]
      },
      {
        key: "trust",
        title: "رقابت بر پایه اعتماد، نه قیمت‌شکنی",
        imageAlt: "کنترل کیفیت و اعتماد در انتخاب سنگ ساختمانی",
        summary:
          "ارزان‌ترین انتخاب همیشه کم‌هزینه‌ترین نیست؛ ما روی کیفیت، خدمات و اعتماد رقابت می‌کنیم تا خرید سنگ شفاف‌تر و مطمئن‌تر شود.",
        paragraphs: [
          "در پروژه‌های ساختمانی، ارزان‌ترین انتخاب همیشه کم‌هزینه‌ترین انتخاب نیست. انتخاب اشتباه سنگ می‌تواند باعث پرت، دوباره‌کاری، تأخیر، ناهماهنگی رنگ و افت کیفیت نهایی شود.",
          "ما در کیفیت، خدمات و اعتماد رقابت می‌کنیم؛ نه در قیمت‌شکنی.",
          "سنگ حسن تلاش می‌کند خرید سنگ را از یک تصمیم پراکنده و پرریسک، به یک فرآیند شفاف‌تر، قابل بررسی‌تر و مطمئن‌تر تبدیل کند."
        ]
      }
    ]
  },
  en: {
    eyebrow: "Why SangeHassan?",
    title: "The stone market should not be this hard",
    intro:
      "Buying building stone is still a fragmented path for many projects: repeated calls, repeated visits, different prices, incomplete information, and choices often limited to a few warehouses.",
    support:
      "The problem is not a lack of good stone. The problem is scattered information, distance from the real source, and hard access to the right option.",
    quote: "Your choice should not be limited to one warehouse inventory; it should be connected to the capacity of a network.",
    cta: "View products",
    cards: [
      {
        key: "market",
        title: "A scattered market, a hard decision",
        imageAlt: "Scattered stone market paths and difficult selection",
        summary:
          "Buyers often choose between several sellers, prices, and narratives; SangeHassan reduces that distance with clearer product, project, and supply visibility.",
        paragraphs: [
          "In the traditional stone market, buyers usually have to decide between several sellers, several prices, and several different narratives. Many better options are never seen because they are not in the customer's search path.",
          "SangeHassan works to reduce that distance through clearer visibility into products, projects, supply sources, and the selection path."
        ]
      },
      {
        key: "network",
        title: "Selection is not limited to one warehouse",
        imageAlt: "Connected stone supply network around a central hub",
        summary:
          "Your options should not be limited to one warehouse; SangeHassan connects producers, quarries, factories, and supply sources to find a better fit.",
        paragraphs: [
          "SangeHassan is not a traditional warehouse. That is a deliberate choice.",
          "We do not want your selection to be limited to what we personally have in stock. Instead of being warehouse-driven, we are connected to a network of producers, quarries, factories, and supply sources to find the better option for each project.",
          "The goal is not to sell only what is available; the goal is to find the stone that is more right for your project."
        ]
      },
      {
        key: "trust",
        title: "Competing on trust, not price-cutting",
        imageAlt: "Quality control and trust in building stone selection",
        summary:
          "The cheapest stone is not always the lowest-cost choice; we compete on quality, service, and trust to make buying clearer and more reliable.",
        paragraphs: [
          "In construction projects, the cheapest choice is not always the lowest-cost choice. The wrong stone can cause waste, rework, delays, color mismatch, and weaker final quality.",
          "We compete on quality, service, and trust; not on price-cutting.",
          "SangeHassan works to turn stone buying from a scattered and risky decision into a clearer, more reviewable, and more reliable process."
        ]
      }
    ]
  },
  ar: {
    eyebrow: "لماذا سانج حسن؟",
    title: "لا ينبغي أن يكون سوق الحجر بهذه الصعوبة",
    intro:
      "ما زال شراء حجر البناء في كثير من المشاريع مساراً معقداً؛ اتصالات متعددة، زيارات متكررة، أسعار مختلفة، معلومات ناقصة، وخيارات غالباً ما تنحصر في مخزون بضعة مستودعات.",
    support:
      "مشكلة سوق الحجر ليست ندرة الحجر الجيد. المشكلة هي تشتت المعلومات، والبعد عن المصدر الحقيقي، وصعوبة الوصول إلى الخيار الصحيح.",
    quote: "لا ينبغي أن يقتصر اختيارك على مخزون مستودع واحد؛ بل يجب أن يتصل بقدرة شبكة كاملة.",
    cta: "عرض المنتجات",
    cards: [
      {
        key: "market",
        title: "سوق متشتت وقرار صعب",
        imageAlt: "مسارات متفرقة في سوق الحجر واختيار صعب",
        summary:
          "غالباً ما يختار المشتري بين عدة بائعين وأسعار وروايات؛ وتعمل سانج حسن على تقليل هذه المسافة بعرض أوضح للمنتجات والمشاريع ومصادر التوريد.",
        paragraphs: [
          "في سوق الحجر التقليدي، يضطر المشتري غالباً إلى الاختيار بين عدة بائعين وعدة أسعار وعدة روايات مختلفة. كثير من الخيارات الأفضل لا تظهر أصلاً لأنها ليست في طريق بحث العميل.",
          "تعمل سانج حسن على تقليل هذه المسافة من خلال عرض أوضح للمنتجات والمشاريع ومصادر التوريد ومسار الاختيار."
        ]
      },
      {
        key: "network",
        title: "الاختيار لا يقتصر على مستودع واحد",
        imageAlt: "شبكة توريد حجر متصلة حول مركز واحد",
        summary:
          "لا ينبغي أن يكون اختيارك محدوداً بمستودع واحد؛ سانج حسن متصل بالمنتجين والمحاجر والمصانع ومصادر التوريد للوصول إلى الأنسب.",
        paragraphs: [
          "سانج حسن ليس مستودعاً تقليدياً. هذا خيار واعٍ.",
          "لا نريد أن يكون اختيارك محدوداً بما نملكه فقط. بدلاً من العمل بمنطق المستودع، نحن متصلون بشبكة من المنتجين والمحاجر والمصانع ومصادر التوريد للوصول إلى الخيار الأنسب لكل مشروع.",
          "هدفنا ليس بيع ما هو متوفر فقط؛ هدفنا العثور على الحجر الأنسب لمشروعك."
        ]
      },
      {
        key: "trust",
        title: "المنافسة على الثقة لا على كسر الأسعار",
        imageAlt: "فحص جودة وثقة في اختيار حجر البناء",
        summary:
          "الخيار الأرخص ليس دائماً الأقل تكلفة؛ نحن ننافس في الجودة والخدمة والثقة حتى يصبح شراء الحجر أوضح وأكثر اطمئناناً.",
        paragraphs: [
          "في مشاريع البناء، الخيار الأرخص ليس دائماً الأقل تكلفة. اختيار الحجر الخطأ قد يؤدي إلى هدر، وإعادة عمل، وتأخير، وعدم تناسق في اللون، وانخفاض جودة النتيجة النهائية.",
          "نحن ننافس في الجودة والخدمة والثقة؛ لا في كسر الأسعار.",
          "تعمل سانج حسن على تحويل شراء الحجر من قرار متشتت ومحفوف بالمخاطر إلى عملية أكثر وضوحاً وقابلية للمراجعة واطمئناناً."
        ]
      }
    ]
  }
};

const teamSectionContent = {
  fa: {
    eyebrow: "تیم",
    title: "تیم پشت سنگ حسن",
    intro:
      "برای ساخت یک شبکه قابل اعتماد در صنعت سنگ، فقط محصول کافی نیست. سنگ حسن با ترکیب تجربه بازار سنگ، فناوری، امور حقوقی و مشاوره طراحی تلاش می‌کند مسیر انتخاب و تأمین سنگ را شفاف‌تر و مطمئن‌تر کند.",
    linkedinLabel: "LinkedIn",
    cards: [
      {
        initials: "FS",
        name: "Founder / Stone Supply Lead",
        role: "هدایت شبکه تأمین",
        bio: "مسئول توسعه مسیرهای تأمین، ارتباط با تولیدکنندگان و هماهنگی ظرفیت‌های بازار سنگ."
      },
      {
        initials: "CT",
        name: "CTO",
        role: "فناوری و محصول",
        bio: "مسئول توسعه محصول دیجیتال، زیرساخت فنی و مسیرهای قابل پیگیری در تجربه کاربر."
      },
      {
        initials: "LC",
        name: "Legal & Contracts",
        role: "امور حقوقی و قراردادها",
        bio: "همراهی در تنظیم قراردادها، مستندسازی و هماهنگی‌های رسمی معاملات."
      },
      {
        initials: "DC",
        name: "Design Consultant",
        role: "مشاوره طراحی و انتخاب سنگ",
        bio: "مشاور انتخاب سنگ، متریال‌شناسی و هماهنگی سنگ با نیاز پروژه‌های معماری."
      }
    ]
  },
  en: {
    eyebrow: "Team",
    title: "The Team Behind SangeHassan",
    intro:
      "A reliable stone network needs more than products. SangeHassan combines stone-market experience, technology, legal coordination, and design consulting to make selection and supply clearer and more dependable.",
    linkedinLabel: "LinkedIn",
    cards: [
      {
        initials: "FS",
        name: "Founder / Stone Supply Lead",
        role: "Supply network leadership",
        bio: "Leads supply paths, producer relationships, and coordination across the stone market."
      },
      {
        initials: "CT",
        name: "CTO",
        role: "Technology and product",
        bio: "Leads digital product development, technical infrastructure, and traceable user workflows."
      },
      {
        initials: "LC",
        name: "Legal & Contracts",
        role: "Legal and contracts",
        bio: "Supports contract preparation, documentation, and formal transaction coordination."
      },
      {
        initials: "DC",
        name: "Design Consultant",
        role: "Stone and design consulting",
        bio: "Advises on stone selection, material fit, and alignment with architectural project needs."
      }
    ]
  },
  ar: {
    eyebrow: "الفريق",
    title: "الفريق وراء سانج حسن",
    intro:
      "بناء شبكة موثوقة في صناعة الحجر يحتاج إلى أكثر من المنتجات. تجمع سانج حسن بين خبرة سوق الحجر والتكنولوجيا والتنسيق القانوني والاستشارات التصميمية لجعل الاختيار والتوريد أوضح وأكثر اطمئناناً.",
    linkedinLabel: "LinkedIn",
    cards: [
      {
        initials: "FS",
        name: "Founder / Stone Supply Lead",
        role: "قيادة شبكة التوريد",
        bio: "مسؤول عن تطوير مسارات التوريد والعلاقات مع المنتجين وتنسيق قدرات سوق الحجر."
      },
      {
        initials: "CT",
        name: "CTO",
        role: "التكنولوجيا والمنتج",
        bio: "مسؤول عن تطوير المنتج الرقمي والبنية التقنية ومسارات تجربة المستخدم القابلة للمتابعة."
      },
      {
        initials: "LC",
        name: "Legal & Contracts",
        role: "الشؤون القانونية والعقود",
        bio: "يدعم إعداد العقود والتوثيق والتنسيقات الرسمية للمعاملات."
      },
      {
        initials: "DC",
        name: "Design Consultant",
        role: "استشارات التصميم واختيار الحجر",
        bio: "يقدم المشورة في اختيار الحجر وملاءمة المواد مع احتياجات المشاريع المعمارية."
      }
    ]
  }
};

export default function Home() {
  const { t, lang } = useTranslation();
  const [sections, setSections] = useState([]);
  const [teamMembers, setTeamMembers] = useState([]);
  const [slideDecks, setSlideDecks] = useState(initialSlideDecks);
  const [activeSlides, setActiveSlides] = useState({ products: 0, blocks: 0 });
  const [previousSlides, setPreviousSlides] = useState({ products: null, blocks: null });
  const previousClearTimerRef = useRef(null);
  const rootRef = useRef(null);

  useEffect(() => {
    if (typeof window === "undefined" || typeof document === "undefined") return;

    const seo = homeSeoContent[lang] || homeSeoContent.fa;
    const pageUrl = getCanonicalUrl("/");
    const ogImage = getAbsoluteUrl(blocksOverlayImage);
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

    const upsertJsonLd = (payload) => {
      const scriptId = "home-jsonld";
      let script = document.getElementById(scriptId);
      const created = !script;

      if (!script) {
        script = document.createElement("script");
        script.setAttribute("id", scriptId);
        script.setAttribute("type", "application/ld+json");
        document.head.appendChild(script);
      }

      const prevContent = script.textContent;
      script.textContent = JSON.stringify(payload);

      cleanups.push(() => {
        if (created) {
          script.remove();
          return;
        }
        script.textContent = prevContent;
      });
    };

    document.title = seo.title;
    upsertCanonical(pageUrl);
    upsertMeta('meta[name="description"]', { name: "description" }, seo.description);
    upsertMeta('meta[name="robots"]', { name: "robots" }, "index,follow");
    upsertMeta('meta[property="og:type"]', { property: "og:type" }, "website");
    upsertMeta('meta[property="og:title"]', { property: "og:title" }, seo.title);
    upsertMeta('meta[property="og:description"]', { property: "og:description" }, seo.description);
    upsertMeta('meta[property="og:url"]', { property: "og:url" }, pageUrl);
    upsertMeta('meta[property="og:image"]', { property: "og:image" }, ogImage);
    upsertMeta('meta[property="og:locale"]', { property: "og:locale" }, seo.locale);
    upsertMeta('meta[name="twitter:card"]', { name: "twitter:card" }, "summary_large_image");
    upsertMeta('meta[name="twitter:title"]', { name: "twitter:title" }, seo.title);
    upsertMeta('meta[name="twitter:description"]', { name: "twitter:description" }, seo.description);
    upsertMeta('meta[name="twitter:image"]', { name: "twitter:image" }, ogImage);

    upsertJsonLd({
      "@context": "https://schema.org",
      "@type": "WebSite",
      inLanguage: lang,
      name: "SangeHassan",
      url: pageUrl,
      description: seo.description,
      publisher: {
        "@type": "Organization",
        name: "SangeHassan",
        url: getSiteOrigin()
      }
    });

    return () => {
      document.title = previousTitle;
      cleanups.reverse().forEach((fn) => fn());
    };
  }, [lang]);

  useEffect(() => {
    let mounted = true;
    fetchJSON("/api/content-sections?page=home")
      .then((res) => {
        if (!mounted) return;
        setSections(res.data || []);
      })
      .catch(() => {
        if (!mounted) return;
        setSections([]);
      });
    return () => {
      mounted = false;
    };
  }, []);

  useEffect(() => {
    let mounted = true;
    fetchJSON("/api/team-members")
      .then((res) => {
        if (!mounted) return;
        setTeamMembers(res.data || []);
      })
      .catch(() => {
        if (!mounted) return;
        setTeamMembers([]);
      });
    return () => {
      mounted = false;
    };
  }, []);

  useEffect(() => {
    const decks = createSlideDecks();
    logSlideDebug("deck-init", {
      products: decks.products.map((slide) => getSlideAssetName(slide.src)),
      blocks: decks.blocks.map((slide) => getSlideAssetName(slide.src))
    });
    setSlideDecks(decks);
    setActiveSlides({ products: 0, blocks: 0 });
    setPreviousSlides({ products: null, blocks: null });
  }, []);

  useEffect(() => {
    const timer = window.setInterval(() => {
      setActiveSlides((current) => {
        const next = {
          products: (current.products + 1) % slideDecks.products.length,
          blocks: (current.blocks + 1) % slideDecks.blocks.length
        };

        logSlideDebug("tick", {
          previous: current,
          next,
          products: {
            previousAsset: getSlideAssetName(slideDecks.products[current.products]?.src),
            nextAsset: getSlideAssetName(slideDecks.products[next.products]?.src)
          },
          blocks: {
            previousAsset: getSlideAssetName(slideDecks.blocks[current.blocks]?.src),
            nextAsset: getSlideAssetName(slideDecks.blocks[next.blocks]?.src)
          }
        });

        setPreviousSlides(current);

        if (previousClearTimerRef.current) {
          window.clearTimeout(previousClearTimerRef.current);
        }
        previousClearTimerRef.current = window.setTimeout(() => {
          setPreviousSlides({ products: null, blocks: null });
        }, slideTransitionMs + 120);

        return next;
      });
    }, 2000);

    return () => {
      window.clearInterval(timer);
      if (previousClearTimerRef.current) {
        window.clearTimeout(previousClearTimerRef.current);
      }
    };
  }, [slideDecks]);

  const sectionsByKey = useMemo(() => {
    const map = {};
    sections.forEach((section) => {
      map[section.key] = section;
    });
    return map;
  }, [sections]);

  const fallbackSections = useMemo(
    () => ({
      blocks: {
        key: "blocks",
        title_en: t("blocks.title"),
        title_fa: t("blocks.title"),
        title_ar: t("blocks.title"),
        subtitle_en: t("blocks.subtitle"),
        subtitle_fa: t("blocks.subtitle"),
        subtitle_ar: t("blocks.subtitle"),
        description_en: t("blocks.subtitle"),
        description_fa: t("blocks.subtitle"),
        description_ar: t("blocks.subtitle"),
        cta_label_en: t("blocks.cta"),
        cta_label_fa: t("blocks.cta"),
        cta_label_ar: t("blocks.cta"),
        cta_href: "/blocks",
        images: []
      },
      finished: {
        key: "finished",
        title_en: t("products.title"),
        title_fa: t("products.title"),
        title_ar: t("products.title"),
        subtitle_en: t("products.subtitle"),
        subtitle_fa: t("products.subtitle"),
        subtitle_ar: t("products.subtitle"),
        description_en: t("products.subtitle"),
        description_fa: t("products.subtitle"),
        description_ar: t("products.subtitle"),
        cta_label_en: t("hero.ctaPrimary"),
        cta_label_fa: t("hero.ctaPrimary"),
        cta_label_ar: t("hero.ctaPrimary"),
        cta_href: "/products",
        images: []
      }
    }),
    [t]
  );

  const blocksSection = sectionsByKey.blocks || fallbackSections.blocks;
  const finishedSection = sectionsByKey.finished || fallbackSections.finished;
  const whyContent = whySangehassanContent[lang] || whySangehassanContent.fa;
  const teamContent = teamSectionContent[lang] || teamSectionContent.fa;
  const dynamicTeamCards = useMemo(
    () =>
      teamMembers.map((member) => ({
        id: member.id,
        name: pickField(member, "name", lang),
        role: pickField(member, "role", lang),
        bio: pickField(member, "bio", lang),
        photo: member.photo_url ? resolveImageUrl(member.photo_url) : "",
        linkedin: member.linkedin_url || ""
      })),
    [teamMembers, lang]
  );
  const teamCards = dynamicTeamCards.length ? dynamicTeamCards : teamContent.cards;
  const whySectionDir = lang === "en" ? "ltr" : "rtl";

  /*
  const renderDealNotification = (deal) => (
    <div
      dir="ltr"
      className="relative h-full overflow-hidden rounded-[1.15rem] border border-white/35 bg-white/[0.12] px-4 py-3 text-left text-sand shadow-[0_28px_75px_rgba(10,8,5,0.45)] backdrop-blur-[20px] sm:px-5 sm:py-3.5"
    >
      <span className="pointer-events-none absolute inset-0 bg-[linear-gradient(125deg,rgba(255,255,255,0.2)_0%,rgba(255,255,255,0.06)_38%,rgba(255,255,255,0)_100%)]" />
      <span className="pointer-events-none absolute -left-10 top-0 h-full w-16 rotate-[14deg] bg-white/15 blur-2xl" />
      <div className="relative">
        <div className="flex items-center justify-between gap-3">
          <div className="flex min-w-0 items-center gap-2">
            <img src={miniHeroMessagesIcon} alt="" className="h-5 w-5 rounded-[6px]" />
            <p className="truncate text-[9px] font-semibold uppercase tracking-[0.18em] text-sand/80 sm:text-[10px]">
              {liveDealsMessagesLabel}
            </p>
          </div>
          <p className="shrink-0 text-[10px] text-sand/70 sm:text-[11px]">{deal.time}</p>
        </div>
        <p className="mt-2 text-[12px] font-semibold text-sand sm:text-[13px]">{liveDealsSenderName}</p>
        <p className="mt-1 text-[11px] leading-snug text-sand/92 sm:text-[12px]">{renderDealMessage(deal, liveDealsConfig)}</p>
        <p className="mt-1.5 text-[10px] font-medium text-sand/68 sm:text-[11px]">{liveDealsMoreMessagesText}</p>
      </div>
    </div>
  );
  */

  useEffect(() => {
    const root = rootRef.current;
    if (!root || typeof window === "undefined" || !window.matchMedia) return;
    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    const items = root.querySelectorAll("[data-home-anim='item']");
    if (!items.length) return;

    const ctx = gsap.context(() => {
      gsap.fromTo(
        items,
        { autoAlpha: 0, y: 20 },
        {
          autoAlpha: 1,
          y: 0,
          duration: reduceMotion ? 0.45 : 1.2,
          delay: reduceMotion ? 0.1 : 0.5,
          stagger: reduceMotion ? 0.03 : 0.09,
          ease: "power3.out",
          overwrite: "auto"
        }
      );
    }, root);

    return () => ctx.revert();
  }, [lang, blocksSection, finishedSection]);

  return (
    <>
      <div ref={rootRef} className="relative h-[100dvh] w-full overflow-hidden">
        <section className="relative flex h-full w-full flex-col overflow-hidden lg:flex-row">
          {[finishedSection, blocksSection].map((section, index) => {
            const isBlocks = section.key === "blocks";
            const title = pickField(section, "title", lang) || (isBlocks ? t("blocks.title") : t("products.title"));
            const subtitle = pickField(section, "subtitle", lang);
            const description = pickField(section, "description", lang);
            const ctaLabel = pickField(section, "cta_label", lang);
            const fallbackCTA = isBlocks ? t("blocks.cta") : t("hero.ctaPrimary");
            const rawCtaHref = section.cta_href || (isBlocks ? "/blocks" : "/products");
            const ctaHref = isBlocks
              ? rawCtaHref === "/blocks/catalog" ? "/blocks" : rawCtaHref
              : `/${lang}/products`;
            const lines = splitLines(description);
            const slides = isBlocks ? slideDecks.blocks : slideDecks.products;
            const activeSlide = isBlocks ? activeSlides.blocks : activeSlides.products;
            const previousSlide = isBlocks ? previousSlides.blocks : previousSlides.products;
            const nextSlide = getNextSlideIndex(activeSlide, slides);
            const visibleSlideIndexes = [previousSlide, activeSlide, nextSlide].filter(
              (slideIndex, slideIndexPosition, slideIndexes) =>
                slideIndex !== null && slideIndexes.indexOf(slideIndex) === slideIndexPosition
            );

            return (
              <Link
                key={section.key || index}
                to={ctaHref}
                className="group relative block h-1/2 w-full overflow-hidden lg:h-full lg:w-1/2"
                aria-label={title}
              >
                <div className="absolute inset-0 z-0">
                  {visibleSlideIndexes.map((slideIndex) => {
                    const image = slides[slideIndex];
                    const isActiveSlide = slideIndex === activeSlide;
                    const isPreviousSlide = slideIndex === previousSlide;
                    const role = isActiveSlide ? "active" : isPreviousSlide ? "previous" : "preload";
                    const panel = isBlocks ? "blocks" : "products";
                    const shouldLoadEager = isActiveSlide;

                    return (
                      <img
                        key={image.src}
                        src={image.src}
                        alt=""
                        width={image.width}
                        height={image.height}
                        className={`landing-slide-layer absolute inset-0 h-full w-full object-cover object-center ${isActiveSlide
                          ? "landing-slide-active"
                          : isPreviousSlide
                            ? "landing-slide-previous"
                            : "landing-slide-preload"
                          }`}
                        data-slide-panel={panel}
                        data-slide-role={role}
                        data-slide-index={slideIndex}
                        loading={shouldLoadEager ? "eager" : "lazy"}
                        decoding="async"
                        fetchpriority={shouldLoadEager ? "high" : "low"}
                        {...getSlideDebugHandlers(panel, role, slideIndex, image)}
                      />
                    );
                  })}
                </div>

                <div className="absolute inset-0 z-[1] bg-primary/25" />
                <div
                  className={`absolute inset-0 z-[2] ${isBlocks
                    ? "bg-gradient-to-br from-primary/35 via-primary/15 to-primary/48"
                    : "bg-gradient-to-br from-primary/40 via-primary/20 to-primary/52"
                    }`}
                />
                <div className="absolute inset-0 z-[3] bg-[radial-gradient(circle_at_25%_20%,rgba(165,141,102,0.16),transparent_48%)]" />

                <div className="relative z-10 flex h-full items-center justify-center px-6 py-10 text-sand sm:px-10 lg:px-12">
                  <div className="mx-auto w-full max-w-[38rem] text-center">
                    <p data-home-anim="item" className="text-[10px] uppercase tracking-[0.34em] text-sand/72 sm:text-xs">
                      {isBlocks ? t("blocks.title") : t("products.title")}
                    </p>
                    <h1 data-home-anim="item" className="mt-2 font-display text-xl leading-tight sm:text-2xl md:text-3xl lg:text-4xl">
                      {title}
                    </h1>
                    {subtitle ? <p data-home-anim="item" className="mx-auto mt-3 max-w-[33rem] text-[11px] text-sand/88 sm:text-xs md:text-sm lg:text-base">{subtitle}</p> : null}

                    {lines.length > 0 && (
                      <ul className="mx-auto mt-4 max-w-[32rem] space-y-1.5 text-left text-[10px] text-sand/88 sm:text-[11px] md:text-xs lg:text-sm">
                        {lines.slice(0, 3).map((line) => (
                          <li key={line} data-home-anim="item" className="flex items-start gap-2">
                            <span className="mt-1.5 h-1.5 w-1.5 rounded-full bg-accent/85" />
                            <span>{line}</span>
                          </li>
                        ))}
                      </ul>
                    )}

                    <div data-home-anim="item" className="mt-7 flex items-center justify-center gap-3">
                      <span className="rounded-full border border-sand/40 bg-sand/10 px-4 py-2 text-[10px] font-semibold uppercase tracking-[0.24em] text-sand sm:px-5 sm:py-2.5 sm:text-xs">
                        {ctaLabel || fallbackCTA}
                      </span>
                      <span className="text-xl text-sand/75 transition-transform duration-500 group-hover:translate-x-1">
                        →
                      </span>
                    </div>
                  </div>
                </div>

                <span className="pointer-events-none absolute inset-0 ring-0 ring-inset transition group-hover:ring-1 group-hover:ring-white/25" />
              </Link>
            );
          })}
          <span className="pointer-events-none absolute inset-x-0 top-1/2 z-20 h-20 -translate-y-1/2 bg-gradient-to-b from-transparent via-primary/10 to-transparent backdrop-blur-sm lg:hidden" />
        </section>
        <a
          href="#home-continuation"
          aria-label="Scroll to next section"
          className="absolute bottom-5 left-1/2 z-30 flex h-11 w-11 -translate-x-1/2 items-center justify-center text-white drop-shadow-[0_6px_16px_rgba(0,0,0,0.45)] transition hover:-translate-y-0.5 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/70"
        >
          <ChevronDown className="h-9 w-9 animate-bounce" aria-hidden="true" strokeWidth={1.4} />
        </a>
      </div>

      <section
        id="home-continuation"
        className="relative isolate overflow-hidden bg-[#062f40] py-20 text-sand sm:py-24 lg:py-28"
        dir={whySectionDir}
        aria-labelledby="why-sangehassan-title"
      >
        <span className="bg-quiet-grid pointer-events-none absolute inset-0 opacity-[0.12]" aria-hidden="true" />
        <span className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(4,30,41,0.98),rgba(8,58,79,0.94)_52%,rgba(3,27,37,0.98))]" aria-hidden="true" />
        <span className="pointer-events-none absolute -left-[12%] top-[18%] h-[34rem] w-[34rem] rounded-full bg-sand/[0.08] blur-[120px]" aria-hidden="true" />
        <span className="pointer-events-none absolute -right-[14%] top-[42%] h-[38rem] w-[38rem] rounded-full bg-white/[0.06] blur-[140px]" aria-hidden="true" />

        <div className="section-shell relative z-10">
          <header className="mx-auto max-w-3xl text-center">
            <p className="text-xs font-semibold uppercase tracking-[0.28em] text-accent">{whyContent.eyebrow}</p>
            <h2 id="why-sangehassan-title" className="mt-4 font-display text-3xl leading-tight text-sand sm:text-4xl lg:text-5xl">
              {whyContent.title}
            </h2>
            <div className="mx-auto mt-5 max-w-2xl space-y-3 text-sm leading-8 text-sand/74 sm:text-base">
              <p>{whyContent.intro}</p>
              <p>{whyContent.support}</p>
            </div>
          </header>

          <div className="mx-auto mt-14 max-w-6xl">
            <div className="space-y-5">
              {whyContent.cards.map((card, index) => {
                const illustration = whySangehassanIllustrations[card.key];
                const imageFirst = index % 2 === 0;

                return (
                  <article
                    key={card.key}
                    className="grid min-h-[18rem] w-full grid-cols-[minmax(0,0.46fr)_minmax(0,0.54fr)] items-center sm:min-h-[22rem]"
                  >
                    <div className={`relative flex h-full min-h-[18rem] items-center justify-center px-3 py-8 sm:min-h-[22rem] sm:px-8 ${imageFirst ? "order-1" : "order-2"}`}>
                      <img
                        src={illustration.src}
                        alt=""
                        width={illustration.width}
                        height={illustration.height}
                        loading="lazy"
                        decoding="async"
                        className={`h-auto w-full object-contain opacity-80 drop-shadow-[0_14px_28px_rgba(0,0,0,0.18)] [filter:brightness(0.88)_contrast(0.9)_saturate(0.9)] ${card.key === "network"
                          ? "max-h-[18rem] max-w-[21rem] sm:max-h-[22rem] sm:max-w-[27rem]"
                          : "max-h-[15rem] max-w-[17rem] sm:max-h-[18rem] sm:max-w-[22rem]"
                          }`}
                        aria-hidden="true"
                      />
                    </div>
                    <div className={`relative flex h-full min-h-[18rem] flex-col items-center justify-center px-4 py-8 text-center sm:min-h-[22rem] sm:px-10 ${imageFirst ? "order-2" : "order-1"}`}>
                      <h3 className="font-display text-xl leading-tight text-white sm:text-3xl">{card.title}</h3>
                      <p className="mt-4 max-w-[34rem] text-xs leading-7 text-sand/78 sm:text-base sm:leading-8">
                        {card.summary}
                      </p>
                    </div>
                  </article>
                );
              })}
            </div>
          </div>

        </div>
      </section>

      <section className="relative isolate overflow-hidden bg-[#041f2b] py-20 text-sand sm:py-24 lg:py-28" dir={whySectionDir} aria-labelledby="home-team-title">
        <span className="bg-quiet-grid pointer-events-none absolute inset-0 opacity-[0.1]" aria-hidden="true" />
        <span className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(4,31,43,0.98),rgba(8,58,79,0.9)_48%,rgba(4,31,43,0.98))]" aria-hidden="true" />

        <div className="section-shell relative z-10">
          <header className="mx-auto max-w-3xl text-center">
            <p className="text-xs font-semibold uppercase tracking-[0.28em] text-accent">{teamContent.eyebrow}</p>
            <h2 id="home-team-title" className="mt-4 font-display text-3xl leading-tight text-sand sm:text-4xl lg:text-5xl">
              {teamContent.title}
            </h2>
            <p className="mx-auto mt-5 max-w-2xl text-sm leading-8 text-sand/74 sm:text-base">
              {teamContent.intro}
            </p>
          </header>

          <div className="mt-14 grid grid-cols-2 items-stretch gap-3 sm:gap-5 lg:grid-cols-4">
            {teamCards.map((member) => (
              <article
                key={member.id || member.name}
                className="group relative h-[24rem] overflow-hidden rounded-lg bg-white/[0.08] shadow-[0_28px_80px_rgba(0,0,0,0.26)] backdrop-blur-xl sm:h-[30rem]"
              >
                {member.photo ? (
                  <img src={member.photo} alt="" className="absolute inset-0 h-full w-full object-cover" aria-hidden="true" />
                ) : (
                  <div className="absolute inset-0 bg-[linear-gradient(145deg,rgba(229,225,221,0.22),rgba(8,58,79,0.18)_36%,rgba(0,0,0,0.42))]" aria-hidden="true" />
                )}
                <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(255,255,255,0.07),rgba(0,0,0,0.08)_42%,rgba(0,0,0,0.72))]" aria-hidden="true" />
                <div className="absolute inset-x-0 bottom-0 flex h-[40%] flex-col justify-center overflow-hidden bg-[linear-gradient(180deg,rgba(3,24,33,0),rgba(3,24,33,0.72)_18%,rgba(3,24,33,0.95))] px-3 py-4 text-center sm:px-6 sm:py-6">
                  <h3 className="line-clamp-2 font-display text-base leading-6 text-white sm:text-xl sm:leading-8">{member.name}</h3>
                  <p className="mt-1 line-clamp-1 text-[9px] font-semibold uppercase tracking-[0.14em] text-accent/88 sm:text-[11px] sm:tracking-[0.2em]">{member.role}</p>
                  <p className="mt-3 line-clamp-4 text-[11px] leading-5 text-white/72 sm:mt-4 sm:line-clamp-3 sm:text-sm sm:leading-7">{member.bio}</p>
                  {member.linkedin ? (
                    <a
                      href={member.linkedin}
                      target="_blank"
                      rel="noreferrer"
                      className="mt-3 inline-flex items-center justify-center text-xs font-semibold text-white transition hover:text-accent sm:mt-4 sm:text-sm"
                    >
                      {teamContent.linkedinLabel} →
                    </a>
                  ) : null}
                </div>
              </article>
            ))}
          </div>
        </div>
      </section>
    </>
  );
}
