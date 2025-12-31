# Query Parameters Usage Guide

## Available Query Parameters

### Common Parameters (All Endpoints)
- `take` - Number of records to return (default: 10, max: 100)
- `skip` - Number of records to skip (default: 0)
- `search` - Search term (searches in name and description fields)
- `beginning` - Date filter in ISO 8601 format (e.g., `2024-01-01T00:00:00Z`)

### Category-Specific
- `parent_id` - Filter by parent category ID

## Usage Examples

### Products Endpoint

```bash
# Basic pagination
GET /products?take=10&skip=0

# Search with pagination
GET /products?search=laptop&take=20&skip=0

# Date filter - products created after a specific date
GET /products?beginning=2024-01-01T00:00:00Z&take=10&skip=0

# Combined filters
GET /products?search=electronics&beginning=2024-01-01T00:00:00Z&take=15&skip=30
```

### Categories Endpoint

```bash
# Basic pagination
GET /categories?take=10&skip=0

# Search categories
GET /categories?search=electronics&take=10&skip=0

# Get subcategories (filter by parent)
GET /categories?parent_id=1&take=10&skip=0

# Get top-level categories (no parent_id specified)
GET /categories?take=10&skip=0

# Date filter
GET /categories?beginning=2024-01-01T00:00:00Z&take=10&skip=0
```

## Response Format

```json
{
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Laptop",
      "price": 999.99,
      ...
    }
  ],
  "pagination": {
    "take": 10,
    "skip": 0,
    "total": 150
  }
}
```

## Implementation Details

- **Search**: Uses `ILIKE` for case-insensitive pattern matching on `name` and `description` fields
- **Date Filter**: Filters records where `created_at >= beginning`
- **Pagination**: Offset-based using `skip` and `take`
- **Sorting**: 
  - Categories: Ordered by `display_order ASC, created_at DESC`
  - Products: Ordered by `created_at DESC`

