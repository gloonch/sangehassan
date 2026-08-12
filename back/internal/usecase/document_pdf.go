package usecase

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/signintech/gopdf"
)

//go:embed assets/Vazirmatn-Regular.ttf
var vazirmatnRegular []byte

type arabicGlyph struct {
	isolated, final, initial, medial rune
	joinsLeft, joinsRight            bool
}

var arabicGlyphs = map[rune]arabicGlyph{
	'ء': {'ﺀ', 'ﺀ', 'ﺀ', 'ﺀ', false, false}, 'آ': {'ﺁ', 'ﺂ', 'ﺁ', 'ﺂ', false, true}, 'ا': {'ﺍ', 'ﺎ', 'ﺍ', 'ﺎ', false, true},
	'ب': {'ﺏ', 'ﺐ', 'ﺑ', 'ﺒ', true, true}, 'پ': {'ﭖ', 'ﭗ', 'ﭘ', 'ﭙ', true, true}, 'ت': {'ﺕ', 'ﺖ', 'ﺗ', 'ﺘ', true, true}, 'ث': {'ﺙ', 'ﺚ', 'ﺛ', 'ﺜ', true, true},
	'ج': {'ﺝ', 'ﺞ', 'ﺟ', 'ﺠ', true, true}, 'چ': {'ﭺ', 'ﭻ', 'ﭼ', 'ﭽ', true, true}, 'ح': {'ﺡ', 'ﺢ', 'ﺣ', 'ﺤ', true, true}, 'خ': {'ﺥ', 'ﺦ', 'ﺧ', 'ﺨ', true, true},
	'د': {'ﺩ', 'ﺪ', 'ﺩ', 'ﺪ', false, true}, 'ذ': {'ﺫ', 'ﺬ', 'ﺫ', 'ﺬ', false, true}, 'ر': {'ﺭ', 'ﺮ', 'ﺭ', 'ﺮ', false, true}, 'ز': {'ﺯ', 'ﺰ', 'ﺯ', 'ﺰ', false, true},
	'ژ': {'ﮊ', 'ﮋ', 'ﮊ', 'ﮋ', false, true}, 'س': {'ﺱ', 'ﺲ', 'ﺳ', 'ﺴ', true, true}, 'ش': {'ﺵ', 'ﺶ', 'ﺷ', 'ﺸ', true, true}, 'ص': {'ﺹ', 'ﺺ', 'ﺻ', 'ﺼ', true, true},
	'ض': {'ﺽ', 'ﺾ', 'ﺿ', 'ﻀ', true, true}, 'ط': {'ﻁ', 'ﻂ', 'ﻃ', 'ﻄ', true, true}, 'ظ': {'ﻅ', 'ﻆ', 'ﻇ', 'ﻈ', true, true}, 'ع': {'ﻉ', 'ﻊ', 'ﻋ', 'ﻌ', true, true},
	'غ': {'ﻍ', 'ﻎ', 'ﻏ', 'ﻐ', true, true}, 'ف': {'ﻑ', 'ﻒ', 'ﻓ', 'ﻔ', true, true}, 'ق': {'ﻕ', 'ﻖ', 'ﻗ', 'ﻘ', true, true}, 'ک': {'ﮎ', 'ﮏ', 'ﮐ', 'ﮑ', true, true},
	'ك': {'ﻙ', 'ﻚ', 'ﻛ', 'ﻜ', true, true}, 'گ': {'ﮒ', 'ﮓ', 'ﮔ', 'ﮕ', true, true}, 'ل': {'ﻝ', 'ﻞ', 'ﻟ', 'ﻠ', true, true}, 'م': {'ﻡ', 'ﻢ', 'ﻣ', 'ﻤ', true, true},
	'ن': {'ﻥ', 'ﻦ', 'ﻧ', 'ﻨ', true, true}, 'و': {'ﻭ', 'ﻮ', 'ﻭ', 'ﻮ', false, true}, 'ه': {'ﻩ', 'ﻪ', 'ﻫ', 'ﻬ', true, true}, 'ة': {'ﺓ', 'ﺔ', 'ﺓ', 'ﺔ', false, true},
	'ی': {'ﯼ', 'ﯽ', 'ﯾ', 'ﯿ', true, true}, 'ي': {'ﻱ', 'ﻲ', 'ﻳ', 'ﻴ', true, true}, 'ى': {'ﻯ', 'ﻰ', 'ﻯ', 'ﻰ', false, true}, 'ئ': {'ﺉ', 'ﺊ', 'ﺋ', 'ﺌ', true, true},
}

func rtlPersian(input string) string {
	r := []rune(strings.TrimSpace(input))
	shaped := make([]rune, len(r))
	for i, ch := range r {
		g, ok := arabicGlyphs[ch]
		if !ok {
			shaped[i] = ch
			continue
		}
		prevJoin := false
		nextJoin := false
		if i > 0 {
			if pg, pok := arabicGlyphs[r[i-1]]; pok {
				prevJoin = pg.joinsLeft && g.joinsRight
			}
		}
		if i+1 < len(r) {
			if ng, nok := arabicGlyphs[r[i+1]]; nok {
				nextJoin = g.joinsLeft && ng.joinsRight
			}
		}
		switch {
		case prevJoin && nextJoin:
			shaped[i] = g.medial
		case prevJoin:
			shaped[i] = g.final
		case nextJoin:
			shaped[i] = g.initial
		default:
			shaped[i] = g.isolated
		}
	}
	// Reverse RTL text while retaining the order of Latin/numeric runs.
	for l, h := 0, len(shaped)-1; l < h; l, h = l+1, h-1 {
		shaped[l], shaped[h] = shaped[h], shaped[l]
	}
	for i := 0; i < len(shaped); {
		if unicode.Is(unicode.Arabic, shaped[i]) || unicode.IsSpace(shaped[i]) {
			i++
			continue
		}
		j := i
		for j < len(shaped) && !unicode.Is(unicode.Arabic, shaped[j]) && !unicode.IsSpace(shaped[j]) {
			j++
		}
		for l, h := i, j-1; l < h; l, h = l+1, h-1 {
			shaped[l], shaped[h] = shaped[h], shaped[l]
		}
		i = j
	}
	return string(shaped)
}

func generatePersianPDF(title string, snapshot map[string]any) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	if err := pdf.AddTTFFontData("Vazirmatn", vazirmatnRegular); err != nil {
		return nil, err
	}
	pdf.AddPage()
	if err := pdf.SetFont("Vazirmatn", "", 18); err != nil {
		return nil, err
	}
	pdf.SetX(36)
	pdf.SetY(36)
	if err := pdf.CellWithOption(&gopdf.Rect{W: 523, H: 32}, rtlPersian(title), gopdf.CellOption{Align: gopdf.Right | gopdf.Middle, Border: gopdf.Bottom}); err != nil {
		return nil, err
	}
	if err := pdf.SetFont("Vazirmatn", "", 11); err != nil {
		return nil, err
	}
	y := 82.0
	writeLine := func(line string) error {
		if y > 770 {
			pdf.AddPage()
			y = 36
		}
		pdf.SetX(36)
		pdf.SetY(y)
		if err := pdf.CellWithOption(&gopdf.Rect{W: 523, H: 24}, rtlPersian(line), gopdf.CellOption{Align: gopdf.Right | gopdf.Middle, Border: gopdf.Bottom}); err != nil {
			return err
		}
		y += 27
		return nil
	}
	keys := make([]string, 0, len(snapshot))
	for k := range snapshot {
		if k != "items" && k != "packages" && k != "containers" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		value := fmt.Sprint(snapshot[k])
		if len([]rune(value)) > 90 {
			value = string([]rune(value)[:90]) + "…"
		}
		line := fmt.Sprintf("%s : %s", documentLabel(k), value)
		if err := writeLine(line); err != nil {
			return nil, err
		}
	}
	for _, section := range []struct {
		key, title string
	}{
		{"items", "اقلام سفارش"}, {"packages", "بسته‌ها"}, {"containers", "کانتینرها"},
	} {
		rows, ok := snapshot[section.key].([]map[string]any)
		if !ok || len(rows) == 0 {
			continue
		}
		if err := writeLine(section.title); err != nil {
			return nil, err
		}
		for index, row := range rows {
			rowKeys := make([]string, 0, len(row))
			for key := range row {
				rowKeys = append(rowKeys, key)
			}
			sort.Strings(rowKeys)
			parts := []string{fmt.Sprintf("%d", index+1)}
			for _, key := range rowKeys {
				if row[key] != nil && fmt.Sprint(row[key]) != "<nil>" {
					parts = append(parts, fmt.Sprintf("%s: %v", documentLabel(key), row[key]))
				}
			}
			line := strings.Join(parts, " | ")
			if len([]rune(line)) > 140 {
				line = string([]rune(line)[:140]) + "…"
			}
			if err := writeLine(line); err != nil {
				return nil, err
			}
		}
	}
	return pdf.GetBytesPdfReturnErr()
}

func documentLabel(key string) string {
	labels := map[string]string{"document_number": "شماره سند", "proforma_number": "شماره پیش‌فاکتور", "order_number": "شماره سفارش", "customer_name": "مشتری", "status": "وضعیت", "currency": "ارز", "subtotal": "جمع", "discount": "تخفیف", "tax": "مالیات", "charges": "هزینه‌های اضافی", "total": "مبلغ نهایی", "issued_at": "تاریخ صدور", "estimated_delivery_at": "تحویل تقریبی", "payment_terms": "شرایط پرداخت", "delivery_terms": "شرایط تحویل", "payment_number": "شماره پرداخت", "amount": "مبلغ", "paid_at": "تاریخ پرداخت", "reference": "شماره پیگیری", "shipment_number": "شماره محموله", "receiver_name": "تحویل‌گیرنده", "delivered_at": "تاریخ تحویل", "description": "شرح", "quantity": "مقدار", "unit": "واحد", "unit_price": "قیمت واحد", "line_amount": "مبلغ ردیف", "package_number": "شماره بسته", "gross_weight": "وزن ناخالص", "net_weight": "وزن خالص", "weight_unit": "واحد وزن", "container_number": "شماره کانتینر", "container_type": "نوع کانتینر", "seal_number": "شماره پلمب"}
	if v := labels[key]; v != "" {
		return v
	}
	return strings.ReplaceAll(key, "_", " ")
}
