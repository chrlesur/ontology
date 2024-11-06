Étape 7 : Développement du système de journalisation

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

Tâche spécifique : Implémentation du système de journalisation

1. Dans le package internal/logger, créez ou mettez à jour le fichier logger.go avec le contenu suivant :

```go
package logger

import (
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "runtime"
    "sync"
    "time"

    "github.com/chrlesur/Ontology/internal/config"
    "github.com/chrlesur/Ontology/internal/i18n"
)

type LogLevel int

const (
    DebugLevel LogLevel = iota
    InfoLevel
    WarningLevel
    ErrorLevel
)

var (
    instance *Logger
    once     sync.Once
)

type Logger struct {
    level  LogLevel
    logger *log.Logger
    file   *os.File
}

func GetLogger() *Logger {
    once.Do(func() {
        instance = &Logger{
            level:  InfoLevel,
            logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
        }
        instance.setupLogFile()
    })
    return instance
}

func (l *Logger) setupLogFile() {
    logDir := config.GetConfig().LogDirectory
    if logDir == "" {
        logDir = "logs"
    }
    
    if err := os.MkdirAll(logDir, 0755); err != nil {
        l.Error(i18n.GetMessage("ErrCreateLogDir"), err)
        return
    }

    logFile := filepath.Join(logDir, fmt.Sprintf("ontology_%s.log", time.Now().Format("2006-01-02")))
    file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        l.Error(i18n.GetMessage("ErrOpenLogFile"), err)
        return
    }

    l.file = file
    l.logger.SetOutput(io.MultiWriter(os.Stdout, file))
}

func (l *Logger) SetLevel(level LogLevel) {
    l.level = level
}

func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
    if level < l.level {
        return
    }

    _, file, line, _ := runtime.Caller(2)
    logMessage := fmt.Sprintf("[%s] %s:%d - %s", level, filepath.Base(file), line, fmt.Sprintf(message, args...))
    l.logger.Println(logMessage)
}

func (l *Logger) Debug(message string, args ...interface{}) {
    l.log(DebugLevel, message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
    l.log(InfoLevel, message, args...)
}

func (l *Logger) Warning(message string, args ...interface{}) {
    l.log(WarningLevel, message, args...)
}

func (l *Logger) Error(message string, args ...interface{}) {
    l.log(ErrorLevel, message, args...)
}

func (l *Logger) Close() {
    if l.file != nil {
        l.file.Close()
    }
}
```

2. Créez un fichier logger_test.go dans le même package et implémentez des tests unitaires pour toutes les fonctions du logger.

3. Mettez à jour le fichier internal/config/config.go pour inclure la configuration du logger :

```go
type Config struct {
    // ... autres champs existants ...
    LogDirectory string `yaml:"log_directory"`
    LogLevel     string `yaml:"log_level"`
}
```

4. Dans le fichier internal/i18n/messages.go, ajoutez les constantes nécessaires pour les messages du logger :

```go
const (
    // ... autres constantes existantes ...
    ErrCreateLogDir = "Failed to create log directory"
    ErrOpenLogFile  = "Failed to open log file"
)
```

5. Mettez à jour tous les autres packages du projet pour utiliser ce nouveau système de journalisation. Par exemple, dans internal/llm/openai.go :

```go
import (
    "github.com/chrlesur/Ontology/internal/logger"
)

func (c *OpenAIClient) Translate(prompt string, context string) (string, error) {
    log := logger.GetLogger()
    log.Debug("Starting translation with OpenAI")
    // ... reste de l'implémentation ...
}
```

6. Assurez-vous que la rotation des logs est gérée correctement. Implémentez une fonction pour archiver les anciens logs si nécessaire.

7. Ajoutez une fonction UpdateProgress dans logger.go pour mettre à jour la dernière ligne de la console avec la progression :

```go
func (l *Logger) UpdateProgress(current, total int) {
    fmt.Printf("\rProgress: %d/%d", current, total)
}
```

8. Vérifiez que le logger respecte les modes silencieux (--silent) et debug (--debug) définis dans les flags de la commande.

9. Assurez-vous que tous les fichiers respectent les limites de taille (pas plus de 200 lignes, pas plus de 80 lignes par fonction).

10. Documentez toutes les fonctions exportées avec des commentaires GoDoc.

11. Implémentez des tests de performance pour s'assurer que le logger n'introduit pas de ralentissement significatif.

Après avoir terminé ces tâches, exécutez tous les tests unitaires et assurez-vous que le code compile sans erreur. Vérifiez également que le logger fonctionne correctement dans différents scénarios (debug, info, warning, error) et que les fichiers de log sont correctement créés et formatés.
