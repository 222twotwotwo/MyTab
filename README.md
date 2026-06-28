# MyTab

MyTab is a Gin + Vue TypeScript app for building a personalized browser tab page.

Users can upload a static web page and its resources, store them on the server, and apply the page to the current MyTab screen. Uploaded files are only stored and served as static files. The server does not execute user HTML, CSS, or JavaScript.

## Features

- Upload a static page folder with HTML, CSS, JavaScript, images, fonts, media, and other static resources.
- Preserve relative resource paths so pages can reference local assets.
- Save multiple uploaded tab pages on the Gin server.
- Apply one saved page to the current browser page.
- Pick a built-in template from the style market and apply it as the current tab page.
- Render the applied page in an iframe sandbox with a restrictive CSP.

## One-Click Scripts

Start:

```powershell
.\start.bat
```

Stop residual server processes:

```powershell
.\end.bat
```

Open `http://localhost:8080`.

## Manual Run

Start the API:

```powershell
cd server
go run -buildvcs=false ./cmd/server
```

Start the web app during development:

```powershell
cd web
npm install
npm run dev
```

Open `http://localhost:5173`.

## Build

```powershell
cd web
npm run build

cd ..\server
go build -buildvcs=false ./cmd/server
```

After the frontend is built, the Gin server can serve `web/dist` when launched from `server` with:

```powershell
go run -buildvcs=false ./cmd/server
```

## API

- `GET /healthz`
- `GET /api/projects`
- `POST /api/projects`
- `GET /api/projects/:id`
- `GET /api/projects/:id/files/:file`
- `GET /api/projects/:id/preview/:file`

The preview route only serves uploaded files. It does not execute them on the server.
