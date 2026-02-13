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

### Get Event Engagement

Accepts either event ID or slug:

```bash
# Using event ID
curl http://localhost:8080/api/events/event_39aJ1km3pr9v1yQYX5gS88e3CUM/engagement

# Using slug (easier!)
curl http://localhost:8080/api/events/tk03/engagement
```

Returns event-level metrics including total sessions, votes, and breakdown by slug with retention data.

### Get Question Engagement

Accepts either event ID or slug, and either question ID or index (1-based):

```bash
# Using event slug and question index (easiest!)
curl http://localhost:8080/api/events/tk03/questions/1/engagement

# Using event ID and question index
curl http://localhost:8080/api/events/event_39aJ1km3pr9v1yQYX5gS88e3CUM/questions/1/engagement

# Using event slug and question ID
curl http://localhost:8080/api/events/tk03/questions/question_39aJ1eE9ihQ3hH9kmOfKdCSueFP/engagement

# Using event ID and question ID
curl http://localhost:8080/api/events/event_39aJ1km3pr9v1yQYX5gS88e3CUM/questions/question_39aJ1eE9ihQ3hH9kmOfKdCSueFP/engagement
```

Returns question-level metrics including votes for each option (A/B) and breakdown by slug.

## Features

Mobile-first • Database-backed sessions • Phone validation (E.164) • Public APIs • £1K VVIP prize draw
