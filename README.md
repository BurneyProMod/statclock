# FACEIT Dot Matrix Display

This project fetches a FACEIT player’s ELO and prints it to stdout (e.g., `Elo: 2123`). The end goal is to drive an dot matrix display mounted on 3D-printed Counter-Strike guns as a "stattrak" counter.

For now, the code only supports FaceIT ELO, but it will be expanded to additional stats such as Win/Loss, CS2 Premiere rank, league placement, and more.

## Roadmap / TODO

- Display additional FACEIT stats (Win/Loss, matches, account age, etc.)
- Display CS2 Premiere stats
- Determine/build the display unit  
  - Likely: **Adafruit 8x8 LED Backpack** (I2C) as an initial target
- Refactor main.go to cycle through stats to display

## Flashing ESP32
### Prereqs:
1) Install pipx
  `sudo apt install -y pipx`
2) Install ESPHome
`pipx install esphome`
### Flashing:
1) In the repo directory, use the command:
`esphome run esp32.yaml`

## FACEIT API Documentation
- https://docs.faceit.com/api/data/

## Leetify API Documentation
- https://leetify.com/blog/leetify-api-developer-guidelines/
## Requirements

- Pipx
- ESP-Home
- Go installed
- A FACEIT Data API key
- `.env` configured with your FACEIT API key and player nickname

## Setup
1) Edit the `.env.example` file in the project root:
```
FACEIT_API_KEY=your_key_here
FACEIT_NICKNAME=username_here
FACEIT_GAME=cs2
```
Then rename to `.env`

2) Install Dependencies:
```
go get github.com/joho/godotenv@latest
go mod tidy
```
4) Run:
`go run .`

Notes
The FACEIT API key should be treated as a secret. Do not embed it into distributed firmware/devices.

## References:
1) https://www.xda-developers.com/built-15-esp32-smart-clock/