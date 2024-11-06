# Plan d'action détaillé pour le développement d'Ontology

## Directives générales de développement

Ces directives s'appliquent à toutes les étapes du développement et doivent être suivies par tous les exécutants :

1. Langage et version : Développer exclusivement en Go, en utilisant la dernière version stable.

2. Journalisation : 
   - Toutes les actions et événements significatifs doivent être logués en utilisant le package `internal/logger`.
   - Implémenter les niveaux de log : debug, info, warning, et error.
   - Supporter l'export des logs vers des fichiers texte et l'affichage sur la console.
   - Implémenter un mode silencieux (--silent) et un mode debug (--debug).

3. Taille et structure du code :
   - Aucun fichier de code source ne doit dépasser 3000 tokens.
   - Chaque package ne doit pas contenir plus de 10 méthodes exportées.
   - La taille totale d'un package ne doit pas dépasser 4000 tokens.
   - Aucune méthode ne doit dépasser 80 lignes de code.

4. Architecture :
   - Utiliser Cobra pour la gestion des commandes CLI.
   - Implémenter une architecture pipeline pour un traitement efficace des documents.

5. Gestion des documents :
   - Supporter les formats : texte, PDF, Markdown, HTML, DOCX.
   - Implémenter la segmentation des documents en chunks de maximum 4000 tokens.
   - Gérer efficacement les très grands documents (dépassant 120 000 tokens).
   - Préserver la structure et le contexte du document à travers les segments.

6. Intégration LLM :
   - Supporter au minimum OpenAI GPT-4, Claude 3.5 Sonnet, et Ollama.
   - Lors de l'interaction avec les LLM, respecter les limites de contexte spécifiques à chaque modèle.
   - Pour Ollama avec des modèles comme llama3.2, limiter à environ 200 tokens par batch.
   - Implémenter un système de gestion des erreurs et de reconnexion pour les appels API aux LLM.

7. Gestion de l'ontologie :
   - Générer une sortie au format QuickStatement TSV pour Wikibase.
   - Implémenter des options d'export en formats RDF et OWL.
   - Assurer la cohérence de l'ontologie entre les segments et les passes multiples.

8. Configuration :
   - Utiliser YAML pour la configuration centralisée.
   - Permettre des surcharges de paramètres par ligne de commande.
   - Toutes les valeurs configurables doivent être définies dans le package `internal/config`.

9. Documentation :
   - Chaque fonction, méthode et type exporté doit avoir un commentaire de documentation conforme aux standards GoDoc.
   - Créer une documentation utilisateur détaillée incluant des guides d'installation, de configuration et d'utilisation.

10. Tests et qualité :
    - Chaque package doit avoir une couverture de tests d'au moins 80%.
    - Inclure des tests unitaires, des tests d'intégration, et des tests de performance.

11. Gestion des erreurs :
    - Toutes les erreurs doivent être gérées et propagées de manière appropriée.
    - Utiliser des error wrapping lorsque c'est pertinent.

12. Concurrence :
    - Utiliser les goroutines et les canaux de manière appropriée pour le traitement parallèle.
    - Veiller à éviter les race conditions.

13. Internationalisation :
    - Tous les messages visibles par l'utilisateur doivent être en anglais.
    - Définir les messages dans le package `internal/i18n`.
    - Concevoir le système pour supporter de futurs efforts de localisation.

14. Performance :
    - Optimiser le code pour la performance, en particulier pour le traitement de grands documents.
    - Utiliser le profilage Go pour identifier les goulots d'étranglement.

15. Sécurité :
    - Implémenter des options pour le chiffrement des données sensibles dans les logs et les fichiers de sortie.
    - Mettre en place un système basique de gestion des droits d'accès.

16. Versionnage :
    - Implémenter un suivi de version commençant à 0.1.0.
    - Utiliser les tags Git pour le versioning.

17. Compilation et distribution :
    - Fournir des binaires exécutables pour Windows, macOS, et Linux.
    - Utiliser un Makefile pour automatiser les tâches de build, test, et release.

## Étapes de développement

### 1. Configuration initiale du projet

Instruction : "Créez la structure exacte suivante pour le projet Go :

/Ontology /cmd /ontology main.go /internal /parser /segmenter /converter /llm /logger /config /pipeline /i18n /pkg go.mod Makefile README.md

Initialisez le module Go avec 'go mod init github.com/chrlesur/Ontology'. Dans le fichier main.go, importez uniquement le package 'cmd/ontology' et appelez la fonction Run()."

### 2. Mise en place de Cobra pour le CLI

Instruction : "Dans cmd/ontology/root.go, créez la commande racine 'ontology' avec Cobra. Définissez les flags globaux suivants : --config (string), --debug (bool), --silent (bool). Dans cmd/ontology/enrich.go, créez la sous-commande 'enrich' avec les flags : --input (string), --output (string), --format (string), --llm (string), --llm-model (string), --passes (int), --rdf (bool), --owl (bool), --recursive (bool). La fonction Run() de 'enrich' doit appeler une fonction du package 'pipeline' nommée ExecutePipeline()."

### 3. Développement des analyseurs de documents

Instruction : "Dans internal/parser, créez une interface Parser avec une seule méthode Parse(path string) ([]byte, error). Implémentez cette interface pour chaque format (text.go, pdf.go, markdown.go, html.go, docx.go). Chaque implémentation doit utiliser uniquement la bibliothèque standard Go ou une bibliothèque tierce spécifiée (par exemple, 'github.com/unidoc/unipdf/v3' pour PDF). Ajoutez une fonction GetMetadata() map[string]string à chaque parser pour extraire les métadonnées du document. Créez une fonction ParseDirectory(path string, recursive bool) ([][]byte, error) qui parcourt un répertoire (et ses sous-répertoires si recursive est true) et parse tous les fichiers supportés."

### 4. Implémentation de la segmentation des documents

Instruction : "Dans internal/segmenter/segmenter.go, créez une fonction Segment(content []byte, maxTokens int) ([][]byte, error) qui divise le contenu en segments de maxTokens. Utilisez un algorithme de tokenization précis compatible avec les modèles GPT (par exemple, en utilisant la bibliothèque 'github.com/pkoukk/tiktoken-go') pour le comptage des tokens. Assurez-vous que chaque segment se termine par une phrase complète. Implémentez une fonction GetContext(segments [][]byte, currentIndex int) string qui retourne le contexte des segments précédents. Créez une fonction CountTokens(content []byte) int qui retourne le nombre exact de tokens dans un contenu donné. Ajoutez une fonction pour calibrer le comptage des tokens en fonction du modèle LLM spécifique utilisé."

### 5. Création du moteur de conversion QuickStatement

Instruction : "Dans internal/converter/quickstatement.go, implémentez une fonction Convert(segment []byte, context string, ontology string) (string, error) qui transforme un segment en format QuickStatement TSV. Utilisez une table de mapping prédéfinie pour les éléments Wikibase. La sortie doit être une chaîne formatée en TSV. Implémentez également des fonctions ConvertToRDF(quickstatement string) (string, error) et ConvertToOWL(quickstatement string) (string, error) pour les exports additionnels."

### 6. Intégration des clients LLM

Instruction : "Dans internal/llm, créez une interface Client avec une méthode Translate(prompt string, context string) (string, error). Implémentez cette interface pour chaque LLM supporté (openai.go, claude.go, ollama.go). Pour chaque implémentation, ajoutez un paramètre 'model' à la méthode de création du client (par exemple, NewOpenAIClient(apiKey string, model string) *OpenAIClient). Supportez au minimum les modèles suivants :
- OpenAI : 'gpt-3.5-turbo', 'gpt-4'
- Claude : 'claude-3-opus-20240229', 'claude-3-sonnet-20240229'
- Ollama : 'llama2', 'llama2:13b', 'llama2:70b'
Utilisez les SDK officiels ou des requêtes HTTP simples pour l'interaction avec les API. Implémentez un système de retry avec backoff exponentiel en cas d'erreur, avec un maximum de 5 tentatives."

### 7. Développement du système de journalisation

Instruction : "Dans internal/logger/logger.go, implémentez des fonctions Debug(), Info(), Warning(), et Error() qui écrivent dans un fichier de log et sur la console. Utilisez le package 'log' de la bibliothèque standard. Implémentez la rotation des logs avec un fichier de 10MB maximum. Ajoutez une fonction UpdateProgress(current, total int) qui met à jour la dernière ligne de la console avec la progression actuelle."

### 8. Implémentation du système de configuration

Instruction : "Dans internal/config/config.go, créez une structure Config qui correspond exactement aux flags CLI, y compris les nouveaux flags --llm-model et --recursive. Implémentez une fonction LoadConfig(path string) (Config, error) qui charge un fichier YAML. Utilisez le package 'gopkg.in/yaml.v2' pour le parsing YAML. Ajoutez une fonction ValidateConfig(config Config) error pour vérifier la validité de la configuration, y compris la validation du modèle LLM spécifié."

### 9. Développement du pipeline de traitement principal

Instruction : "Dans internal/pipeline/pipeline.go, créez une fonction ExecutePipeline(config Config) error qui orchestre le flux de travail complet. Cette fonction doit gérer le traitement d'un seul fichier ou d'un répertoire entier selon la valeur de config.Input. Si c'est un répertoire, utilisez la fonction ParseDirectory du package parser. Pour chaque document, appelez séquentiellement les fonctions des autres packages dans cet ordre : parser, segmenter, converter, llm. Utilisez un WaitGroup pour le traitement parallèle des segments. Implémentez une boucle pour les passes multiples d'enrichissement de l'ontologie. Assurez-vous que le LLM client est créé avec le modèle spécifié dans config.LLMModel."

### 10. Implémentation des fonctionnalités de sécurité

Instruction : "Dans internal/logger/logger.go, ajoutez une fonction SetEncryption(key string) qui active le chiffrement AES pour les logs. Dans internal/config/config.go, ajoutez un champ 'AccessLevel' à la structure Config et implémentez une fonction CheckAccess(level string) bool."

### 11. Développement des tests unitaires et d'intégration

Instruction : "Pour chaque package, créez un fichier *_test.go avec des tests unitaires pour chaque fonction publique. Utilisez le package 'testing' de la bibliothèque standard. Dans le dossier racine, créez un fichier integration_test.go avec un test qui exécute le pipeline complet sur un petit document d'exemple. Ajoutez des benchmarks pour mesurer les performances de chaque composant."

### 12. Création de la documentation

Instruction : "Ajoutez des commentaires de documentation Go à toutes les fonctions et structures publiques. Dans le README.md, incluez les sections suivantes : Installation, Configuration, Utilisation, Exemples, Contribution, Licence. Utilisez des blocs de code Markdown pour les exemples de commandes CLI. Générez la documentation GoDoc pour tous les packages."

### 13. Mise en place du système de build et de release

Instruction : "Dans le Makefile, définissez les cibles : build, test, lint, clean. La cible build doit compiler des binaires pour Windows, macOS et Linux en utilisant la commande 'go build' avec les flags appropriés pour chaque OS. Utilisez les tags Git pour le versioning, en commençant par v0.1.0. Ajoutez une cible 'release' qui crée une archive ZIP contenant les binaires et la documentation."

### 14. Internationalisation

Instruction : "Dans internal/i18n/messages.go, définissez toutes les chaînes de caractères visibles par l'utilisateur en anglais. Utilisez ces constantes partout dans le code au lieu de chaînes de caractères codées en dur."

### 15. Finalisation et préparation à la release

Instruction : "Exécutez 'go fmt' et 'go vet' sur tout le code. Vérifiez que chaque fichier ne dépasse pas 3000 tokens, chaque fonction 80 lignes, et chaque package 10 méthodes publiques. Créez un tag Git v0.1.0 et poussez-le sur le dépôt distant. Générez un changelog détaillé des fonctionnalités implémentées."

## Vérification finale

Après avoir relu attentivement l'expression de besoins et ce plan d'action, je confirme que tous les éléments essentiels ont été couverts. Le plan d'action respecte les exigences initiales, intègre les modifications discutées (comme le traitement récursif des répertoires et la sélection des modèles LLM spécifiques), et couvre tous les aspects techniques et fonctionnels demandés.

Les points clés tels que la gestion des formats de documents, la segmentation, l'interaction avec différents LLM, la génération d'ontologies au format QuickStatement, les exports RDF et OWL, la gestion des métadonnées, et les contraintes techniques (taille des fichiers, structure des packages, etc.) sont tous adressés.

Le plan respecte également les exigences en termes de journalisation, configuration, tests, documentation, et internationalisation. Les aspects de sécurité, de performance et de distribution sont également pris en compte.

Si des ajustements ou des précisions supplémentaires sont nécessaires, n'hésitez pas à le mentionner.