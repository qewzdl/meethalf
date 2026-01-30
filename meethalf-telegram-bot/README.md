# meethalf-telegram-bot

Base Telegram bot scaffold on Go with clean architecture layering.

## Run

Create `.env` (or copy `.env.example`) and set your token:

```bash
BOT_TOKEN=telegram-bot-token
```

Start:

```bash
go run ./cmd/bot
```

## Commands

- /start - greet user by profile name when available, otherwise Telegram name; show inline Start matching, Search with AI, and My profile (or Create profile when missing) buttons, plus Preferences only after a profile is created. On the first /start, the bot shows a formal 16+ age confirmation statement before opening the main menu.
- /cancel - cancel the current action and return to the main menu
- /language - open language selection (optional args: `en` or `ru`)

## Admin

Set `BOT_ADMIN_USERNAME` to one or more Telegram usernames (comma-separated). Admins see an `Admin dashboard` button in the main
menu that opens admin management. The admin dashboard supports `All users`, `Banned list`, `Shadow banned list`, `Hidden profiles`, `Moderator list`,
and `Reports` lists (reported users include report counts) that return users (including their Telegram usernames) from the Meethalf API
with pagination, plus `Ban a user` / `Lift ban` buttons that prompt for the profile id or `@username` (you can also use
`/ban <user_id|@username>` and `/unban <user_id|@username>`), plus `Shadow ban` / `Lift shadow ban`
(`/shadow_ban <user_id|@username>` and `/shadow_unban <user_id|@username>`), plus `Hide profile` / `Show profile`
(`/hide_profile <user_id|@username>` and `/show_profile <user_id|@username>`). It also supports `Grant moderator` / `Revoke moderator` actions
(`/moderator <user_id|@username>` and `/unmoderator <user_id|@username>`), and `Reset matches` to clear match history and
decisions (`/reset_choices <user_id|@username>`). `Reset 16+ check` clears the first-start age confirmation
(`/reset_start <user_id|@username>`). `Clear reports` removes a user from the reported list
(`/clear_reports <user_id|@username>`).
Moderators (profiles with `is_moderator=true` in the Meethalf API) see a `Moderator access enabled.` badge and a
`Moderator dashboard` button, but only with `All users`, `Banned list`, `Shadow banned list`, `Hidden profiles`, and `Reports` lists plus
`Ban a user` / `Lift ban`, `Shadow ban` / `Lift shadow ban`, `Hide profile` / `Show profile` (regular users only), and `Clear reports`.
Moderator management and the moderators list remain admin-only.
Admin UI labels are localized for English and Russian, including shadow-ban actions.

## Profile setup

The bot guides users through a nine-step profile setup: a short anti-bot verification check (choose the correct answer
from four buttons), name (use the Telegram name button or type a custom one), gender, birth date (DD.MM.YYYY), country
(Russia, Kazakhstan, or Belarus), city (selected from the supported list), description, emoji selection, and a photo
album (1-4 photos). The verification step plus gender, country, city, and emoji are selected via inline buttons. After
sending at least one photo, use the Done button to finish the setup. Drafts are
stored in the same session store so they can survive restarts when Redis is enabled. The setup header includes an estimated
total completion time (shown in minutes) calculated from per-step durations.
Birth date validation requires users to be at least 16 years old; younger users cannot access the bot.
Each setup prompt includes a Back to menu button to discard the draft and return to the main menu. Starting from the
Gender step, a Previous step button lets you return to the prior question. Profile edit steps use a Cancel button that
returns to the Profile Edit menu.
After the setup is completed, the bot deletes all setup messages in the chat from the start of the setup through the
Profile created confirmation.

Supported cities:

- Russia: Moscow, Saint Petersburg, Novosibirsk, Krasnodar, Omsk, Rostov-on-Don, Perm, Krasnoyarsk, Yekaterinburg, Kazan,
  Nizhny Novgorod, Ufa, Chelyabinsk, Samara, Voronezh, Volgograd.
- Kazakhstan: Astana, Almaty, Semey, Pavlodar, Shymkent, Aktobe, Karaganda, Taraz, Ust-Kamenogorsk, Atyrau.
- Belarus: Minsk, Gomel, Mogilev, Vitebsk, Grodno, Brest, Bobruisk, Baranovichi, Borisov.

Supported profile emoji codes:

- LDR - 👑
- STR - 🧠
- ANA - 🧩
- CRT - 🎨
- COM - 🤝
- EMP - ❤️
- MED - 🕊
- PRF - 🧼
- RSR - 🧭
- INN - 💡
- EXE - 🛠
- ADV - 🔥
- CNT - ☕️
- RLS - 🧱
- MOT - 🎯
- SKP - 🛡

## Profile view

Use the inline `My profile` button to see the saved profile details. If the profile does not exist, the button is labeled
`Create profile` and starts the setup flow. When a profile is missing and the user opens My profile or Preview, the bot
returns a not-found message and shows the `Create profile` button instead of Preview/Edit actions. The response includes a
`Preview profile` button (shows how other users see the profile) above the `Edit profile` button to start editing. The
edit button opens a menu so users can choose which field to update (name, gender, birth date, country, city, description,
emoji, or photos). The menu includes a `Back to profile` button. When the profile has photos, the bot sends them as an
album with the profile details in the caption.
Profile creation and updates send a confirmation message followed by the full profile details with the `Edit profile`
button (including the number of saved photos).

## Loading messages

Operations that call the Meethalf API (profile view, edits, saves, and confirmed deletes) first send short, action-specific
loading messages so users see progress while the request is in flight. The loading message is deleted after the final
response is sent.

## Message replacement

When you press an inline action button, the bot first deletes the message that contained it and only then sends the next
prompt or result. For album-based profile cards that need a separate action message, the bot also removes the album so the
entire card is replaced. This keeps the chat tidy and makes each step feel like a replacement.

## Profile settings

The inline `Preferences` button appears only when a profile already exists and opens profile settings. The menu includes a `Delete profile`
button that opens a confirmation step, a `Language` button to switch between English and Russian, an
`Advanced search` toggle (off by default) that controls whether the bot asks for match accuracy, and a
`Hide my profile` / `Show my profile` toggle that controls whether your profile appears in search results. If the
profile does not exist, Preferences returns a not-found message and shows the `Create profile` button instead of delete.
Confirming removes the profile via the Meethalf API, and cancel keeps the profile unchanged.

## Language

The bot supports English and Russian. The initial language is picked from the Telegram `language_code` when available,
otherwise English. You can change it anytime in Preferences > Language or by sending `/language` (optionally `/language en` or `/language ru`); the selection is stored in the session store and
applies to all future bot responses.

## Search flow

Use the `Start matching` button from `/start` to start browsing. If you press it without a profile, the bot asks you to
create one first. The bot always asks for the gender to search. With advanced search disabled (default), it immediately
starts matching with accuracy 4 (strict) and the Meethalf API relaxes accuracy step-by-step down to 0 if needed. When
advanced search is enabled in Preferences, the bot also asks for the match accuracy level (0-4). The prompt explains the
scale (0 is wider/random, 4 is strict/precise) and the keyboard shows a single-row 4-0 button layout. If no profiles
match the selected accuracy, the search widens automatically until it finds candidates. The Meethalf API enforces age
eligibility: users aged 16-17 see only 16-17 profiles, and users aged 18+ never see profiles under 18. After
that it shows profile cards with action buttons:

- 👎 - skip and show the next profile
- ❤️ - send a like; the recipient gets a like notification (immediately when possible) with a button to open your profile
- Report 🚩 - report and show the next profile
- Previous profile - open the previous profile when available
- My history - open the list of viewed profiles across sessions (with current decisions) and change them

When you receive likes, the bot sends a notification right away when it has your chat session; otherwise it shows
notifications on `/start` with a button to view the sender profile. Mutual likes trigger a match message that shares each
user's nickname (Telegram `@username` when available, falling back to the profile name). Viewing other profiles (search
results and likes) requires an existing profile; otherwise the bot prompts you to create one.
Search prompts and match actions include a Back to menu button to exit the flow and return to the main menu. The match
accuracy step shows a Cancel button that returns to gender selection. When no matching profiles are available, the bot
shows a Refresh feed button next to Back to menu.

Use the `Search with AI` button in the main menu to describe who you want to meet in free-form text. The bot sends your
message to the Meethalf API, which uses OpenRouter to analyze preferences and returns the most suitable profile. The
result is shown with the same match action buttons.

## Docker

```bash
docker compose up --build
```

Stop:

```bash
docker compose down
```

## Environment

- APP_ENV (dev)
- BOT_TOKEN ()
- BOT_ADMIN_USERNAME ()
- BOT_DEBUG (false)
- BOT_ALLOWED_UPDATES (message,callback_query)
- BOT_POLLING_TIMEOUT (10s)
- BOT_API_ENDPOINT ()
- BOT_PROXY_URL ()
- BOT_WORKERS (4)
- BOT_QUEUE_SIZE (100)
- SESSION_STORE (memory)
- SESSION_TTL (24h)
- REDIS_ENABLED (false)
- REDIS_HOST (localhost)
- REDIS_PORT (6379)
- REDIS_USERNAME ()
- REDIS_PASSWORD ()
- REDIS_DB (0)
- REDIS_CONNECT_TIMEOUT (5s)
- REDIS_READ_TIMEOUT (3s)
- REDIS_WRITE_TIMEOUT (3s)
- REDIS_POOL_SIZE (0)
- REDIS_MIN_IDLE_CONNS (0)
- API_BASE_URL (http://localhost:8080)
- API_TIMEOUT (5s)

`BOT_TOKEN` is required on startup. `.env` is loaded automatically if present. Long polling is used by
default. Use `SESSION_STORE=redis` with `REDIS_ENABLED=true` to share session state across multiple bot
instances. Use `BOT_API_ENDPOINT` to override the Telegram API endpoint (accepts a full format string
like `https://api.telegram.org/bot%s/%s` or a base URL). Set `BOT_PROXY_URL` to force a proxy for
Telegram requests; if it is empty, `HTTP_PROXY` and `HTTPS_PROXY` are still honored. Set
`BOT_ADMIN_USERNAME` to the admin Telegram username (with or without `@`) to mark admin users in the
bot greeting; comma-separated values are accepted.

## Structure

- cmd/bot - entrypoint
- internal/app - dependency wiring and run
- internal/config - config
- internal/domain - domain entities
- internal/usecase/bot - bot usecase (command routing, profile flow, response builders)
- internal/storage/memory - in-memory repositories
- internal/storage/redis - Redis repositories
- internal/transport/api - HTTP client for Meethalf API
- internal/transport/telegram - Telegram transport (poller, handler, sender)
- internal/logger - logging

## Architecture boundaries

Run the boundary check:

```bash
go test ./internal/architecture
```

It enforces clean architecture layering inside `internal`, keeps Telegram transport isolated from the Meethalf API client
adapter, and blocks imports from the API module.



