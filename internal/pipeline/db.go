package pipeline

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/chrlesur/Ontology/internal/model"
	_ "modernc.org/sqlite"
)

func initDB() (*sql.DB, error) {
	log.Debug("Initializing in-memory SQLite database")
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Error("Failed to open database: %v", err)
		return nil, fmt.Errorf("échec de l'ouverture de la base de données : %w", err)
	}

	// Création de la table des entités et de l'index
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS entities (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            type TEXT NOT NULL,
            description TEXT,
            positions TEXT,  
            created_at DATETIME,
            updated_at DATETIME,
            source TEXT
        );
        CREATE INDEX IF NOT EXISTS idx_entities_name ON entities(name);
    `)
	if err != nil {
		log.Error("Failed to create entities table or index: %v", err)
		return nil, fmt.Errorf("échec de la création de la table entities ou de l'index : %w", err)
	}
	log.Debug("Entities table and index created successfully")

	// Création de la table des relations et des index
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS relations (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            source TEXT NOT NULL,
            type TEXT NOT NULL,
            target TEXT NOT NULL,
            description TEXT,
            weight REAL,
            direction TEXT,
            created_at DATETIME,
            updated_at DATETIME,
            UNIQUE(source, type, target)
        );
        CREATE INDEX IF NOT EXISTS idx_relations_source ON relations(source);
        CREATE INDEX IF NOT EXISTS idx_relations_target ON relations(target);
    `)
	if err != nil {
		log.Error("Failed to create relations table or indexes: %v", err)
		return nil, fmt.Errorf("échec de la création de la table relations ou des index : %w", err)
	}
	log.Debug("Relations table and indexes created successfully")

	log.Debug("Database initialized successfully")
	return db, nil
}

func UpsertEntity(db *sql.DB, entity *model.OntologyElement) error {
    log.Debug("Upserting entity: %+v", entity)
    positionsJSON, err := json.Marshal(entity.Positions)
    if err != nil {
        return fmt.Errorf("failed to marshal positions: %w", err)
    }
    _, err = db.Exec(`
        INSERT INTO entities (name, type, description, positions, created_at, updated_at, source)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(name) DO UPDATE SET
        type = ?,
        description = ?,
        positions = ?,
        updated_at = ?,
        source = ?
    `, entity.Name, entity.Type, entity.Description, positionsJSON, entity.CreatedAt, entity.UpdatedAt, entity.Source,
       entity.Type, entity.Description, positionsJSON, entity.UpdatedAt, entity.Source)
    if err != nil {
        log.Error("Failed to upsert entity: %v", err)
        return fmt.Errorf("failed to upsert entity: %w", err)
    }
    log.Debug("Entity upserted successfully with positions: %v", entity.Positions)
    return nil
}

func UpsertRelation(db *sql.DB, relation *model.Relation) error {
    log.Debug("Upserting relation: %+v", relation)
    _, err := db.Exec(`
        INSERT INTO relations (source, type, target, description, weight, direction, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(source, type, target) DO UPDATE SET
        description = ?,
        weight = ?,
        direction = ?,
        updated_at = ?
    `, relation.Source, relation.Type, relation.Target, relation.Description, relation.Weight, relation.Direction,
        relation.CreatedAt, relation.UpdatedAt,
        relation.Description, relation.Weight, relation.Direction, relation.UpdatedAt)
    if err != nil {
        log.Error("Failed to upsert relation: %v", err)
        return fmt.Errorf("failed to upsert relation: %w", err)
    }
    return nil
}

func GetAllEntities(db *sql.DB) ([]*model.OntologyElement, error) {
	log.Debug("Starting GetAllEntities")

	rows, err := db.Query("SELECT name, type, description, positions, created_at, updated_at, source FROM entities")
	if err != nil {
		log.Error("Failed to query entities: %v", err)
		return nil, fmt.Errorf("échec de la récupération des entités : %w", err)
	}
	defer rows.Close()

	var entities []*model.OntologyElement
	for rows.Next() {
		var e model.OntologyElement
		var positionsJSON []byte
        err := rows.Scan(&e.Name, &e.Type, &e.Description, &positionsJSON, &e.CreatedAt, &e.UpdatedAt, &e.Source)
		if err != nil {
			log.Error("Failed to scan entity: %v", err)
			return nil, fmt.Errorf("échec du scan d'une entité : %w", err)
		}

		log.Debug("Retrieved entity from DB: Name=%s, Type=%s, Description=%s, PositionsJSON=%s",
			e.Name, e.Type, e.Description, string(positionsJSON))

		if len(positionsJSON) > 0 {
			err = json.Unmarshal(positionsJSON, &e.Positions)
			if err != nil {
				log.Error("Failed to unmarshal positions for entity %s: %v", e.Name, err)
				return nil, fmt.Errorf("échec du décodage des positions pour l'entité %s : %w", e.Name, err)
			}
			log.Debug("Unmarshalled positions for entity %s: %v", e.Name, e.Positions)
		} else {
			log.Debug("No positions found for entity %s", e.Name)
		}

		entities = append(entities, &e)
	}

	log.Debug("Retrieved %d entities in total", len(entities))
	return entities, nil
}

func GetAllRelations(db *sql.DB) ([]*model.Relation, error) {
	log.Debug("Getting all relations")
	rows, err := db.Query("SELECT source, type, target, description, weight, direction, created_at, updated_at FROM relations")
	if err != nil {
		log.Error("Failed to query relations: %v", err)
		return nil, fmt.Errorf("échec de la récupération des relations : %w", err)
	}
	defer rows.Close()

	var relations []*model.Relation
	for rows.Next() {
		var r model.Relation
		err := rows.Scan(&r.Source, &r.Type, &r.Target, &r.Description, &r.Weight, &r.Direction, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			log.Error("Failed to scan relation: %v", err)
			return nil, fmt.Errorf("échec du scan d'une relation : %w", err)
		}
		log.Debug("Retrieved relation: %+v", r)
		relations = append(relations, &r)
	}
	log.Debug("Total relations retrieved: %d", len(relations))
	return relations, nil
}
