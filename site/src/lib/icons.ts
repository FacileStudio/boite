import { registerIcons, type IconData } from "@facile/muse";
import data from "./icons-data.json";

export const LOGO = "solar:box-bold-duotone";
export const GITHUB = "solar:code-square-linear";
export const MENU = "solar:hamburger-menu-linear";

export function registerSiteIcons(): void {
  registerIcons(data as Record<string, IconData>);
}
