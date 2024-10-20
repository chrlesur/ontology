En tant que développeur Go expérimenté, votre tâche est d'implémenter la segmentation des documents pour le projet Ontology. Voici vos directives :

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

En tant que développeur Go expérimenté, votre tâche est d'implémenter la segmentation des documents pour le projet Ontology. Voici vos directives :

Instructions spécifiques pour l'étape 4 - Implémentation de la segmentation des documents :

1. Dans internal/segmenter/segmenter.go, créez une fonction Segment(content []byte, maxTokens int) ([][]byte, error) qui divise le contenu en segments de maxTokens.

2. Implémentez un comptage précis des tokens en utilisant l'algorithme de tokenization GPT. Vous pouvez utiliser une bibliothèque existante comme "github.com/pkoukk/tiktoken-go" pour une tokenization précise compatible avec les modèles GPT.

3. Assurez-vous que chaque segment se termine par une phrase complète. Ne coupez jamais au milieu d'une phrase.

4. Implémentez une fonction GetContext(segments [][]byte, currentIndex int) string qui retourne le contexte des segments précédents. Cette fonction devrait retourner un résumé ou les dernières phrases des segments précédents, sans dépasser une limite de tokens définie (par exemple, 500 tokens).

5. Créez une fonction CountTokens(content []byte) int qui retourne le nombre exact de tokens dans un contenu donné, en utilisant l'algorithme de tokenization GPT.

6. Implémentez une gestion efficace de la mémoire pour traiter de très grands documents (dépassant 120 000 tokens).

7. Ajoutez des logs appropriés à chaque étape de la segmentation, en utilisant les niveaux de log appropriés.

8. Créez des tests unitaires pour toutes les fonctions, y compris des cas de test pour des documents de différentes tailles et structures, en vérifiant la précision du comptage des tokens.

9. Optimisez les performances en utilisant des techniques comme le buffering et le streaming pour les grands documents, tout en maintenant la précision du comptage des tokens.

10. Assurez-vous que la segmentation préserve les métadonnées importantes du document original (par exemple, les en-têtes de section).

11. Implémentez une fonction MergeSegments(segments [][]byte) []byte pour reconstituer le document original si nécessaire.

12. Documentez chaque fonction avec des commentaires GoDoc détaillés, y compris des exemples d'utilisation et des explications sur la méthode de tokenization utilisée.

13. Créez une structure de configuration dans internal/config pour stocker les paramètres de segmentation (comme maxTokens, contextSize, etc.).

14. Assurez-vous que tous les messages d'erreur et les logs sont définis dans internal/i18n et utilisés de manière cohérente.

15. Implémentez une gestion des erreurs robuste, en particulier pour les cas où le document ne peut pas être segmenté correctement.

16. Ajoutez une fonction pour calibrer le comptage des tokens en fonction du modèle LLM spécifique utilisé (par exemple, GPT-3.5, GPT-4, etc.), car différents modèles peuvent avoir des méthodes de tokenization légèrement différentes.

Veillez à ce que le code soit efficace, bien structuré et capable de gérer une variété de formats de documents et de tailles, tout en assurant un comptage précis des tokens. Une fois terminé, testez rigoureusement la segmentation avec différents types de documents pour assurer sa fiabilité, sa précision et sa performance.