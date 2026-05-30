const notificationTimes = [
  "35 minutes ago",
  "45 minutes ago",
  "1 hour ago",
  "2 hours ago",
  "3 hours ago",
  "4 hours ago",
  "5 hours ago",
  "6 hours ago",
  "7 hours ago",
  "8 hours ago",
  "9 hours ago",
  "10 hours ago",
  "11 hours ago",
  "12 hours ago",
  "14 hours ago",
  "16 hours ago",
  "18 hours ago",
  "20 hours ago",
  "22 hours ago",
  "1 day ago",
  "2 days ago",
  "3 days ago",
  "4 days ago",
  "5 days ago"
];

const buildDeals = ({
  slabProducts,
  blockProducts,
  accessoryProducts,
  slabAmounts,
  blockAmounts,
  accessoryAmounts
}) => {
  const deals = [];
  let timeCursor = 0;

  const appendGroup = (products, amounts) => {
    products.forEach((product) => {
      amounts.forEach((amount) => {
        deals.push({
          product,
          amount,
          time: notificationTimes[timeCursor % notificationTimes.length]
        });
        timeCursor += 1;
      });
    });
  };

  appendGroup(slabProducts, slabAmounts);
  appendGroup(blockProducts, blockAmounts);
  appendGroup(accessoryProducts, accessoryAmounts);

  return deals;
};

export const pickRandomIndex = (length) => (length > 0 ? Math.floor(Math.random() * length) : 0);

export const liveDealsByLang = {
  fa: {
    messagesLabel: "MESSAGES",
    senderName: "SangeHassan",
    messagePrefix: "معامله انجام شد",
    moreMessagesText: "28 more messages",
    deals: buildDeals({
      slabProducts: [
        "اسلب گرانیت",
        "اسلب تراورتن",
        "اسلب مرمریت",
        "اسلب مرمر",
        "اسلب چینی کریستال",
        "اسلب گرانودیوریت"
      ],
      blockProducts: [
        "کوپ گرانیت",
        "کوپ تراورتن",
        "کوپ مرمریت",
        "کوپ مرمر",
        "کوپ چینی کریستال",
        "کوپ گرانودیوریت"
      ],
      accessoryProducts: [
        "اکسسوری گرانیت",
        "اکسسوری تراورتن",
        "اکسسوری مرمریت",
        "اکسسوری مرمر",
        "اکسسوری چینی کریستال",
        "اکسسوری گرانودیوریت"
      ],
      slabAmounts: ["۱۰ متر", "۱۲ متر", "۱۸ متر", "۲۴ متر", "۳۲ متر", "۴۸ متر", "۶۴ متر"],
      blockAmounts: ["۱ کوپ", "۲ کوپ", "۳ کوپ", "۴ کوپ", "۵ کوپ"],
      accessoryAmounts: ["۴ عدد", "۸ عدد", "۱۲ عدد", "۱۶ عدد", "۲۰ عدد", "۲۴ عدد"]
    })
  },
  en: {
    messagesLabel: "MESSAGES",
    senderName: "SangeHassan",
    messagePrefix: "Deal completed",
    moreMessagesText: "28 more messages",
    deals: buildDeals({
      slabProducts: [
        "Granite Slab",
        "Travertine Slab",
        "Marmarite Slab",
        "Marmar Slab",
        "Chinese Crystal Slab",
        "Granodiorite Slab"
      ],
      blockProducts: [
        "Granite Block",
        "Travertine Block",
        "Marmarite Block",
        "Marmar Block",
        "Chinese Crystal Block",
        "Granodiorite Block"
      ],
      accessoryProducts: [
        "Granite Accessory",
        "Travertine Accessory",
        "Marmarite Accessory",
        "Marmar Accessory",
        "Chinese Crystal Accessory",
        "Granodiorite Accessory"
      ],
      slabAmounts: ["10 sqm", "12 sqm", "18 sqm", "24 sqm", "32 sqm", "48 sqm", "64 sqm"],
      blockAmounts: ["1 block", "2 blocks", "3 blocks", "4 blocks", "5 blocks"],
      accessoryAmounts: ["4 units", "8 units", "12 units", "16 units", "20 units", "24 units"]
    })
  },
  ar: {
    messagesLabel: "MESSAGES",
    senderName: "SangeHassan",
    messagePrefix: "تمت الصفقة",
    moreMessagesText: "28 more messages",
    deals: buildDeals({
      slabProducts: [
        "ألواح جرانيت",
        "ألواح ترافرتين",
        "ألواح مرمريت",
        "ألواح مرمر",
        "ألواح كريستال صيني",
        "ألواح جرانوديوريت"
      ],
      blockProducts: [
        "بلوك جرانيت",
        "بلوك ترافرتين",
        "بلوك مرمريت",
        "بلوك مرمر",
        "بلوك كريستال صيني",
        "بلوك جرانوديوريت"
      ],
      accessoryProducts: [
        "إكسسوار جرانيت",
        "إكسسوار ترافرتين",
        "إكسسوار مرمريت",
        "إكسسوار مرمر",
        "إكسسوار كريستال صيني",
        "إكسسوار جرانوديوريت"
      ],
      slabAmounts: ["١٠ متر", "١٢ متر", "١٨ متر", "٢٤ متر", "٣٢ متر", "٤٨ متر", "٦٤ متر"],
      blockAmounts: ["١ بلوك", "٢ بلوك", "٣ بلوك", "٤ بلوك", "٥ بلوك"],
      accessoryAmounts: ["٤ قطع", "٨ قطع", "١٢ قطع", "١٦ قطع", "٢٠ قطع", "٢٤ قطع"]
    })
  }
};

export const getLiveDealsConfig = (lang) => liveDealsByLang[lang] || liveDealsByLang.fa;

export const renderDealMessage = (deal, config) => `${config.messagePrefix}: ${deal.amount} | ${deal.product}`;
