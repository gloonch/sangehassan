import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import sharp from "sharp";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const iconDir = path.resolve(__dirname, "../../shared/assets/landing_icons");
const icons = [
  "market_complexity_icon_transparent",
  "network_supply_icon_transparent",
  "trust_quality_icon_transparent"
];

await fs.mkdir(iconDir, { recursive: true });

for (const name of icons) {
  const input = path.join(iconDir, `${name}.png`);
  const output = path.join(iconDir, `${name}.webp`);

  await sharp(input)
    .resize({ width: 768, height: 768, fit: "inside", withoutEnlargement: true })
    .webp({ quality: 82, effort: 6 })
    .toFile(output);

  const [metadata, stats] = await Promise.all([sharp(output).metadata(), fs.stat(output)]);
  console.log(`${path.relative(process.cwd(), output)} ${metadata.width}x${metadata.height} ${Math.round(stats.size / 1024)} KiB`);
}
