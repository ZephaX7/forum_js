# 🚗 BMW Community Forum

Un centre de discussion moderne pour la communauté passionnée de BMW!

## 🚀 Fonctionnalités

- ✅ **Authentification utilisateur** - Inscription et connexion sécurisée
- ✅ **Discussions organisées** - Créez et répondez à des discussions par catégorie
- ✅ **Système de réponses** - Engagez-vous avec d'autres membres
- ✅ **Catégories** - Mécanique, Tuning, Galerie, Événements
- ✅ **Interface moderne** - Design responsive et intuitif
- ✅ **Base de données SQLite** - Persistance des données

## 🛠️ Installation

### Prérequis
- Go 1.20 ou supérieur
- Git

### Étapes

1. **Clonez le projet**
```bash
cd forum_js
```

2. **Téléchargez les dépendances**
```bash
go mod download
go mod tidy
```

3. **Compilez l'application**
```bash
go build -o forum main.go routes.go
```

4. **Lancez le serveur**
```bash
./forum
```

L'application sera disponible sur `http://localhost:8080`

## 📁 Structure du projet

```
forum_js/
├── main.go              # Point d'entrée et configuration BD
├── routes.go            # Définition des routes et API
├── go.mod               # Dépendances Go
├── templates/           # Pages HTML
│   ├── index.html       # Page d'accueil
│   ├── login.html       # Page de connexion
│   ├── register.html    # Page d'inscription
│   └── forum.html       # Page du forum
└── static/              # Ressources statiques
    ├── style.css        # Styles globaux
    └── forum.js         # JavaScript du forum
```

## 💡 Utilisation

1. **Visitez la page d'accueil** `http://localhost:8080`
2. **Créez un compte** ou connectez-vous
3. **Découvrez les discussions** ou créez la vôtre
4. **Participez** en répondant aux sujets

## 🎨 Personnalisation

- Modifiez `static/style.css` pour changer l'apparence
- Ajoutez de nouvelles catégories dans `templates/forum.html`
- Personnalisez les textes dans les fichiers HTML

## 🔒 Sécurité

⚠️ **Note importante**: Cette application est un prototype. Pour la production:
- Hashez les mots de passe avec bcrypt
- Implémentez les sessions avec JWT
- Validez et sanitizez les entrées utilisateur
- Utilisez HTTPS
- Limitez les requêtes (rate limiting)


---

**Bienvenue dans le centre de discussion BMW! 🏁**
