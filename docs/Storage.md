# Design
Article service implements it's storage layer using Command Query Responsibility Segregation and Event Sourcing patterns.

Writes are stored as events in EventStore and eventually service replays these events on Redis which is used for queries. 

## EventStore
Event data schema:
```jsonc
{   
    "aggregate_id": "string", // Hardcoded to 'article'
    "type": "string", // Eg. 'article-created'
    "body": {
        "id": "string", // UUID
        "user_id": "string", // UUID
        "title": "string",
        "body": "string",
        "created_at": "string", // RFC3339 format
        "updated_at": "string" // RFC3339 format
    },
    "timestamp": "string" // RFC3339 format
}
```

## Redis
Articles are saved in hash maps where a key for each map is `article:<id>`.
Article IDs are also saved in a separate string set for faster lookups when quering multiple articles.

Each user has its own string set containing IDs of all articles created by that user. Keys for these sets are `users:<userId>`.

Article hash map schema:
```jsonc
{
	"id": "string", // UUID
	"user_id": "string", // UUID
	"title": "string",
	"body": "string",
	"created_at": "string", // RFC3339 format
	"updated_at": "string" // RFC3339 format
}
```
