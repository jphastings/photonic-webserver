# Photonic Webserver

This project aims to be a **fully solar powered** static website server that's **simple to set up** and **simple to operate**, no matter where you are.

It runs on a [Raspberry Pi Zero 2w](https://www.raspberrypi.com/products/raspberry-pi-zero-2-w/), is powered by a [PhotonPowerZero](https://github.com/DavidMurrayP2P/PhotonPowerZero) and is orchestrated by [GoKrazy](https://gokrazy.org/).

## What you need

You will need at least:

- A [Raspberry Pi Zero 2w](https://www.raspberrypi.com/products/raspberry-pi-zero-2-w/) computer
- A [PhotonPowerZero](https://github.com/DavidMurrayP2P/PhotonPowerZero) solar power board
- A solar panel (eg. the [Voltaic 10W, 6V panel](https://voltaicsystems.com/10-watt-panel-etfe/))
- A LiPo battery (always use high quality batteries, [guidance here](https://github.com/DavidMurrayP2P/PhotonPowerZero/blob/main/LIPO_PURCHASING.md))
- And a microSD card at least 2GB in size

- A (free) account with [Cloudflare](https://cloudflare.com), and a domain configured with them
- A WiFi network you have access to wherever you'll be installing your panel
- Software for uploading files via WebDAV (eg. macOS Finder, [Panic's Transmit](https://panic.com/transmit/), Windows' Explorer, or [davfs](https://savannah.nongnu.org/projects/davfs2) on Linux)

## What you get

- A low power webserver for static sites
- Runs entirely off solar power
- Punches _out_ of your Wifi network, so no need to configure your router
- Can serve multiple domains at once
- Can update what's being served from anywhere on the web via WebDAV
- Can include current battery percentage and/or a graph of recent battery power within your sites (coming soon!)

## Getting started

_These instructions currently assume a reasonable knowledge of the *nix command line; I hope to make this process even easier in the future._

### First time

Install [Go](https://go.dev/) and [GoKrazy](https://gokrazy.org/quickstart/):

```bash
go install github.com/gokrazy/tools/cmd/gok@main
```

Clone this repository to somewhere on your computer:

```bash
git clone https://github.com/jphastings/photonic-webserver.git
cd photonic-webserver
```

Configure your Wifi settings:

```bash
cp wifi.json.example wifi.json
# Edit wifi.json to have your wifi credentials
```

Create a [Cloudflare tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/get-started/create-remote-tunnel/) and add the details to `cloudflared/config.yaml`. You can route as many domains to this tunnel as you like, each can have a separate static site.

```bash
cp cloudflared/config.yaml.example cloudflared/config.yaml
# Edit cloudflared/config.yaml with the details from Cloudflare
```

Change the password to access this GoKrazy instance, in the `"HTTPPassword": "<here>"` section near the top.

Create and add a password for updating your site(s):

```bash
brew install caddy
caddy hash-password
# Enter your password (twice) and copy the `$2a$lots.of.characters.y` into config.json after `WEBDAV_PASSWORD_HASH=`
```

Insert a microSD card into your system, [find its device name](https://gokrazy.org/quickstart/#step-1-insert-an-sd-card) and write this project to it with the following, where `/dev/sdx` is the device name of your microSD card:

```bash
gok overwrite --full /dev/sdx
```

Insert this SD card into your Raspberry Pi, with he PhotonPowerZero attached and solar panel/battery plugged in.

> [!CAUTION]
> Never power the Raspberry Pi via micro USB at the same time as a solar panel or battery.

You can test that it was successful by visiting [http://photon:8080/](http://photon:8080/) in your browser while on the same WiFi network as your Raspberry Pi. Note the `http` (not `https`) as this address will only work inside your home network.

You can also visit [http://photon/](http://photon/) while on the same WiFi network, the GoKrazy dashboard, if you'd like to see the internals of what is being run.

### Updating your sites

Updating any one of your sites is as simple as connecting to that site's domain over WebDAV with the username `admin` and the password generated above. You can do this with any tool you're familiar with, here's an example using [rclone](https://rclone.org/):

```bash
cd path/to/your/site

rclone copy . :webdav: --webdav-url=https://<your_site.example.com>/ --webdav-user=admin --webdav-pass=<your-password> --progress
```

> [!TIP]
> The domain you use to connect over WebDAV is the same site you will be updating. Each domain has its own directory (`/www/{your_site.example.com}`) on the fourth partition of the SD card in your Raspberry Pi.

### In the future

If I release updates to this project you can copy them to your Raspberry Pi easily with:

```bash
cd photonic-webserver

git pull origin main
# You may have to deal with merge conflicts in config.json, because of the WEBDAV_PASSWORD_HASH line
# I'm seeking ways to prevent this
gok update
```

Note that, when running `gok update`, your server will power down, and (currently) may take a long time to power back on again depending on how full your battery is.
