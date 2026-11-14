import type { Row } from "../row";

/** The hero: slogan, subtitle, buttons and the terminal demo label. */
export const hero = {
  "hero.title": [
    "Break things safely.",
    "Cassez tout, sans risque.",
    "Rómpelo todo, sin riesgo.",
    "Alles kaputt machen, ohne Risiko.",
  ],
  "hero.subtitle": [
    "A throwaway machine for your code and your AI agents. Take a snapshot, break everything, roll back in one command.",
    "Une machine jetable pour votre code et vos agents IA. Prenez un instantané, cassez tout, revenez en arrière en une commande.",
    "Una máquina desechable para tu código y tus agentes de IA. Haz una instantánea, rómpelo todo y vuelve atrás con un comando.",
    "Eine Wegwerf-Maschine für Ihren Code und Ihre KI-Agenten. Snapshot machen, alles kaputt machen, mit einem Befehl zurück.",
  ],
  "hero.install": ["Install for free", "Installer gratuitement", "Instalar gratis", "Kostenlos installieren"],
  "hero.bulle": ["Try it in Bulle", "Essayer dans Bulle", "Probar en Bulle", "In Bulle testen"],
  "hero.note": [
    "Free and open source, for Linux. No Linux? Bulle's agents already work in a Boîte, with the Pro plan.",
    "Gratuit et open source, pour Linux. Pas de Linux ? Les agents de Bulle travaillent déjà dans une Boîte, avec l'offre Pro.",
    "Gratis y de código abierto, para Linux. ¿Sin Linux? Los agentes de Bulle ya trabajan en una Boîte, con el plan Pro.",
    "Kostenlos und Open Source, für Linux. Kein Linux? Die Agenten von Bulle arbeiten schon in einer Boîte, mit dem Pro-Tarif.",
  ],
  "hero.demoLabel": [
    "A terminal creates a Boîte machine, takes a snapshot, deletes /usr, then rolls back.",
    "Un terminal crée une machine Boîte, prend un instantané, supprime /usr, puis revient en arrière.",
    "Un terminal crea una máquina Boîte, hace una instantánea, borra /usr y vuelve atrás.",
    "Ein Terminal erstellt eine Boîte-Maschine, macht einen Snapshot, löscht /usr und rollt zurück.",
  ],
} satisfies Record<string, Row>;
