En tant que développeur Go expérimenté, votre tâche est de développer les analyseurs de documents pour le projet Ontology. Voici vos directives :

Directives générales (à suivre pour toutes les étapes du projet) :
1. Utilisez exclusivement Go dans sa dernière version stable.
2. Assurez-vous qu'aucun fichier de code source ne dépasse 3000 tokens.
3. Limitez chaque package à un maximum de 10 méthodes exportées.
4. Aucune méthode ne doit dépasser 80 lignes de code.
5. Suivez les meilleures pratiques et les modèles idiomatiques de Go.
6. Tous les messages visibles par l'utilisateur doivent être en anglais.
7. Chaque fonction, méthode et type exporté doit avoir un commentaire de documentation conforme aux standards GoDoc.
8. Utilisez le package 'internal/logger' pour toute journalisation. Implémentez les niveaux de log : debug, info, warning, et error.
9. Toutes les valeurs configurables doivent être définies dans le package 'internal/config'.
10. Gérez toutes les erreurs de manière appropriée, en utilisant error wrapping lorsque c'est pertinent.
11. Pour les messages utilisateur, utilisez les constantes définies dans le package 'internal/i18n'.
12. Assurez-vous que le code est prêt pour de futurs efforts de localisation.
13. Optimisez le code pour la performance, particulièrement pour le traitement de grands documents.
14. Implémentez des tests unitaires pour chaque nouvelle fonction ou méthode.
15. Veillez à ce que le code soit sécurisé, en particulier lors du traitement des entrées utilisateur.

Instructions spécifiques pour l'étape 3 - Développement des analyseurs de documents :

1. Dans internal/parser, créez une interface Parser avec une seule méthode Parse(path string) ([]byte, error).

2. Implémentez cette interface pour chaque format supporté dans des fichiers séparés :
   - text.go
   - pdf.go
   - markdown.go
   - html.go
   - docx.go

3. Pour chaque implémentation, utilisez uniquement la bibliothèque standard Go ou une bibliothèque tierce spécifiée. Pour PDF, utilisez 'github.com/unidoc/unipdf/v3'.

4. Ajoutez une fonction GetMetadata() map[string]string à chaque parser pour extraire les métadonnées du document.

5. Créez une fonction ParseDirectory(path string, recursive bool) ([][]byte, error) qui parcourt un répertoire (et ses sous-répertoires si recursive est true) et parse tous les fichiers supportés.

6. Implémentez une gestion robuste des erreurs pour les documents mal formés ou incomplets, en utilisant le système de journalisation.

7. Optimisez les parsers pour gérer efficacement les très grands documents (dépassant 120 000 tokens).

8. Assurez-vous que chaque parser préserve la structure du document autant que possible.

9. Implémentez des tests unitaires pour chaque parser et la fonction ParseDirectory.

10. Utilisez des goroutines et des canaux de manière appropriée pour le traitement parallèle dans ParseDirectory, si cela peut améliorer les performances.

11. Ajoutez des logs appropriés à chaque étape du parsing, en utilisant les niveaux de log appropriés.

12. Créez une factory function GetParser(format string) (Parser, error) qui retourne le parser approprié basé sur le format spécifié.

13. Assurez-vous que tous les messages d'erreur et les logs sont définis dans internal/i18n et utilisés de manière cohérente.

14. Documentez chaque fonction et méthode avec des commentaires GoDoc détaillés, y compris des exemples d'utilisation si nécessaire.

Veillez à ce que le code soit bien structuré, performant et robuste. Une fois terminé, assurez-vous que tous les parsers fonctionnent correctement avec différents types et tailles de documents.