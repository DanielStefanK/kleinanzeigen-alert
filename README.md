
# Kleinanzeigen-alert


A very simple telegram bot that notifies you of new Ebay-Kleinanzeigen listings.



## Demo
![](example.gif)


## Installation
Get your telegram token from @botfarther

Run with Docker:

```bash
    git clone https://github.com/DanielStefanK/kleinanzeigen-alert.git kl && cd kl
    nano docker-compose.yaml //replace mytoken with your token
    docker-compose up
```
Just run it:

```bash
    git clone https://github.com/DanielStefanK/kleinanzeigen-alert.git kl && cd kl
    export TELEGRAM_APITOKEN=mytoken //replace mytoken with your token
    go get
    go run main.go
```

## Usage/Examples

 add a new search query:
```bash
/add search term, city, radius, pricemin, pricemax, type of sale
```
lists all search queries and show the corresponding ids:
```bash
/list
```
remove search query with the id ID:
```bash
/remove ID
```

## Authors

- [@DanielStefanK](https://github.com/DanielStefanK)
- [@Johannes-ece](https://github.com/Johannes-ece)

  