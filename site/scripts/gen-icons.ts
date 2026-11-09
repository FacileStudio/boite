import { writeFileSync } from "node:fs";
import data from "@iconify-json/solar/icons.json";
import { ICON_NAMES } from "../src/lib/icon-names";

const solar = data as { width: number; height: number; icons: Record<string, { body: string }> };

function build(): Record<string, { width: number; height: number; body: string }> {
  const out: Record<string, { width: number; height: number; body: string }> = {};
  for (const name of ICON_NAMES) {
    const icon = solar.icons[name.replace("solar:", "")];
    if (!icon) throw new Error(`missing solar icon ${name}`);
    out[name] = { width: solar.width, height: solar.height, body: icon.body };
  }
  return out;
}

writeFileSync(new URL("../src/lib/icons-data.json", import.meta.url), `${JSON.stringify(build(), null, 2)}\n`);
