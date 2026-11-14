import type { Row } from "../row";

/** The three demo tiles under the hero. */
export const cards = {
  "cards.title": [
    "A machine of its own, ready to break",
    "Une machine à part, prête à casser",
    "Una máquina aparte, lista para romper",
    "Eine eigene Maschine, bereit zum Kaputtmachen",
  ],
  "cards.agent.title": [
    "Your AI agent works in its own machine, never on yours.",
    "Votre agent IA travaille dans sa propre machine, jamais sur la vôtre.",
    "Tu agente de IA trabaja en su propia máquina, nunca en la tuya.",
    "Ihr KI-Agent arbeitet in seiner eigenen Maschine, nie auf Ihrer.",
  ],
  "cards.agent.body": [
    "Bulle installs and tests in the machine, then reports back.",
    "Bulle installe et teste dans la machine, puis vous donne le résultat.",
    "Bulle instala y prueba en la máquina, y luego te da el resultado.",
    "Bulle installiert und testet in der Maschine und meldet das Ergebnis.",
  ],
  "cards.rollback.title": [
    "A snapshot, a mistake, a rollback.",
    "Un instantané, une erreur, un retour en arrière.",
    "Una instantánea, un error, una vuelta atrás.",
    "Ein Snapshot, ein Fehler, ein Rollback.",
  ],
  "cards.rollback.body": [
    "Delete /usr if you like. One command brings the machine back.",
    "Supprimez /usr si ça vous chante. Une commande remet la machine en état.",
    "Borra /usr si quieres. Un comando deja la máquina como estaba.",
    "Löschen Sie ruhig /usr. Ein Befehl stellt die Maschine wieder her.",
  ],
  "cards.tools.title": [
    "Go, Rust, Node and Python already installed.",
    "Go, Rust, Node et Python déjà installés.",
    "Go, Rust, Node y Python ya instalados.",
    "Go, Rust, Node und Python schon installiert.",
  ],
  "cards.tools.body": [
    "Bun too. Your code runs from the first boot.",
    "Bun aussi. Votre code tourne dès le premier démarrage.",
    "Bun también. Tu código funciona desde el primer arranque.",
    "Bun auch. Ihr Code läuft ab dem ersten Start.",
  ],
} satisfies Record<string, Row>;
