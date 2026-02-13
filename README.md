# Pick6 - Total Kombat Fight Predictions

Prediction game for Total Kombat fight events with prize draw entry.

## Stack

Go • chi • templ • PostgreSQL • sqlc • Docker

## Quick Start

```bash
make local-dev
```

Visit: http://localhost:8080/tk03

## Development

```bash
sqlc generate          # Generate SQL types
templ generate         # Generate templates
go build -o pick6      # Build binary
```

## Environment Variables

```env
DATABASE_URL=postgres://user:pass@localhost:5432/pick6?sslmode=disable
SECURE_COOKIE=false
PORT=8080
BASE_URL=http://localhost:8080
```

## Project Structure

```
handlers/    HTTP handlers
templates/   Templ templates
database/    SQL queries & migrations
middleware/  Session auth
static/      CSS & images
```

## API Examples

RESTful API designed for broadcast graphics systems. Poll endpoints every 1-2 seconds for live updates.

### Get Event Overview

```bash
# Using slug (recommended)
curl http://localhost:8080/api/events/tk03

# Using event ID
curl http://localhost:8080/api/events/event_39aJ1km3pr9v1yQYX5gS88e3CUM
```

Returns event metadata, all questions summary, and total engagement.

### Get All Questions (Initial Load)

```bash
# Get all questions with full engagement
curl http://localhost:8080/api/events/tk03/questions
```

Returns all questions with metadata, images, and engagement. Use for initial page load.

### Get Single Question (Live Polling)

**Primary endpoint for broadcast graphics - poll this every 1-2 seconds:**

```bash
# Using slug and index (recommended for broadcast)
curl http://localhost:8080/api/events/tk03/questions/1

# Using event ID and index
curl http://localhost:8080/api/events/event_39aJ1km3pr9v1yQYX5gS88e3CUM/questions/1

# Using question ID
curl http://localhost:8080/api/events/tk03/questions/question_39aJ1eE9ihQ3hH9kmOfKdCSueFP
```

Returns complete question data including:
- Question text and choices
- Full image URLs (ready to display)
- Vote counts (total and per slug)
- Vote percentages (calculated)
- Real-time engagement stats

## Features

Mobile-first • Database-backed sessions • Phone validation (E.164) • Public APIs • £1K VVIP prize draw
