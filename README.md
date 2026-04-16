# F1 Tracker

Aplicacion de escritorio para trackear tiempos de F1 entre amigos usando Google Sheets como backend. Construida con [Fyne](https://fyne.io/) para la UI y la API REST de Google Sheets v4 para leer/escribir datos.

## Requisitos

- Go 1.25.9+
- Credenciales OAuth2 de Google Cloud Console
- Un spreadsheet de Google Sheets (o crear uno nuevo desde la app)

## Configurar credenciales de Google

1. Ir a [Google Cloud Console](https://console.cloud.google.com/)
2. Crear un proyecto (o usar uno existente)
3. Habilitar la **Google Sheets API** en APIs & Services > Enabled APIs
4. Ir a **APIs & Services > Credentials** y crear un **OAuth 2.0 Client ID**
   - Tipo: **Desktop app** o **Web application**
   - Si es Web application, agregar `http://localhost:8081/callback` como Authorized Redirect URI
5. Copiar el **Client ID** y **Client Secret**

## Correr en modo desarrollo

```bash
go run ./cmd
```

Al ejecutar por primera vez, la app muestra una pantalla de configuracion donde se ingresan:

| Campo               | Descripcion                                                    |
|---------------------|----------------------------------------------------------------|
| Google Client ID    | Client ID de OAuth2 (debe terminar en `.apps.googleusercontent.com`) |
| Google Client Secret| Client Secret de OAuth2                                        |
| Puerto callback     | Puerto del servidor local para el callback de OAuth (default: `8081`) |
| Spreadsheet ID      | ID del spreadsheet de Google Sheets (se encuentra en la URL)   |

Despues de guardar la configuracion, la app redirige a la pantalla de login donde se abre el navegador para autorizar con Google.

## Integracion con Google Sheets

La app usa la API REST de Google Sheets v4 para:

- **Leer metadata del spreadsheet** — obtiene la lista de hojas/tabs existentes
- **Crear hojas** — agrega tabs nuevos via `batchUpdate` (uno por usuario)
- **Escribir datos** — escribe headers y tiempos de vuelta en las hojas

Cada usuario tiene su propia hoja/tab en el spreadsheet con el siguiente formato:

| Circuito | Mejor Vuelta | Mejor S1 | Mejor S2 | Mejor S3 | S1 Vuelta | S2 Vuelta | S3 Vuelta | Auto | Fecha |
|----------|-------------|-----------|-----------|-----------|-----------|-----------|-----------|------|-------|

Al iniciar sesion, la app verifica si la hoja del usuario existe. Si no existe, la crea automaticamente con los headers.

## Donde se guardan las credenciales y tokens

La app usa `os.UserConfigDir()` para determinar el directorio de configuracion del sistema operativo y guarda todo en un archivo JSON:

| Sistema operativo | Ruta del archivo                                                    |
|-------------------|---------------------------------------------------------------------|
| **macOS**         | `~/Library/Application Support/f1-tracker/config-f1-tracker.json`   |
| **Windows**       | `%AppData%\f1-tracker\config-f1-tracker.json`                       |
| **Linux**         | `~/.config/f1-tracker/config-f1-tracker.json`                       |

### Estructura del archivo de configuracion

```json
{
  "config": {
    "GoogleClientID": "123456.apps.googleusercontent.com",
    "GoogleClientSecret": "GOCSPX-...",
    "CallbackPort": "8081",
    "SpreadsheetID": "1ABC...",
    "Username": "JoniDG"
  },
  "token": {
    "access_token": "ya29...",
    "token_type": "Bearer",
    "refresh_token": "1//...",
    "expiry": "2026-04-14T04:42:02Z"
  }
}
```

> **Nota:** Este archivo contiene credenciales sensibles. No lo commitees ni lo compartas.

## Compilar

```bash
go build -o f1-tracker ./cmd
```

## Tests

```bash
go test ./...
```

## Arquitectura

```
cmd/                  # Punto de entrada (main.go)
internal/
  domain/             # Structs del dominio (User, Config, Track, Spreadsheet)
  repository/         # Acceso a datos (config local, Google Sheets API, Google userinfo API)
  service/            # Logica de negocio (auth con OAuth2+PKCE, tracker)
  ui/                 # Interfaz grafica con Fyne (config, login)
  mocks/              # Mocks para tests (testify)
  defines/            # Constantes de configuracion y lista de circuitos
```

### Capas principales

- **domain/** — Structs puros sin dependencias externas: `Config`, `User`, `TrackTime`, modelos de Google Sheets API (`SpreadsheetData`, `BatchUpdateRequest`, etc.)
- **repository/** — Acceso a datos externos:
  - `ConfigRepository` — lee/escribe config local con Viper (JSON)
  - `UserRepository` — consulta Google userinfo API
  - `SheetsRepository` — interactua con Google Sheets API v4 (lectura, escritura, creacion de hojas)
- **service/** — Logica de negocio:
  - `AuthService` — flujo OAuth2+PKCE, refresh de tokens, gestion de config
  - `TrackerService` — operaciones sobre el spreadsheet (verificar/crear hojas, headers)
- **ui/** — Interfaz grafica con Fyne (navegacion entre pantallas)

## Flujo de la aplicacion

1. **Config screen** — Se muestra si no hay credenciales validas configuradas
2. **Login screen** — Se muestra si hay credenciales pero no hay token de Google
3. **Post-login** — Se obtiene info del usuario, se verifica/crea su hoja en el spreadsheet con headers F1, y se muestra la pantalla principal

## Dependencias

| Paquete | Uso |
|---------|-----|
| `fyne.io/fyne/v2` | UI desktop cross-platform |
| `github.com/gin-gonic/gin` | Servidor callback para OAuth |
| `github.com/go-resty/resty/v2` | HTTP client (con `SetPathParam` para URLs) |
| `github.com/pkg/browser` | Abrir navegador para login |
| `github.com/spf13/viper` | Config JSON local |
| `golang.org/x/oauth2` | PKCE helpers + token refresh |
| `github.com/stretchr/testify` | Testing (mocks + assertions) |
