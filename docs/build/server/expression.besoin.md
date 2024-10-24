Voici l'expression de besoin en format Markdown :

# Expression de besoin - Serveur Ontology avec API et interface web

## 1. Objectif général

Développer un logiciel serveur avec API s'appuyant sur le code existant du client Ontology, capable de gérer plusieurs ontologies et de fournir un moteur de recherche avancé.

## 2. Fonctionnalités principales

### 2.1 Importation et gestion des données

- Importer et gérer plusieurs fichiers d'ontologie aux formats TSV, RDF ou OWL
- Importer automatiquement les fichiers de contexte JSON associés (NOMDUDOCUMENT_context.json) pour chaque ontologie
- Extraire et stocker les métadonnées de chaque fichier :
  - Localisation
  - Taille
  - SHA256
- Permettre l'ajout, la mise à jour et la suppression d'ontologies

### 2.2 Gestion des ontologies

- Importer chaque ontologie avec tous ses éléments et leurs positions
- Stocker les informations de positions en base de données pour chaque ontologie
- Stocker les ontologies complètes en base de données
- Gérer les relations entre les différentes ontologies si nécessaire

### 2.3 Moteur de recherche

- Implémenter une API de recherche capable de chercher dans toutes les ontologies stockées
- Utiliser un LLM pour composer les prompts de recherche
- Permettre la recherche dans une ontologie spécifique ou dans toutes les ontologies
- Retourner les résultats de recherche incluant :
  - Nom du document/ontologie
  - Position dans le document
  - Extrait de texte (30 mots avant et après la position par défaut)

### 2.4 Interface utilisateur

- Développer une interface web en HTML5/React
- Fournir une interface pour :
  - L'importation et la gestion de multiples ontologies
  - La recherche dans une ou plusieurs ontologies
  - L'affichage des résultats de recherche
  - La visualisation et la navigation dans les ontologies

## 3. Spécifications techniques

### 3.1 Backend

- Développer un serveur en Go (dernière version stable)
- Utiliser et adapter le code existant du client Ontology
- Implémenter une API RESTful avec des endpoints pour gérer plusieurs ontologies
- Utiliser les goroutines et les canaux pour le traitement concurrent lorsque c'est approprié
- Assurer la compatibilité avec différents LLM et leurs limites de contexte spécifiques

### 3.2 Base de données

- Utiliser une base de données appropriée (par exemple PostgreSQL) pour stocker efficacement plusieurs ontologies et leurs informations contextuelles
- Concevoir un schéma de base de données permettant de gérer plusieurs ontologies et leurs relations

### 3.3 Frontend

- Développer une interface web en React
- Implémenter des composants pour la gestion et la visualisation de multiples ontologies

### 3.4 Intégration LLM

- Intégrer le LLM existant pour la composition des prompts de recherche
- Adapter le système pour prendre en compte la recherche dans plusieurs ontologies

## 4. Contraintes techniques et directives de développement

### 4.1 Structure du code

- Limiter chaque fichier de code source Go à un maximum de 3000 tokens
- Limiter chaque package à un maximum de 10 méthodes exportées
- Limiter chaque méthode à un maximum de 80 lignes de code

### 4.2 Meilleures pratiques

- Suivre les meilleures pratiques et les modèles idiomatiques de Go
- Utiliser le package 'internal/logger' pour toute journalisation, avec les niveaux : debug, info, warning, et error
- Définir toutes les valeurs configurables dans le package 'internal/config'
- Gérer toutes les erreurs de manière appropriée, en utilisant error wrapping lorsque c'est pertinent
- Utiliser les constantes définies dans le package 'internal/i18n' pour les messages utilisateur

### 4.3 Documentation et tests

- Fournir des commentaires de documentation conformes aux standards GoDoc pour chaque fonction, méthode et type exporté
- Développer des tests unitaires pour chaque nouvelle fonction ou méthode
- Fournir une documentation complète de l'API et du déploiement

### 4.4 Internationalisation et localisation

- Tous les messages visibles par l'utilisateur doivent être en anglais
- Préparer le code pour de futurs efforts de localisation

### 4.5 Performance et sécurité

- Optimiser le code pour la performance, particulièrement pour le traitement de grands documents
- Assurer la sécurité du code, en particulier lors du traitement des entrées utilisateur

## 5. Sécurité et performance

- Assurer la sécurité des données et de l'API
- Optimiser les performances pour la gestion et la recherche dans de grandes quantités d'ontologies
- Implémenter un système de gestion des utilisateurs et des droits d'accès aux différentes ontologies

## 6. Évolutivité et maintenance

- Concevoir le système de manière à faciliter l'ajout de nouvelles fonctionnalités
- Prévoir des mécanismes de sauvegarde et de restauration des données
- Implémenter un système de logging pour le suivi des opérations et le débogage