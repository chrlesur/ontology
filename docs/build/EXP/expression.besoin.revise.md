# Créateur d'ontologie

## Version

Version 0.4.0 Révision basée sur l'implémentation et les retours d'exécution détaillés

## Aperçu du projet

Développer un logiciel en Go qui, à partir de divers formats de documents (texte, PDF, Markdown, HTML, DOCX), crée une ontologie au format QuickStatement pour être compatible avec Wikibase. Le logiciel doit identifier et extraire chaque élément d'information du document d'entrée, tout en gérant efficacement les très grands documents, les variations linguistiques et les termes composés.

## Fonctionnalités principales

### 1. Support Multi-format d'Entrée
- Accepter les entrées en formats texte, PDF, Markdown, HTML et DOCX
- Implémenter des analyseurs robustes pour chaque format supporté
- Prise en charge de très grands documents (dépassant 120 000 tokens)
- Extraction cohérente des métadonnées à travers les différents formats

### 2. Traitement linguistique et normalisation
- Implémenter une fonction de normalisation robuste pour les mots et termes
  - Conversion en minuscules
  - Remplacement des underscores par des espaces
  - Suppression de la ponctuation et des caractères non alphanumériques
  - Gestion appropriée des espaces
- Développer des méthodes pour gérer les variations linguistiques
- Évaluer et potentiellement intégrer des bibliothèques de traitement du langage naturel

### 3. Indexation et recherche de positions
- Créer un système d'indexation sophistiqué
  - Indexer les mots individuels et leurs combinaisons (jusqu'à 3 mots)
  - Gérer efficacement les termes composés
- Implémenter une fonction de recherche de positions avancée
  - Supporter la recherche exacte et partielle
  - Gérer efficacement les termes composés et leurs variations
- Optimiser les performances d'indexation et de recherche pour les grands documents

### 4. Génération et Enrichissement d'une ontologie
- Générer une sortie QuickStatement détaillée utilisant le vocabulaire Wikibase
- Implémenter un processus d'enrichissement d'ontologie multi-passes
- Développer une logique de fusion intelligente pour intégrer les nouveaux résultats de manière cohérente
- Gérer efficacement les positions multiples pour chaque concept
- Assurer une extraction complète des informations des documents d'entrée
- Traiter plusieurs documents en utilisant la même ontologie
- Ajouter des options d'export en formats RDF et OWL

### 5. Intégration LLM
- Supporter plusieurs LLMs, au minimum OpenAI GPT-4, Claude 3.5 Sonnet, Ollama
- Implémenter une gestion robuste des limites de taux des API LLM
- Optimiser les appels API aux LLMs pour réduire les coûts et les temps de traitement
- Développer des prompts spécifiques pour l'extraction d'entités, de relations et leurs descriptions

### 6. Système de Journalisation et Débogage
- Implémenter un système de journalisation polyvalent avec support pour les niveaux debug, info, warning et error
- Ajouter un mode de débogage détaillé activable via une option --debug
- Implémenter des logs détaillés à chaque étape du processus pour faciliter le débogage
- Assurer que l'activation du mode debug n'affecte pas significativement les performances en mode normal

### 7. Architecture et Modularité
- Utiliser Cobra pour la gestion des commandes CLI
- Implémenter une architecture pipeline pour un traitement efficace des documents
- Créer une couche d'abstraction pour les LLMs pour faciliter l'ajout futur de nouveaux modèles
- Concevoir une architecture flexible pour éviter les cycles d'importation entre packages

### 8. Gestion des Erreurs et Robustesse
- Implémenter une gestion fine des erreurs pour les appels API aux LLMs
- Assurer une validation rigoureuse des entrées et des sorties à chaque étape du pipeline
- Gérer de manière appropriée les erreurs spécifiques aux API LLM et aux opérations de fichiers

### 9. Optimisation des Performances
- Optimiser le traitement par lots des grands documents
- Implémenter des stratégies d'optimisation de la mémoire
- Utiliser des techniques de streaming et buffering pour les grands fichiers
- Optimiser l'indexation et la recherche pour les documents volumineux et les termes composés

### 10. Tests et Validation
- Implémenter des tests unitaires pour chaque composant
- Créer des tests de bout en bout pour le pipeline complet
- Ajouter des tests spécifiques pour la validation de l'ontologie enrichie
- Implémenter des tests de performance pour le traitement parallèle et multi-passes
- Créer des cas de test pour les scénarios linguistiques complexes (termes composés, variations)

## Exigences Détaillées

Pour tous les modules :
- Limiter la taille d'un package à maximum 10 méthodes et découper le code logiciel finement
- Optimiser le code pour la performance, particulièrement pour le traitement de grands documents
- Assurer une gestion robuste des erreurs à travers l'application
- Implémenter des logs détaillés pour chaque étape du processus

### 1. Analyse et Segmentation des Documents
- Développer des modules séparés pour l'analyse de chaque format supporté
- Implémenter un mécanisme de segmentation sophistiqué pour décomposer les grands documents
- Assurer une gestion robuste des erreurs pour les documents mal formés ou incomplets
- Préserver la structure du document et le contexte à travers les segments

### 2. Traitement Linguistique
- Implémenter une fonction de normalisation robuste pour les mots et termes
- Développer des méthodes pour gérer efficacement les variations linguistiques et les termes composés
- Optimiser les performances de traitement pour les grands volumes de texte

### 3. Indexation et Recherche
- Créer un système d'indexation efficace pour les mots individuels et les termes composés
- Implémenter une fonction de recherche flexible supportant la recherche exacte et partielle
- Optimiser les performances d'indexation et de recherche pour les grands documents

### 4. Enrichissement de l'Ontologie
- Développer un processus d'enrichissement itératif de l'ontologie
- Implémenter une gestion efficace des positions multiples pour chaque concept
- Assurer une fusion cohérente des résultats entre les passes d'enrichissement

### 5. Intégration LLM et Gestion des Prompts
- Créer une interface commune pour les clients LLM
- Développer des prompts spécifiques pour l'extraction d'entités, de relations et leurs descriptions
- Implémenter une gestion robuste des limites de taux d'API et des erreurs

### 6. Pipeline de Traitement Principal
- Intégrer tous les composants dans un pipeline de traitement cohérent
- Implémenter un traitement parallèle et multi-passes efficace
- Assurer une gestion robuste des erreurs et des cas limites

## Contraintes Techniques
- Développer en langage de programmation Go
- Suivre les meilleures pratiques et les modèles idiomatiques de Go
- Utiliser les goroutines et les canaux pour le traitement concurrent lorsque c'est approprié
- Assurer la compatibilité avec différents LLM et leurs limites de contexte spécifiques
- Une méthode ne peut pas faire plus de 80 lignes
- Un fichier de code source Go d'un package pas plus de 10 méthodes

## Livrables
1. Dépôt de code source avec des packages Go bien structurés
2. Binaires exécutables pour les principaux systèmes d'exploitation (Windows, macOS, Linux)
3. Suite de tests complète incluant des tests de performance et de stress
4. Documentation utilisateur et technique détaillée
5. Fichiers de configuration d'exemple
6. README avec un guide de démarrage rapide et des instructions d'utilisation de base
7. Licence GPL3