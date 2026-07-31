import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import sharp from "sharp";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const assetsDir = path.resolve(__dirname, "../../shared/assets");

const variants = [
  ...["finish-slide-01", "finish-slide-02", "finish-slide-03", "finish-slide-04", "product-slide-01", "product-slide-02", "product-slide-03"]
    .map((name) => ({ source: `landing_page/products/${name}.webp`, output: `landing_page/products/${name}-mobile.webp`, width: 480, quality: 70 })),
  ...Array.from({ length: 8 }, (_, index) => String(index + 1).padStart(2, "0"))
    .map((number) => ({ source: `landing_page/blocks/block-slide-${number}.webp`, output: `landing_page/blocks/block-slide-${number}-mobile.webp`, width: 480, quality: 70 })),
  ...["market_complexity_icon_transparent", "network_supply_icon_transparent", "trust_quality_icon_transparent"]
    .map((name) => ({ source: `landing_icons/${name}.webp`, output: `landing_icons/${name}-mobile.webp`, width: 400, quality: 76 })),
  ...["tradeboard-browse-listings", "tradeboard-post-listing", "tradeboard-secure-review"]
    .map((name) => ({ source: `tradeboard/${name}.webp`, output: `tradeboard/${name}-mobile.webp`, width: 320, quality: 76 })),
  { source: "logo.png", output: "logo-240.webp", width: 240, quality: 82 },
  { source: "logo_white.png", output: "logo-white-240.webp", width: 240, quality: 82 }
];

for (const variant of variants) {
  const input = path.join(assetsDir, variant.source);
  const output = path.join(assetsDir, variant.output);

  await sharp(input)
    .resize({ width: variant.width, withoutEnlargement: true })
    .webp({ quality: variant.quality, effort: 6 })
    .toFile(output);

  const [metadata, stats] = await Promise.all([sharp(output).metadata(), fs.stat(output)]);
  console.log(`${variant.output} ${metadata.width}x${metadata.height} ${Math.round(stats.size / 1024)} KiB`);
}
