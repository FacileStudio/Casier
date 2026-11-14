import type { Row } from '../row';

/** The public landing page and the sign-in tagline, worded from the suite's canonical copy. */
export const landing = {
	'login.tagline': [
		'Your team’s secrets, encrypted and injected into your apps.',
		'Les secrets de l’équipe, chiffrés et injectés dans vos apps.',
		'Los secretos del equipo, cifrados e inyectados en tus apps.',
		'Team-Secrets, verschlüsselt und direkt in Ihre Apps.'
	],
	'landing.title': [
		'Casier · No more .env files lying around',
		'Casier · Fini les .env qui traînent',
		'Casier · Se acabaron los .env por ahí sueltos',
		'Casier · Schluss mit herumliegenden .env-Dateien'
	],
	'landing.metaDescription': [
		'Your team’s secrets, encrypted in one place and injected into your apps at launch. Every read is logged. Free to start.',
		'Les secrets de votre équipe, chiffrés au même endroit et injectés dans vos apps au lancement. Chaque lecture est tracée. Gratuit pour commencer.',
		'Los secretos de tu equipo, cifrados en un solo lugar e inyectados en tus apps al arrancar. Cada lectura queda registrada. Gratis para empezar.',
		'Die Secrets Ihres Teams, verschlüsselt an einem Ort und beim Start in Ihre Apps eingespielt. Jeder Zugriff wird protokolliert. Kostenlos starten.'
	],
	'landing.heroTitle': [
		'No more .env files lying around.',
		'Fini les .env qui traînent.',
		'Se acabaron los .env por ahí sueltos.',
		'Schluss mit herumliegenden .env-Dateien.'
	],
	'landing.description': [
		'Your team’s secrets, encrypted in one place and injected into your apps at launch. Every read is logged.',
		'Les secrets de votre équipe, chiffrés au même endroit et injectés dans vos apps au lancement. Chaque lecture est tracée.',
		'Los secretos de tu equipo, cifrados en un solo lugar e inyectados en tus apps al arrancar. Cada lectura queda registrada.',
		'Die Secrets Ihres Teams, verschlüsselt an einem Ort und beim Start in Ihre Apps eingespielt. Jeder Zugriff wird protokolliert.'
	],
	'landing.demo.label': [
		'A secret is added to the prod environment, then casier run starts an app with it. Every value stays masked.',
		'Un secret est ajouté à l’environnement prod, puis casier run lance une application avec. Chaque valeur reste masquée.',
		'Se añade un secreto al entorno prod y casier run arranca una aplicación con él. Todos los valores siguen ocultos.',
		'Ein Secret kommt in die Umgebung prod, dann startet casier run eine App damit. Jeder Wert bleibt maskiert.'
	],
	'landing.nav.features': ['Features', 'Fonctionnalités', 'Funciones', 'Funktionen'],
	'landing.nav.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.nav.developers': ['Developers', 'Développeurs', 'Desarrolladores', 'Entwickler'],
	'landing.tiles.title': ['Made for a team', 'Pensé pour une équipe', 'Pensado para un equipo', 'Für Teams gemacht'],
	'landing.tile.audit.title': [
		'Every read is logged',
		'Chaque lecture est tracée',
		'Cada lectura queda registrada',
		'Jeder Zugriff wird protokolliert'
	],
	'landing.tile.audit.body': [
		'The audit log shows who read which secret, and when.',
		'Le journal d’audit montre qui a lu quel secret, et quand.',
		'El registro de auditoría muestra quién leyó qué secreto y cuándo.',
		'Das Audit-Log zeigt, wer welches Secret wann gelesen hat.'
	],
	'landing.tile.audit.label': [
		'The audit log fills up: a person reveals a value, a CI token exports prod, an admin rolls a secret back.',
		'Le journal d’audit se remplit : une personne révèle une valeur, un jeton de CI exporte prod, un admin revient à une version.',
		'El registro de auditoría se llena: una persona revela un valor, un token de CI exporta prod, un admin revierte un secreto.',
		'Das Audit-Log füllt sich: Eine Person zeigt einen Wert an, ein CI-Token exportiert prod, ein Admin setzt ein Secret zurück.'
	],
	'landing.tile.env.title': [
		'One environment per deploy',
		'Un environnement par déploiement',
		'Un entorno por despliegue',
		'Eine Umgebung pro Deployment'
	],
	'landing.tile.env.body': [
		'dev, staging and prod each keep their own values and history.',
		'dev, staging et prod gardent chacun leurs valeurs et leur historique.',
		'dev, staging y prod guardan cada uno sus valores y su historial.',
		'dev, staging und prod behalten jeweils eigene Werte und einen eigenen Verlauf.'
	],
	'landing.tile.env.label': [
		'Switching from dev to staging to prod shows each environment’s own secrets and versions.',
		'Passer de dev à staging puis à prod montre les secrets et les versions propres à chaque environnement.',
		'Pasar de dev a staging y a prod muestra los secretos y versiones de cada entorno.',
		'Der Wechsel von dev zu staging zu prod zeigt die eigenen Secrets und Versionen jeder Umgebung.'
	],
	'landing.tile.bulle.title': [
		'Bulle asks, you approve',
		'Bulle demande, vous approuvez',
		'Bulle pide, tú apruebas',
		'Bulle fragt, Sie geben frei'
	],
	'landing.tile.bulle.body': [
		'Bulle reads a secret only once you allow it, one time, within 15 minutes.',
		'Bulle ne lit un secret qu’après votre accord, une seule fois, dans les 15 minutes.',
		'Bulle solo lee un secreto cuando lo permites, una vez, en 15 minutos.',
		'Bulle liest ein Secret erst nach Ihrer Freigabe, einmal, innerhalb von 15 Minuten.'
	],
	'landing.tile.bulle.label': [
		'Bulle asks to read STRIPE_SECRET_KEY in prod. You click Allow reading; the value is never shown on the card.',
		'Bulle demande à lire STRIPE_SECRET_KEY en prod. Vous cliquez sur Autoriser la lecture ; la valeur n’apparaît jamais sur la carte.',
		'Bulle pide leer STRIPE_SECRET_KEY en prod. Haces clic en Permitir la lectura; el valor nunca aparece en la tarjeta.',
		'Bulle möchte STRIPE_SECRET_KEY in prod lesen. Sie klicken auf Lesen erlauben; der Wert erscheint nie auf der Karte.'
	],
	'landing.tile.bulle.allowed': [
		'Allowed: Bulle can read it once.',
		'Autorisé : Bulle peut le lire une fois.',
		'Permitido: Bulle puede leerlo una vez.',
		'Freigegeben: Bulle darf es einmal lesen.'
	],
	'landing.tile.bulle.badge': ['Pro plan', 'Offre Pro', 'Plan Pro', 'Pro-Tarif'],
	'landing.features.title': ['Safe by default', 'Sûr par défaut', 'Seguro por defecto', 'Standardmäßig sicher'],
	'landing.features.line': [
		'What keeps a secret where it belongs.',
		'Ce qui garde chaque secret à sa place.',
		'Lo que mantiene cada secreto en su sitio.',
		'Was jedes Secret an seinem Platz hält.'
	],
	'landing.feature.encrypted.title': ['Encrypted at rest', 'Chiffré au repos', 'Cifrado en reposo', 'Verschlüsselt gespeichert'],
	'landing.feature.encrypted.body': [
		'Each project has its own AES-256-GCM key.',
		'Chaque projet a sa propre clé AES-256-GCM.',
		'Cada proyecto tiene su propia clave AES-256-GCM.',
		'Jedes Projekt hat einen eigenen AES-256-GCM-Schlüssel.'
	],
	'landing.feature.ci.title': ['Made for CI', 'Pensé pour la CI', 'Pensado para CI', 'Für CI gemacht'],
	'landing.feature.ci.body': [
		'Scoped tokens give a machine only the environments it needs.',
		'Des jetons limités donnent à une machine les seuls environnements utiles.',
		'Los tokens acotados dan a una máquina solo los entornos que necesita.',
		'Eingeschränkte Tokens geben einer Maschine nur die Umgebungen, die sie braucht.'
	],
	'landing.feature.roles.title': ['A role per person', 'Un rôle par personne', 'Un rol por persona', 'Eine Rolle pro Person'],
	'landing.feature.roles.body': [
		'Owner, admin, member or viewer, set per project.',
		'Propriétaire, admin, membre ou lecteur, projet par projet.',
		'Propietario, admin, miembro o lector, proyecto a proyecto.',
		'Eigentümer, Admin, Mitglied oder Leser, je Projekt.'
	],
	'landing.feature.history.title': ['Undo a bad change', 'Annulez une erreur', 'Deshaz un error', 'Fehler rückgängig machen'],
	'landing.feature.history.body': [
		'Roll a secret back to an earlier version, restore a deleted one for 30 days.',
		'Revenez à une version précédente, restaurez un secret supprimé pendant 30 jours.',
		'Vuelve a una versión anterior y restaura un secreto borrado durante 30 días.',
		'Zu einer früheren Version zurück, gelöschte Secrets 30 Tage lang wiederherstellen.'
	],
	'landing.with.bulle': [
		'Bulle asks for a secret, you approve.',
		'Bulle demande un secret, vous approuvez.',
		'Bulle pide un secreto y tú lo apruebas.',
		'Bulle fragt nach einem Secret, Sie geben es frei.'
	],
	'landing.with.boite': [
		'Your secrets reach the machine, never a .env file.',
		'Vos secrets arrivent dans la machine, jamais dans un fichier .env.',
		'Tus secretos llegan a la máquina, nunca a un .env.',
		'Ihre Secrets landen in der VM, nie in einer .env-Datei.'
	],
	'landing.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.plan.included': [
		'Included in every plan, the free one too.',
		'Inclus dans chaque offre, gratuite comprise.',
		'Incluido en todos los planes, también el gratuito.',
		'In jedem Tarif enthalten, auch im kostenlosen.'
	],
	'landing.dev.line': [
		'A CLI (casier run), CI tokens, an API and an MCP server. Open source, and you can run it yourself: see self-hosting.',
		'Un CLI (casier run), des jetons pour la CI, une API et un serveur MCP. Open source, installable chez vous : voir l’auto-hébergement.',
		'Un CLI (casier run), tokens para la CI, una API y un servidor MCP. Código abierto, y puedes instalarlo tú mismo: ver autoalojamiento.',
		'Eine CLI (casier run), CI-Tokens, eine API und ein MCP-Server. Open Source, auch selbst betreibbar: siehe Self-Hosting.'
	],
	'landing.dev.cli': [
		'casier run puts secrets in your app’s environment, not in a .env file.',
		'casier run place les secrets dans l’environnement de votre app, pas dans un fichier .env.',
		'casier run pone los secretos en el entorno de tu app, no en un archivo .env.',
		'casier run legt Secrets in die Umgebung Ihrer App, nicht in eine .env-Datei.'
	],
	'landing.dev.api': [
		'A REST API for projects, environments, secrets and tokens.',
		'Une API REST pour les projets, environnements, secrets et jetons.',
		'Una API REST para proyectos, entornos, secretos y tokens.',
		'Eine REST-API für Projekte, Umgebungen, Secrets und Tokens.'
	],
	'landing.dev.mcp': [
		'Your AI client lists secrets, and reads a value only after you confirm it in Casier.',
		'Votre client IA liste les secrets et ne lit une valeur qu’après votre confirmation dans Casier.',
		'Tu cliente de IA lista los secretos y solo lee un valor tras tu confirmación en Casier.',
		'Ihr KI-Client listet Secrets und liest einen Wert erst nach Ihrer Bestätigung in Casier.'
	],
	'landing.dev.source': [
		'Open code: your data stays yours.',
		'Code ouvert, vos données restent les vôtres.',
		'Código abierto: tus datos siguen siendo tuyos.',
		'Offener Code: Ihre Daten bleiben Ihre.'
	],
	'landing.dev.selfHostHref': [
		'https://suite.facile.studio/en/auto-hebergement',
		'https://suite.facile.studio/auto-hebergement',
		'https://suite.facile.studio/es/auto-hebergement',
		'https://suite.facile.studio/de/auto-hebergement'
	],
	'landing.ctaTitle': [
		'Put your secrets away today',
		'Rangez vos secrets dès aujourd’hui',
		'Guarda tus secretos hoy',
		'Räumen Sie Ihre Secrets heute auf'
	],
	'landing.ctaText': [
		'Free to start, with your Facile account.',
		'Gratuit pour commencer, avec votre compte Facile.',
		'Gratis para empezar, con tu cuenta de Facile.',
		'Kostenlos starten, mit Ihrem Facile-Konto.'
	]
} as const satisfies Record<string, Row>;
