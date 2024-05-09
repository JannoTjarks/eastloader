# eastloader
**A CLI Tool to download the "OstfriesischeNachrichten".**

The Publisher "Zeitungsgruppe Ostfriesland" is using the products of the company
[Visiolink](https://www.visiolink.com/). Visiolink provides a api, which is 
used for all download steps.
A simple api overview can be viewed by accessing this url:
[Visiolink Settings/API](https://device.e-pages.dk/settings/current.php?vl_platform=desktop&vl_app_id=dk.e-pages.ostfriesischenachrichten&vl_app_version=1.21.02)

Currently this project allows to download the "Ostfriesische Nachrichten" and 
the "Ostfriesen Zeitung".

To start a download, you need just to build the project with go build for 
your architecture, or just run a `go run cmd/oz-downloader/main.go` respectively
`go run cmd/eastloader/main.go`. It will automatically 
download the newest issue.

### Environment variables
At the moment the user credentials have to to be set as environment variables.
In the future there will be also other methods to set the credentials.

#### Ostfriesische Nachrichten
##### ON_DOWNLOADER_USERNAME
Username of your on-online.de account
##### ON_DOWNLOADER_PASSWORD
Password of your on-online.de account

#### Ostfriesen Zeitung
##### OZ_DOWNLOADER_USERNAME
Username of your on-online.de account
##### OZ_DOWNLOADER_PASSWORD
Password of your on-online.de account
