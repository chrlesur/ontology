Sous-étape 5.2 : Implémentation de la conversion de base (convert.go)

Tâche : Créez le fichier convert.go dans le package internal/converter et implémentez la fonction Convert de l'interface Converter.

Instructions spécifiques :

1. Créez le fichier convert.go dans le package internal/converter.

2. Importez les packages nécessaires, notamment internal/logger, internal/i18n, et internal/config.

3. Implémentez la méthode Convert pour la structure QuickStatementConverter :
   func (qsc *QuickStatementConverter) Convert(segment []byte, context string, ontology string) (string, error)

4. Dans la méthode Convert, implémentez la logique suivante :
   a. Parsez le segment d'entrée en structures de données internes (utilisez les types définis dans quickstatement.go).
   b. Appliquez le contexte et l'ontologie fournis pour enrichir les données parsées.
   c. Convertissez les structures de données enrichies en format QuickStatement TSV.

5. Créez des fonctions helper privées pour chaque étape majeure du processus de conversion :
   - parseSegment(segment []byte) ([]Statement, error)
   - applyContextAndOntology(statements []Statement, context string, ontology string) ([]Statement, error)
   - toQuickStatementTSV(statements []Statement) (string, error)

6. Utilisez le package internal/logger pour enregistrer les étapes importantes du processus de conversion.

7. Utilisez le package internal/i18n pour tous les messages d'erreur ou de log.

8. Gérez les erreurs de manière appropriée, en utilisant error wrapping lorsque c'est pertinent.

9. Optimisez le code pour la performance, en particulier pour le traitement de grands segments.

10. Ajoutez des commentaires GoDoc pour chaque fonction exportée.

11. Assurez-vous que le fichier ne dépasse pas 200 lignes de code.

12. Créez un fichier convert_test.go et ajoutez des tests unitaires pour la méthode Convert et les fonctions helper.

Exemple de structure attendue pour la méthode Convert :

func (qsc *QuickStatementConverter) Convert(segment []byte, context string, ontology string) (string, error) {
    logger.Debug(i18n.GetMessage("ConvertStarted"))

    statements, err := qsc.parseSegment(segment)
    if err != nil {
        return "", fmt.Errorf("failed to parse segment: %w", err)
    }

    enrichedStatements, err := qsc.applyContextAndOntology(statements, context, ontology)
    if err != nil {
        return "", fmt.Errorf("failed to apply context and ontology: %w", err)
    }

    result, err := qsc.toQuickStatementTSV(enrichedStatements)
    if err != nil {
        return "", fmt.Errorf("failed to convert to QuickStatement TSV: %w", err)
    }

    logger.Debug(i18n.GetMessage("ConvertFinished"))
    return result, nil
}

// Implement helper functions: parseSegment, applyContextAndOntology, toQuickStatementTSV