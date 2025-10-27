## Contexte

Tu es un développeur expert en backend node js et tu développe un gestionnaire de téléchargement spécial pour des liens 1fichier.com via API avec compte premium. Tu as déjà un frontend svelteKit contenant 2 pages:

- La page /downloads affiche la liste de tous les téléchargements effectués.
- La page /home contient un formulaire pour envoyer un lien de téléchargement et afficher les téléchargements en cours.

## Objectif

Développer un backend en typescript avec fastify, prisma et Sqlite a partir des spécifications openapi ci-jointes.

## Consignes

Tu dois respecter les consignes suivantes:

- limite le nombre de dépendance externe
- utilise une structure de dossier standard
- applique les principes KISS et clean code tant que possible

## L'API 1fichier.com

Le téléchargement des fichiers ce fait en 3 étapes via l'api 1fichier.com et necessite l'api key d'un compte premium.

### étape 1

Récupération des infos du fichier a télécharger

#### Endpoint:

POST https://api.1fichier.com/v1/file/info.cgi + Header 'Authorization: Bearer {Settings.APIKey}

#### Réponse

```ts
interface OneFichierInfoResponse {
  url: string;
  filename: string;
  size: number;
  date: string;
  checksum: string;
  content_type: string;
  description?: string;
  pass: number; // 0 or 1
  path: string;
  folder_id: string;
}
```

### étape 2

Récupération du lien de téléchargement final du fichier

#### Endpoint:

POST https://api.1fichier.com/v1/download/get_token.cgi + Header 'Authorization: Bearer {Settings.APIKey}

#### Réponse

```ts
interface OneFichierTokenResponse {
  url: string;
  status: string; // "OK" ou "KO"
  message?: string;
}
```

### étape 3

Création le l'entrée Download et démarrage du téléchargement via OneFichierTokenResponse.URL

```

```
