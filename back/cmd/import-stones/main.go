package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"

	"sangehassan/back/internal/adapters/persistence/postgres"
	"sangehassan/back/internal/config"
)

type dataset struct {
	Products []productRecord `json:"products"`
}

type productRecord struct {
	Category string   `json:"category"`
	NameFA   string   `json:"name_fa"`
	Aliases  []string `json:"aliases"`
	Variants []string `json:"variants"`
	Mines    []string `json:"mines"`
	Finishes []string `json:"finishes"`
}

type counts struct {
	Total int64
	ByCat map[string]int64
}

func main() {
	dataPath := flag.String("data", "./data/stones.json", "path to stones dataset JSON")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	raw, err := os.ReadFile(*dataPath)
	if err != nil {
		log.Fatalf("read data file: %v", err)
	}
	var payload dataset
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Fatalf("parse data file: %v", err)
	}
	if len(payload.Products) == 0 {
		log.Fatalf("no products found in dataset")
	}

	db, err := postgres.NewDB(cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		log.Fatalf("start tx: %v", err)
	}

	categoryIDs := make(map[string]int64)
	invalid := 0

	for _, rec := range payload.Products {
		catName := strings.TrimSpace(rec.Category)
		nameFA := strings.TrimSpace(rec.NameFA)
		if catName == "" || nameFA == "" {
			invalid++
			continue
		}

		catID, ok := categoryIDs[catName]
		if !ok {
			id, err := ensureCategory(ctx, tx, catName)
			if err != nil {
				_ = tx.Rollback()
				log.Fatalf("ensure category %s: %v", catName, err)
			}
			catID = id
			categoryIDs[catName] = id
		}

		productSlug := slugify(nameFA)
		aliases := normalizeList(rec.Aliases)
		variants := normalizeList(rec.Variants)
		mines := normalizeList(rec.Mines)
		finishes := normalizeList(rec.Finishes)

		var productID int64
		productID, err = findProductID(ctx, tx, catID, nameFA)
		if err != nil {
			_ = tx.Rollback()
			log.Fatalf("lookup product %s: %v", nameFA, err)
		}

		if productID > 0 {
			err = tx.QueryRowContext(ctx, `
				UPDATE products SET
				  title_en = $1,
				  title_fa = $2,
				  title_ar = $3,
				  slug = $4,
				  aliases = $5,
				  variants = $6,
				  mines = $7,
				  finishes = $8,
				  main_category_id = $9,
				  updated_at = NOW()
				WHERE id = $10
				RETURNING id
			`,
				nameFA, nameFA, nameFA, productSlug,
				pq.Array(aliases), pq.Array(variants), pq.Array(mines), pq.Array(finishes),
				catID, productID,
			).Scan(&productID)
			if err != nil {
				_ = tx.Rollback()
				log.Fatalf("update product %s: %v", nameFA, err)
			}
		} else {
			err = tx.QueryRowContext(ctx, `
				INSERT INTO products (
				  title_en, title_fa, title_ar, slug,
				  aliases, variants, mines, finishes,
				  main_category_id
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (slug) DO UPDATE SET
				  title_en = EXCLUDED.title_en,
				  title_fa = EXCLUDED.title_fa,
				  title_ar = EXCLUDED.title_ar,
				  aliases = EXCLUDED.aliases,
				  variants = EXCLUDED.variants,
				  mines = EXCLUDED.mines,
				  finishes = EXCLUDED.finishes,
				  main_category_id = EXCLUDED.main_category_id,
				  updated_at = NOW()
				RETURNING id
			`,
				nameFA, nameFA, nameFA, productSlug,
				pq.Array(aliases), pq.Array(variants), pq.Array(mines), pq.Array(finishes),
				catID,
			).Scan(&productID)
			if err != nil {
				_ = tx.Rollback()
				log.Fatalf("upsert product %s: %v", nameFA, err)
			}
		}

		if err := ensureProductCategory(ctx, tx, productID, catID); err != nil {
			_ = tx.Rollback()
			log.Fatalf("link product category %s: %v", nameFA, err)
		}
	}

	stats, err := loadCounts(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("load counts: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit: %v", err)
	}

	fmt.Printf("Import done at %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Total products: %d\n", stats.Total)
	for cat, count := range stats.ByCat {
		fmt.Printf("- %s: %d\n", cat, count)
	}
	if invalid > 0 {
		fmt.Printf("Skipped invalid rows: %d\n", invalid)
	} else {
		fmt.Printf("Skipped invalid rows: 0\n")
	}
}

func ensureCategory(ctx context.Context, tx *sql.Tx, titleFA string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM categories WHERE lower(title_fa) = lower($1) LIMIT 1
	`, titleFA).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	slug := slugify(titleFA)
	err = tx.QueryRowContext(ctx, `
		INSERT INTO categories (title_en, title_fa, title_ar, slug)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (slug) DO UPDATE SET
		  title_en = EXCLUDED.title_en,
		  title_fa = EXCLUDED.title_fa,
		  title_ar = EXCLUDED.title_ar,
		  updated_at = NOW()
		RETURNING id
	`, titleFA, titleFA, titleFA, slug).Scan(&id)
	return id, err
}

func ensureProductCategory(ctx context.Context, tx *sql.Tx, productID, categoryID int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO product_categories (product_id, category_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, productID, categoryID)
	return err
}

func findProductID(ctx context.Context, tx *sql.Tx, categoryID int64, titleFA string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM products
		WHERE main_category_id = $1 AND lower(title_fa) = lower($2)
		LIMIT 1
	`, categoryID, titleFA).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func loadCounts(ctx context.Context, tx *sql.Tx) (counts, error) {
	var result counts
	result.ByCat = make(map[string]int64)

	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM products`).Scan(&result.Total); err != nil {
		return result, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT COALESCE(c.title_fa, 'بدون دسته‌بندی') AS cat, COUNT(*)
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		GROUP BY cat
		ORDER BY cat
	`)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var count int64
		if err := rows.Scan(&name, &count); err != nil {
			return result, err
		}
		result.ByCat[name] = count
	}
	return result, rows.Err()
}

func normalizeList(values []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(values))
	for _, v := range values {
		clean := strings.TrimSpace(v)
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

var translitMap = map[rune]string{
	'ا': "a", 'آ': "a", 'أ': "a", 'إ': "e",
	'ب': "b",
	'پ': "p",
	'ت': "t",
	'ث': "s",
	'ج': "j",
	'چ': "ch",
	'ح': "h",
	'خ': "kh",
	'د': "d",
	'ذ': "z",
	'ر': "r",
	'ز': "z",
	'ژ': "zh",
	'س': "s",
	'ش': "sh",
	'ص': "s",
	'ض': "z",
	'ط': "t",
	'ظ': "z",
	'ع': "a",
	'غ': "gh",
	'ف': "f",
	'ق': "gh",
	'ك': "k", 'ک': "k",
	'گ': "g",
	'ل': "l",
	'م': "m",
	'ن': "n",
	'و': "v",
	'ؤ': "o",
	'ه': "h", 'ة': "h", 'ۀ': "e",
	'ی': "y", 'ي': "y", 'ئ': "y",
	'ء': "",
	'۰': "0", '۱': "1", '۲': "2", '۳': "3", '۴': "4", '۵': "5", '۶': "6", '۷': "7", '۸': "8", '۹': "9",
}

func slugify(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	for _, r := range value {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			builder.WriteRune(r)
		case r == ' ' || r == '_' || r == '-' || r == 'ـ':
			builder.WriteRune('-')
		case translitMap[r] != "":
			builder.WriteString(translitMap[r])
		default:
			builder.WriteRune('-')
		}
	}

	value = slugRegex.ReplaceAllString(builder.String(), "-")
	value = strings.Trim(value, "-")
	return value
}
