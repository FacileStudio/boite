import type { Catalogs, Message } from "@facile/muse";
import type { Row } from "../row";
import { cards } from "./cards";
import { demos } from "./demos";
import { hero } from "./hero";
import { sections } from "./sections";
import { shell } from "./shell";

const rows = { ...shell, ...hero, ...cards, ...demos, ...sections };

export type SiteKey = keyof typeof rows;
type SiteCatalog = Record<SiteKey, Message>;

function column(index: number): SiteCatalog {
  const entries = Object.entries(rows as Record<string, Row>).map(([key, row]) => [key, row[index]]);
  return Object.fromEntries(entries) as SiteCatalog;
}

/** The site's catalogs: English is the source, and every row carries all four locales. */
export const catalogs: Catalogs<SiteCatalog> = {
  en: column(0),
  fr: column(1),
  es: column(2),
  de: column(3),
};
