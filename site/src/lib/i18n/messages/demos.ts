import type { Row } from "../row";

/** Words inside the tile demos and their accessible labels. */
export const demos = {
  "demo.agent.label": [
    "Bulle runs a test suite inside a Boîte machine and reports the result.",
    "Bulle lance des tests dans une machine Boîte et donne le résultat.",
    "Bulle ejecuta pruebas en una máquina Boîte y da el resultado.",
    "Bulle führt Tests in einer Boîte-Maschine aus und meldet das Ergebnis.",
  ],
  "demo.agent.prompt": [
    "Run the tests on fix/login.",
    "Lance les tests sur fix/login.",
    "Ejecuta las pruebas en fix/login.",
    "Führe die Tests auf fix/login aus.",
  ],
  "demo.agent.answer": [
    "All 48 tests pass. The code ran in the Boîte, never on your machine.",
    "Les 48 tests passent. Le code a tourné dans la Boîte, jamais sur votre machine.",
    "Las 48 pruebas pasan. El código se ejecutó en la Boîte, nunca en tu máquina.",
    "Alle 48 Tests bestehen. Der Code lief in der Boîte, nie auf Ihrem Rechner.",
  ],
  "demo.agent.composer": ["Message Bulle", "Écrire à Bulle", "Escribe a Bulle", "Nachricht an Bulle"],
  "demo.rollback.label": [
    "A snapshot is saved, /usr is deleted and the machine breaks, then a rollback restores it.",
    "Un instantané est enregistré, /usr est supprimé et la machine casse, puis un retour en arrière la répare.",
    "Se guarda una instantánea, se borra /usr y la máquina se rompe; luego una vuelta atrás la restaura.",
    "Ein Snapshot wird gespeichert, /usr gelöscht, die Maschine geht kaputt, ein Rollback stellt sie wieder her.",
  ],
  "demo.rollback.saved": ["Clean state saved", "État propre enregistré", "Estado limpio guardado", "Sauberer Stand gespeichert"],
  "demo.rollback.broken": ["The machine is broken", "La machine est cassée", "La máquina está rota", "Die Maschine ist kaputt"],
  "demo.rollback.back": ["Back to the clean state", "Retour à l'état propre", "De vuelta al estado limpio", "Zurück zum sauberen Stand"],
  "demo.tools.label": [
    "A new Boîte machine lists Go, Rust, Node, Python and Bun as ready.",
    "Une nouvelle machine Boîte affiche Go, Rust, Node, Python et Bun, prêts à l'emploi.",
    "Una nueva máquina Boîte muestra Go, Rust, Node, Python y Bun listos.",
    "Eine neue Boîte-Maschine zeigt Go, Rust, Node, Python und Bun als bereit.",
  ],
  "demo.tools.ready": ["Ready", "Prêt", "Listo", "Bereit"],
} satisfies Record<string, Row>;
