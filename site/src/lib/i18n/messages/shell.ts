import type { Row } from "../row";

/** Page titles, meta, header, footer and theme strings. */
export const shell = {
  "meta.title": [
    "Boîte · Break things safely",
    "Boîte · Cassez tout, sans risque",
    "Boîte · Rómpelo todo, sin riesgo",
    "Boîte · Alles kaputt machen, ohne Risiko",
  ],
  "meta.description": [
    "A throwaway machine for your code and your AI agents. Take a snapshot, break everything, roll back in one command.",
    "Une machine jetable pour votre code et vos agents IA. Prenez un instantané, cassez tout, revenez en arrière en une commande.",
    "Una máquina desechable para tu código y tus agentes de IA. Haz una instantánea, rómpelo todo y vuelve atrás con un comando.",
    "Eine Wegwerf-Maschine für Ihren Code und Ihre KI-Agenten. Snapshot machen, alles kaputt machen, mit einem Befehl zurück.",
  ],
  "meta.downloads": ["Downloads · Boîte", "Téléchargements · Boîte", "Descargas · Boîte", "Downloads · Boîte"],
  "nav.demos": ["How it works", "Comment ça marche", "Cómo funciona", "So funktioniert es"],
  "nav.pricing": ["Pricing", "Tarifs", "Precios", "Preise"],
  "nav.install": ["Install", "Installer", "Instalar", "Installieren"],
  "nav.downloads": ["Downloads", "Téléchargements", "Descargas", "Downloads"],
  "footer.tagline": [
    "Throwaway machines for your code and your AI agents.",
    "Des machines jetables pour votre code et vos agents IA.",
    "Máquinas desechables para tu código y tus agentes de IA.",
    "Wegwerf-Maschinen für Ihren Code und Ihre KI-Agenten.",
  ],
  "footer.suite": ["Facile suite", "Suite Facile", "Suite Facile", "Facile-Suite"],
  "footer.source": ["Open source", "Code ouvert", "Código abierto", "Open Source"],
  "footer.repo": ["GitHub repository", "Dépôt GitHub", "Repositorio de GitHub", "GitHub-Repository"],
  "footer.script": ["Install script", "Script d'installation", "Script de instalación", "Installationsskript"],
  "footer.license": ["GPL-3.0 license", "Licence GPL-3.0", "Licencia GPL-3.0", "GPL-3.0-Lizenz"],
  "theme.toLight": [
    "Switch to light theme",
    "Passer au thème clair",
    "Cambiar al tema claro",
    "Zum hellen Design wechseln",
  ],
  "theme.toDark": [
    "Switch to dark theme",
    "Passer au thème sombre",
    "Cambiar al tema oscuro",
    "Zum dunklen Design wechseln",
  ],
} satisfies Record<string, Row>;
