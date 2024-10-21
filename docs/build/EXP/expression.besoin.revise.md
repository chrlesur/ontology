# Créateur d'ontologie

## Version

Version 0.3.0 Révision basée sur l'implémentation et les retours d'exécution

## Aperçu du projet

Développer un logiciel en Go qui, à partir de divers formats de documents (texte, PDF, Markdown, HTML, DOCX), crée une ontologie au format QuickStatement pour être compatible avec Wikibase. Le logiciel doit identifier et extraire chaque élément d'information du document d'entrée, tout en gérant efficacement les très grands documents et en permettant un enrichissement itératif de l'ontologie.

## Fonctionnalités principales

### 1. Support Multi-format d'Entrée
- Accepter les entrées en formats texte, PDF, Markdown, HTML et DOCX
- Implémenter des analyseurs robustes pour chaque format supporté
- Pour les PDFs, utiliser la bibliothèque github.com/ledongthuc/pdf pour l'extraction de texte et de métadonnées
- Prise en charge de très grands documents (dépassant 120 000 tokens)
- Extraction cohérente des métadonnées à travers les différents formats

### 2. Génération et Enrichissement d'une ontologie au format QuickStatement / Wikibase
- Générer une sortie QuickStatement détaillée utilisant le vocabulaire Wikibase
- Implémenter un processus d'enrichissement d'ontologie multi-passes
- Utiliser des prompts spécifiques pour l'enrichissement initial et la fusion des résultats
- Développer une logique de fusion intelligente pour intégrer les nouveaux résultats de manière cohérente
- Gérer le contexte entre les segments et les passes d'enrichissement
- Implémenter un traitement complexe des chaînes de caractères pour gérer correctement les caractères d'échappement
- Assurer une extraction complète des informations des documents d'entrée
- Découper les documents larges en segments basés sur le nombre de tokens
- S'assurer que chaque élément de l'ontologie est unitaire entre les segments et que le tout est cohérent
- Traiter plusieurs documents en utilisant la même ontologie
- Le résultat de l'exécution du logiciel est un fichier ayant l'extension .tsv (tab separated value)
- Ajouter des options d'export en formats RDF et OWL

### 3. Segmentation et traitement du contenu
- Implémenter une segmentation sophistiquée créant des segments cohérents basés sur le nombre de tokens
- Utiliser tiktoken-go pour le comptage précis des tokens
- Optimiser la fonction de segmentation pour gérer efficacement de grands volumes de texte
- Implémenter une gestion efficace de la mémoire pour les très grands documents
- Assurer une intégration fluide entre le segmenter et le client LLM, en ajustant la taille des segments et la gestion du contexte

### 4. Intégration LLM
- Supporter plusieurs LLMs, au minimum OpenAI GPT-4, Claude 3.5 Sonnet, Ollama
- Implémenter une gestion robuste des limites de taux des API LLM, incluant un système de "token bucket" dans le client Claude (claude.go)
- Implémenter un backoff exponentiel avec un maximum de 5 tentatives pour les erreurs de limite de taux
- Gérer les différences entre les APIs d'OpenAI, Claude, et Ollama avec des adaptateurs spécifiques
- Optimiser les appels API aux LLMs pour réduire les coûts et les temps de traitement

### 5. Système de Prompts
- Implémenter un système de templates de prompts sophistiqué et flexible
- Créer une structure PromptTemplate avec des méthodes pour formater les prompts
- Créer des prompts spécifiques pour l'extraction d'entités, de relations, l'enrichissement d'ontologie et la fusion des résultats
- Assurer la compatibilité des prompts avec différents LLMs

### 6. Système de Journalisation
- Implémenter un système de journalisation polyvalent avec support pour les niveaux debug, info, warning et error
- Ajouter un mode de débogage détaillé activable via une option --debug
- Assurer que l'activation du mode debug n'affecte pas significativement les performances en mode normal
- Exporter les logs vers des fichiers texte et les afficher sur la console
- Implémenter un mode silencieux (--silent) pour désactiver la sortie console des logs
- Inclure des métriques de performance dans les logs, notamment pour le traitement parallèle et multi-passes

### 7. Gestion de la Configuration
- Utiliser YAML pour une configuration centralisée
- Permettre des surcharges de paramètres par ligne de commande
- Inclure des options de configuration pour les différents LLMs et leurs modèles spécifiques

### 8. Architecture et Modularité
- Utiliser Cobra pour la gestion des commandes CLI
- Implémenter une architecture pipeline pour un traitement efficace des documents
- Créer une couche d'abstraction pour les LLMs pour faciliter l'ajout futur de nouveaux modèles
- Séparer le système de prompts en son propre module pour améliorer la modularité et la réutilisabilité
- Implémenter un traitement parallèle des segments avec des goroutines pour améliorer les performances

### 9. Gestion des Erreurs et Robustesse
- Implémenter une gestion fine des erreurs pour les appels API aux LLMs, y compris la gestion des timeouts et des retries
- Assurer une validation rigoureuse des entrées et des sorties à chaque étape du pipeline
- Gérer de manière appropriée les erreurs spécifiques aux API LLM
- Implémenter une gestion robuste des erreurs lors de la fusion des résultats entre les passes

### 10. Tests et Validation
- Implémenter des tests unitaires pour chaque composant
- Créer des tests de bout en bout pour le pipeline complet
- Ajouter des tests de performance et de charge pour valider le comportement avec de grands volumes de données
- Inclure des tests spécifiques pour le mode de débogage et les nouvelles fonctionnalités de journalisation
- Implémenter des tests de charge spécifiques pour vérifier le comportement du système sous des conditions de limite de taux
- Ajouter des tests spécifiques pour la validation de l'ontologie enrichie après fusion
- Implémenter des tests de performance pour le traitement parallèle et multi-passes

### 11. Optimisation de la mémoire et des performances
- Implémenter des stratégies d'optimisation de la mémoire pour le traitement de très grands documents, en particulier dans le segmenter et le convertisseur
- Utiliser des techniques de streaming et de buffering pour minimiser l'utilisation de la mémoire lors du traitement de grands fichiers
- Optimiser le traitement par lots des grands documents pour éviter les problèmes de mémoire
- Implémenter des stratégies d'optimisation pour les appels API aux LLMs afin de réduire les coûts et les temps de traitement

### 12. Sécurité et Confidentialité
- Ajouter une option pour le chiffrement des données sensibles dans les logs et les fichiers de sortie
- Implémenter un système basique de gestion des droits d'accès pour les différentes fonctionnalités

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
4. Documentation utilisateur et technique avec des directives d'optimisation des performances
5. Fichiers de configuration d'exemple
6. README avec un guide de démarrage rapide et des instructions d'utilisation de base
7. Licence GPL3