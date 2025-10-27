## Contexte

Tu es un développeur expert en go et tu développe un gestionnaire de téléchargement spécial pour des liens 1fichier.com via API avec compte premium. Tu as déjà un frontend svelteKit contenant 2 pages:
- La page /downloads affiche la liste de tous les téléchargements effectués.
- La page /home contient un formulaire pour envoyer un lien de téléchargement et afficher les téléchargements en cours.

## Objectif

Développer un backend en go avec Fiber, Gorm et Sqlite a partir des spécifications openapi ci-jointes.

## Consignes

Tu dois respecter les consignes suivantes:
- limite le nombre de dépendance externe
- utilise une structure de dossier standard
- applique le principe clean code
- génere des fichiers séparés
- sépare clairement le gestion du server sent event 
- sépare clairement le gestion de l'écriture du fichier sur le filesystem


## L'API 1fichier.com

Le téléchargement des fichiers ce fait en 3 étapes via l'api 1fichier.com et necessite l'api key d'un compte premium.


### étape 1

Récupération des infos du fichier a télécharger

#### Endpoint:

POST https://api.1fichier.com/v1/file/info.cgi + Header 'Authorization: Bearer {Settings.APIKey}

#### Réponse

```go
type OneFichierInfoResponse struct {
	URL string `json:"url"`
	Filename string `json:"filename"`
	Size int64 `json:"size"`
	Date time.Time `json:"date"`
	Checksum string `json:"checksum"`
	ContentType string `json:"content_type"`
	Description \*string `json:"description,omitempty"`
	Pass int `json:"pass"` // 0 ou 1
	Path string `json:"path"`
	FolderID string `json:"folder_id"`
}
```

### étape 2

Récupération du lien de téléchargement final du fichier

#### Endpoint:

POST https://api.1fichier.com/v1/download/get_token.cgi + Header 'Authorization: Bearer {Settings.APIKey}

#### Réponse

```go
type OneFichierTokenResponse struct {
	URL string `json:"url"`
	Status string `json:"status"` // "OK" ou "KO"
	Message \*string `json:"message,omitempty"`
}
```

### étape 3

Création le l'entrée Download et démarrage du téléchargement via ```OneFichierTokenResponse.URL```
