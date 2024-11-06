
Étape 11 : Implémentation de l'interopérabilité et de l'API

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
Tâche spécifique : Création d'une API RESTful pour faciliter l'intégration avec d'autres outils ou workflows

1. Créez un nouveau package internal/api avec un fichier server.go :

```go
package api

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/chrlesur/Ontology/internal/pipeline"
    "github.com/chrlesur/Ontology/internal/logger"
    "github.com/chrlesur/Ontology/internal/config"
)

type Server struct {
    router   *mux.Router
    pipeline *pipeline.Pipeline
    logger   *logger.Logger
    config   *config.Config
}

func NewServer() (*Server, error) {
    p, err := pipeline.NewPipeline()
    if err != nil {
        return nil, err
    }

    s := &Server{
        router:   mux.NewRouter(),
        pipeline: p,
        logger:   logger.GetLogger(),
        config:   config.GetConfig(),
    }

    s.routes()
    return s, nil
}

func (s *Server) routes() {
    s.router.HandleFunc("/api/v1/process", s.handleProcess()).Methods("POST")
    s.router.HandleFunc("/api/v1/status", s.handleStatus()).Methods("GET")
}

func (s *Server) handleProcess() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var input struct {
            Content string `json:"content"`
            Passes  int    `json:"passes"`
        }

        if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
            s.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
            return
        }

        result, err := s.pipeline.ExecutePipeline(input.Content, input.Passes)
        if err != nil {
            s.respondWithError(w, http.StatusInternalServerError, err.Error())
            return
        }

        s.respondWithJSON(w, http.StatusOK, map[string]string{"result": result})
    }
}

func (s *Server) handleStatus() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        s.respondWithJSON(w, http.StatusOK, map[string]string{"status": "operational"})
    }
}

func (s *Server) respondWithError(w http.ResponseWriter, code int, message string) {
    s.respondWithJSON(w, code, map[string]string{"error": message})
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, _ := json.Marshal(payload)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.router.ServeHTTP(w, r)
}
```

2. Créez un fichier main.go dans cmd/api pour lancer le serveur API :

```go
package main

import (
    "log"
    "net/http"
    "github.com/chrlesur/Ontology/internal/api"
    "github.com/chrlesur/Ontology/internal/config"
)

func main() {
    cfg := config.GetConfig()
    server, err := api.NewServer()
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    log.Printf("Starting server on %s", cfg.APIAddress)
    log.Fatal(http.ListenAndServe(cfg.APIAddress, server))
}
```

3. Mettez à jour internal/config/config.go pour inclure la configuration de l'API :

```go
type Config struct {
    // ... autres champs existants ...
    APIAddress string `yaml:"api_address"`
}
```

4. Ajoutez des tests unitaires pour l'API dans un fichier api_test.go :

```go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHandleProcess(t *testing.T) {
    server, _ := NewServer()

    input := map[string]interface{}{
        "content": "Test content",
        "passes":  1,
    }
    body, _ := json.Marshal(input)

    req, _ := http.NewRequest("POST", "/api/v1/process", bytes.NewBuffer(body))
    rr := httptest.NewRecorder()

    server.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    // Add more assertions as needed
}

// Add more test functions for other endpoints
```

5. Implémentez un système de versioning d'API simple :

   a. Ajoutez une constante pour la version actuelle de l'API dans internal/api/server.go :

   ```go
   const APIVersion = "v1"
   ```

   b. Utilisez cette constante dans la définition des routes :

   ```go
   func (s *Server) routes() {
       s.router.HandleFunc("/api/"+APIVersion+"/process", s.handleProcess()).Methods("POST")
       s.router.HandleFunc("/api/"+APIVersion+"/status", s.handleStatus()).Methods("GET")
   }
   ```

6. Ajoutez une documentation Swagger pour l'API :

   a. Installez go-swagger : `go get -u github.com/go-swagger/go-swagger/cmd/swagger`
   
   b. Ajoutez des commentaires Swagger dans internal/api/server.go :

   ```go
   // swagger:operation POST /api/v1/process processContent
   // ---
   // summary: Process content and generate ontology
   // parameters:
   // - name: input
   //   in: body
   //   description: Content to process
   //   required: true
   //   schema:
   //     "$ref": "#/definitions/ProcessInput"
   // responses:
   //   '200':
   //     description: Successful operation
   //     schema:
   //       "$ref": "#/definitions/ProcessOutput"
   func (s *Server) handleProcess() http.HandlerFunc {
       // ... implementation ...
   }
   ```

   c. Générez la spécification Swagger : `swagger generate spec -o ./swagger.json`

7. Implémentez un middleware pour la gestion des erreurs et la journalisation des requêtes :

```go
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        s.logger.Info("Request received: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

func (s *Server) routes() {
    s.router.Use(s.loggingMiddleware)
    // ... définition des routes ...
}
```

8. Assurez-vous que toutes les fonctions respectent les limites de taille (pas plus de 80 lignes par fonction).

9. Documentez toutes les fonctions exportées avec des commentaires GoDoc.

10. Mettez à jour le README.md pour inclure des informations sur l'utilisation de l'API.

Après avoir terminé ces tâches, exécutez tous les tests unitaires et assurez-vous que le code compile sans erreur. Testez l'API manuellement pour vérifier qu'elle fonctionne correctement et qu'elle s'intègre bien avec le reste du système.
