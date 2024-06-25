# Kleinanzeigen-alert


Telegram bot that notifies you of new Kleinanzeigen listings.
Run it yourself with the instructions below.


## Installation
Get your telegram token from [@botfarther](https://t.me/botfarther)

Run with docker-compose:

```bash
    git clone https://github.com/DanielStefanK/kleinanzeigen-alert.git alert && cd alert
    nano docker-compose.yaml //replace mytoken with your obtained token
    docker-compose up
```
with go on your system:

```bash
    git clone https://github.com/DanielStefanK/kleinanzeigen-alert.git alert && cd alert
    export TELEGRAM_APITOKEN=mytoken //replace mytoken with your obtained token
    go get
    go run main.go
```

## Usage/Examples

### Add search
write `/add {search term}, {city/zip}, {radius}, {optional max price without "€" and no decimal}?, {optional min price without "€" and no decimal}?`
e.g. `/add bicycle, Cologne, 20`
This will perform a search every minute and you will get the latest entries here.

### Search lists of everything
write `/list`
This will list all your current searches

### Remove searches
write `/remove {ID}`
You get the ID from the list command. This will delete the search and you will no longer receive messages for it.



### Add Custom Link
write `/link {Link}`
You can also provide a custom link for the bot to scrape. This link is validated to be something like `https://www.kleinanzeigen.de/s-XXXX`.


## Author
- [@DanielStefanK](https://github.com/DanielStefanK)