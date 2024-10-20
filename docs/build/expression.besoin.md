# Créateur d'ontologie

## Version

Version 0.1.0 Brainstorm

## Aperçu du projet

Développer un logiciel en Go qui, à partir de divers formats de documents (texte, PDF, Markdown, HTML, DOCX), crée une ontologie au format QuickStatement pour être compatible avec Wikibase. 

Le logiciel identifiera et extraira chaque élément d'information du document d'entrée, aussi petit ou apparemment insignifiant soit-il, tout en gérant efficacement les très grands documents.

## Fonctionnalités principales

1. **Support Multi-format d'Entrée**
   - Accepter les entrées en formats texte, PDF, Markdown, HTML et DOCX
   - Implémenter des analyseurs robustes pour chaque format supporté
   - Prise en charge de très grands documents (dépassant 120 000 tokens)

2. **Génération d'une ontologie au format QuickStatement / Wikibase**
   - Génération multi-passes : Le logiciel passe le nombre de fois indiqué en ligne de commande sur le document pour raffiner l'ontologie. La première passe est la constitution initiale de l'ontologie ; les passes suivantes visent à affiner les éléments. Le nombre est un paramètre de la ligne de commande.
   - Générer une sortie QuickStatement détaillée utilisant le vocabulaire Wikibase
   - Assurer une extraction complète des informations des documents d'entrée
   - Découper les documents larges en segments de maximum 4000 tokens pour ne pas impacter les output contexte des LLMs et gérer la sortie pour respecter la limite de 4 000 tokens par segment
   - S'assurer que chaque élément de l'ontologie est unitaire entre les segments et que le tout est cohérent. Chaque segment doit avoir la connaissance des éléments ontologiques du segment précédent pour que le LLM puisse s'appuyer dessus.
   - Lorsque l'on appelle le LLM pour le traitement d'un nouveau segment, on commence par lui donner l'état de l'ontologie.
   - Il doit être possible de traiter plusieurs documents en utilisant la même ontologie : donc on doit pouvoir donner une ontologie en ligne de commande pour que la première passe ne soit pas nécessaire.
   - Il faut un mode batch où l'on donne un répertoire et tous les fichiers présents dans le répertoire et les sous-répertoires sont traités. Si une ontologie est donnée en paramètre de ligne de commande, c'est elle qui sera utilisée et qui sera enrichie avec l'ensemble des documents.
   - Le résultat de l'exécution du logiciel est un fichier ayant l'extension .tsv (tab separated value) qui a le nom de l'ontologie. Le nom de l'ontologie doit être donné en ligne de commande par l'utilisateur. 
   - Ajouter des options d'export en formats RDF et OWL via des paramètres de ligne de commande
   - Utiliser un comptage précis des tokens compatible avec les modèles GPT pour assurer une segmentation exacte et cohérente des documents.

3. **Architecture Modulaire**
   - Utiliser Cobra pour la gestion des commandes CLI
   - Implémenter une architecture pipeline pour un traitement efficace des documents
   - Le verbe pour créer l'ontologie est "enrich" par exemple : ontology enrich <nom ontologie> <nom du fichier ou du répertoire> <flags>
   - On doit pouvoir supporter plusieurs LLMs, au minimum OpenAI GPT-4, Claude 3.5 Sonnet, Ollama
  
4. **Système de Journalisation**
   - Implémenter un système de journalisation polyvalent avec support pour les niveaux debug, info, warning et error
   - Exporter les logs vers des fichiers texte et les afficher sur la console pour le serveur et le CLI
   - Implémenter un mode silencieux (--silent) pour désactiver la sortie console des logs
   - Implémenter un mode debug (--debug) pour une journalisation très détaillée
   - Tous les output console doivent passer par le système de journalisation
   - Chaque étape doit être journalisée

5. **Gestion de la Configuration**
   - Utiliser YAML pour une configuration centralisée
   - Permettre des surcharges de paramètres par ligne de commande

6. **Versionnage**
   - Implémenter un suivi de version commençant à 0.1.0

## Exigences Détaillées

Pour tous les modules : 

- Limiter la taille d'un package à maximum 10 méthodes et découper le code logiciel finement

### 1. Analyse et Segmentation des Documents

- Développer des modules séparés pour l'analyse de chaque format supporté (texte, PDF, Markdown, HTML, DOCX)
- Implémenter un mécanisme de segmentation pour décomposer les grands documents en segments traitables
- Assurer une gestion robuste des erreurs pour les documents mal formés ou incomplets via le système de journalisation 
- Implémenter une interface commune pour tous les analyseurs afin de standardiser le processus d'extraction
- Préserver la structure du document et le contexte à travers les segments
- Implémenter un système de gestion des métadonnées du document (auteur, date de création, version, etc.). Les métadonnées doivent pouvoir être données en paramètre de ligne de commande.

### 2. Conversion QuickStatement TSV

- Créer un mappage complet des éléments du document vers QuickStatement/Wikibase
- Vérifier qu'il n'y a pas de concepts qui sont similaires ou il y ait peu de nuances ou alors que la nuance puisse être un paramètre d'un concept.
- Développer un moteur de conversion flexible capable de gérer diverses structures de documents
- S'assurer que chaque segment de sortie QuickStatement TSV respecte la limite de 4 000 tokens
- Implémenter un mécanisme pour lier les segments QuickStatement pour une représentation cohérente
- Intégrer des clients API pour différents LLM externes :
  - Claude (Anthropic)
  - GPT (OpenAI)
  - Ollama (pour les modèles locaux)
- Implémenter une interface commune `TranslationClient` pour tous les clients LLM
- Permettre la sélection du LLM à utiliser via la configuration ou les options de ligne de commande
- Ajouter une option `-i` pour enrichir les instructions de conversion envoyées au LLM
- Optimiser les requêtes aux LLM pour maximiser l'utilisation du contexte tout en respectant les limites de tokens
- Implémenter un système de gestion des erreurs et de reconnexion pour les appels API aux LLM : si un LLM ne répond pas dans le timeout, réessayer au bout de 20s en lui laissant plus de temps pour répondre par tranche de 20 secondes ; maximum 5 essais avant échec.

### 3. Journalisation et Surveillance

- Développer un module de journalisation centralisé supportant la sortie vers fichiers et console
- Implémenter la rotation et l'archivage des logs pour les logs basés sur fichiers
- Intégrer la journalisation dans toute l'application pour un suivi complet des opérations
- Implémenter la surveillance et le reporting des performances pour les tâches de traitement à grande échelle
- Ajouter des métriques de performance spécifiques à la gestion documentaire (temps de traitement par page, taux d'extraction, etc.)
- La dernière ligne du terminal doit afficher la progression de l'analyse d'un document et indiquer la progression du nombre total de documents. La console s'affiche au-dessus et n'empiète pas.
- Afficher le temps global écoulé depuis le début du traitement et le temps passé sur chaque chunk (segment) traité

### 4. Client CLI

- Développer une interface CLI conviviale en utilisant Cobra
- Implémenter des commandes pour :
  - La conversion de fichiers uniques
  - Le traitement par lots de plusieurs fichiers
  - La gestion de la configuration
  - Le contrôle du niveau de log
  - Le suivi de la progression pour le traitement de grands documents
  - La sélection du LLM à utiliser
- Ajouter une option `-i` pour spécifier des instructions supplémentaires pour la conversion
- Fournir une aide détaillée et des informations d'utilisation pour chaque commande
- Implémenter un mode interactif pour des interrogations à la volée d'un document sur la base d'une ontologie

### 5. Système de Configuration

- Développer un système de configuration basé sur YAML
- Implémenter le chargement de fichiers de configuration avec des surcharges spécifiques à l'environnement
- Permettre des surcharges de paramètres de configuration par ligne de commande
- Inclure des options de réglage des performances pour le traitement parallèle et la gestion de la mémoire

### 6. Documentation

- Créer une documentation utilisateur détaillée incluant des guides d'installation, de configuration et d'utilisation
- Développer une documentation technique pour l'utilisation et l'intégration de l'API
- Fournir des exemples et des meilleures pratiques pour une utilisation efficace de l'outil
- Inclure des directives pour l'optimisation des performances avec de grands documents

### 7. Internationalisation

- S'assurer que tout le texte visible par l'utilisateur est en anglais
- Concevoir le système pour supporter de futurs efforts de localisation
- Implémenter un support pour les jeux de caractères internationaux dans le traitement des documents
- Documenter le code produit

### 8. Sécurité et Confidentialité

- Ajouter une option pour le chiffrement des données sensibles dans les logs et les fichiers de sortie
- Implémenter un système basique de gestion des droits d'accès pour les différentes fonctionnalités

### 9. Interopérabilité

- Concevoir une API simple ou des hooks pour faciliter l'intégration future avec d'autres outils ou workflows

## Contraintes Techniques

- Développer en langage de programmation Go
- S'assurer qu'aucun fichier ne dépasse 3000 tokens
- Suivre les meilleures pratiques et les modèles idiomatiques de Go
- Utiliser les goroutines et les canaux pour le traitement concurrent lorsque c'est approprié
- Assurer la compatibilité avec différents LLM et leurs limites de contexte spécifiques (par exemple, limiter à environ 200 tokens par batch pour Ollama avec des modèles comme llama3.2)
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