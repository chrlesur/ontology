D'accord, j'ai pris en compte vos commentaires. Voici le plan d'action mis à jour intégrant les améliorations acceptées :

1. Créer un nouveau package "storage"
   - Créer le dossier `internal/storage`
   - Définir l'interface `Storage` avec les méthodes essentielles (Read, Write, List, Delete, Exists)

2. Implémenter le stockage local
   - Créer `internal/storage/local.go`
   - Adapter le code existant pour implémenter l'interface `Storage`

3. Implémenter le stockage S3
   - Créer `internal/storage/s3.go`
   - Implémenter l'interface `Storage` pour S3 en utilisant le SDK AWS Go v2
   - Permettre la configuration d'endpoints personnalisés (pour Dell ECS)

4. Créer un détecteur de type de stockage
   - Créer `internal/storage/detector.go`
   - Implémenter la logique de détection (local vs S3) basée sur le chemin d'entrée

5. Mettre à jour la configuration
   - Modifier `internal/config/config.go`
   - Ajouter les options pour S3 (bucket, région, endpoint, access_key_id, secret_access_key)
   - Ajouter une option pour forcer le type de stockage
   - Assurer que ces options sont chargées depuis le fichier YAML

6. Créer une factory pour le stockage
   - Créer `internal/storage/factory.go`
   - Implémenter une fonction pour créer le bon type de stockage basé sur la configuration et le chemin

7. Modifier le pipeline
   - Mettre à jour `internal/pipeline/pipeline.go`
   - Remplacer les opérations de fichiers directes par l'utilisation de l'interface `Storage`

8. Adapter les commandes CLI
   - Modifier `internal/ontology/enrich.go` et autres fichiers pertinents
   - Mettre à jour la logique pour utiliser la nouvelle abstraction de stockage
   - Ajouter des options CLI pour les paramètres S3 si nécessaire

9. Implémenter la gestion des URI S3
   - Créer une fonction utilitaire pour parser les URI S3 dans `internal/storage/utils.go`

10. Améliorer la gestion des erreurs
    - Créer `internal/storage/errors.go`
    - Implémenter des types d'erreurs spécifiques pour les problèmes de stockage S3
    - Ajouter une logique de retry pour les erreurs temporaires (timeouts, erreurs de connexion)

11. Intégrer des métriques et un logging amélioré
    - Modifier `internal/logger/logger.go` pour ajouter des niveaux de log spécifiques aux opérations de stockage
    - Implémenter des métriques de base (temps d'opération, taille des données transférées)
    - Intégrer ces métriques dans le système de logging existant

12. Implémenter un système de cache
    - Créer `internal/storage/cache.go`
    - Implémenter un cache en mémoire simple pour les opérations de lecture fréquentes
    - Intégrer le cache dans l'implémentation S3

13. Optimiser avec des opérations parallèles (si faisable simplement)
    - Modifier `internal/storage/s3.go` pour utiliser des goroutines lors des opérations de lecture/écriture en masse
    - Implémenter un pool de workers pour limiter la concurrence

14. Mettre à jour les tests
    - Créer des tests unitaires pour chaque nouvelle composante
    - Mettre à jour les tests existants si nécessaire
    - Créer des tests d'intégration pour S3
    - Ajouter des tests pour les nouvelles fonctionnalités (cache, gestion d'erreurs, parallélisation)

15. Mettre à jour la documentation
    - Mettre à jour le README avec les instructions pour utiliser S3
    - Documenter les nouvelles options de configuration dans le fichier YAML
    - Ajouter des exemples d'utilisation des nouvelles fonctionnalités

16. Refactoring et optimisation
    - Revoir l'ensemble du code pour s'assurer du respect des limites de taille (max 80 lignes par méthode)
    - Optimiser et nettoyer le code si nécessaire

Ce plan mis à jour intègre les améliorations demandées tout en restant dans le cadre des contraintes du projet. La parallélisation (point 13) est incluse de manière conditionnelle, à implémenter seulement si cela peut être fait simplement sans trop compliquer le code.