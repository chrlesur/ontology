# API Documentation

Ontology does not currently provide a RESTful API. It is designed as a command-line tool. However, the internal structure allows for potential API development in the future.

## Potential API Endpoints

If an API were to be implemented, it might include the following endpoints:

1. POST /ontology/enrich
   - Description: Enrich an ontology from input documents
   - Request body: 
     - input: string (file path or content)
     - format: string (optional, file format)
     - llm: string (optional, LLM provider)
     - passes: integer (optional, number of enrichment passes)
   - Response: Enriched ontology in QuickStatement format

2. GET /ontology/export
   - Description: Export an ontology in different formats
   - Query parameters:
     - format: string (rdf, owl)
   - Response: Ontology in the specified format

3. GET /config
   - Description: Retrieve current configuration
   - Response: Current configuration in JSON format

4. PUT /config
   - Description: Update configuration
   - Request body: Updated configuration in JSON format
   - Response: Updated configuration

Note: These are conceptual endpoints and are not currently implemented.