# Code Citations

## License: MIT
https://github.com/DanielStefanK/kleinanzeigen-alert/blob/3c76febff6bc89bb7d439907f849391fca7a9f16/pkg/telegram/commandHandler.go

```
func formatAdRaw(ad scraper.Ad, term string, id int) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("%s - %s\n", ad.Title, ad.Price))
	b.WriteString(f("in %s \n", ad.Location))
	b.WriteString(f("For search \"%s\" (ID: %v)\n", term, id))
	b.WriteString(f("Link: %s", ad.Link))

	return b.String
```

