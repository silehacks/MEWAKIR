# Mewakir

Mewakir is a lightweight face-to-face meeting experience inspired by Google Meet. It combines a Go backend with MongoDB persistence and a modern React frontend to let people create meetings, share join credentials, and talk over live audio and video.

## Features

- Name-based access flow with a friendly login screen
- Create meetings that generate random IDs and passcodes stored in MongoDB
- Join existing meetings via ID and passcode validation
- WebRTC audio/video calling with mute and camera toggles
- Minimal, modern UI with a responsive layout

## Architecture Overview

The project is split into two top-level packages:

- `backend/` – Go HTTP API with WebSocket signaling for WebRTC peers
- `frontend/` – React single-page app built with Vite

### Backend

The Go backend exposes JSON APIs and a WebSocket endpoint.

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/login` | POST | Registers a display name (no auth persistence yet). |
| `/api/meetings` | POST | Creates a new meeting with generated ID and passcode. |
| `/api/meetings/join` | POST | Validates meeting ID/passcode combinations. |
| `/ws/signaling` | GET (WS) | WebSocket for WebRTC signaling messages. |

Meetings are persisted inside MongoDB (configured via `MONGODB_URI` and `MONGODB_DB`). The WebSocket hub broadcasts signaling payloads to all peers inside the same meeting.

### Frontend

The React app offers three primary routes: login, home, and meeting. It stores the visitor name in localStorage, lets people create or join meetings, and manages WebRTC connections via a custom hook that exchanges signaling messages over the backend WebSocket.

## Getting Started

### Prerequisites

- Go 1.21+
- MongoDB instance (local or Atlas) reachable via connection string
- Node.js 18+ and pnpm/npm/yarn (for the frontend)

### Backend setup

```bash
cd backend
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_DB="mewakir"
go run .
```

The server listens on port `8080` by default. Configure an alternative port with the `PORT` environment variable.

### Frontend setup

```bash
cd frontend
npm install
npm run dev
```

The Vite dev server runs on `http://localhost:5173` and proxies API and WebSocket calls to the Go backend.

### Production build

```bash
cd frontend
npm run build
npm run preview
```

## Environment Variables

| Variable | Description | Default |
| --- | --- | --- |
| `MONGODB_URI` | Connection string used by the Go backend | `mongodb://localhost:27017` |
| `MONGODB_DB` | MongoDB database name | `mewakir` |
| `PORT` | Port for the Go HTTP server | `8080` |

## Future Improvements

- Persistent user accounts and authentication
- In-meeting chat and participant list
- Screen sharing, recording, and moderation controls
- Improved handling of multi-party calls and ICE restarts
