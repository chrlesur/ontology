Tâche : Implémentation des clients LLM pour le projet Ontology

Directives générales (à suivre impérativement pour toutes les étapes du projet) :
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

Instructions spécifiques pour l'implémentation des clients LLM :

1. Créez un nouveau package 'internal/llm' avec les fichiers suivants :
   - client.go
   - openai.go
   - claude.go
   - ollama.go
   - factory.go

2. Dans client.go, définissez l'interface Client :
   ```go
   type Client interface {
       Translate(prompt string, context string) (string, error)
   }
   ```

3. Implémentez chaque client LLM dans son fichier respectif. Chaque implémentation doit :
   - Supporter les modèles spécifiés (OpenAI: 'gpt-3.5-turbo', 'gpt-4'; Claude: 'claude-3-opus-20240229', 'claude-3-sonnet-20240229'; Ollama: 'llama2', 'llama2:13b', 'llama2:70b')
   - Avoir une fonction NewXXXClient(apiKey string, model string) (*XXXClient, error)
   - Implémenter un système de retry avec backoff exponentiel (max 5 tentatives)
   - Respecter les limites de contexte spécifiques à chaque modèle

4. Dans factory.go, implémentez :
   ```go
   func GetClient(llmType string, apiKey string, model string) (Client, error)
   ```

5. Utilisez les constantes de 'internal/i18n' pour tous les messages d'erreur et de log.

6. Ajoutez des tests unitaires dans des fichiers *_test.go pour chaque client et la factory.

7. Assurez-vous que chaque fichier ne dépasse pas 200 lignes de code.

8. Utilisez le package 'internal/config' pour toutes les valeurs configurables (comme les URL d'API).

9. Implémentez une gestion robuste des erreurs, en utilisant error wrapping et en loggant les erreurs appropriées.

10. Optimisez les clients pour la performance, en particulier pour le traitement de grands volumes de données.

Exemple de structure pour openai.go (à adapter pour les autres clients) :

```go
package llm

import (
    "github.com/sashabaranov/go-openai"
    "github.com/chrlesur/Ontology/internal/logger"
    "github.com/chrlesur/Ontology/internal/i18n"
    "github.com/chrlesur/Ontology/internal/config"
)

type OpenAIClient struct {
    client *openai.Client
    model  string
}

func NewOpenAIClient(apiKey string, model string) (*OpenAIClient, error) {
    // Implémentation...
}

func (c *OpenAIClient) Translate(prompt string, context string) (string, error) {
    // Implémentation avec gestion des retries...
}

// Implémentez d'autres méthodes nécessaires...
```

Assurez-vous que le code compile sans erreur, que tous les tests passent, et que vous respectez toutes les directives générales avant de soumettre votre travail.
