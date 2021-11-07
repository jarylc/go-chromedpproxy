[![ChromeDP Proxy](go-chromedpportal.png)](https://github.com/jarylc/go-chromedpproxy)

# Carousell GoBot
Inspired by [claabs/puppeteer-extra-plugin-portal](https://github.com/claabs/puppeteer-extra-plugin-portal)

A Go module for a ChromeDP abstraction layer that additionally hosts a webserver to remotely view ChromeDP sessions. Essentially opening a "portal" to the page. Perfect for Go automations that require manual intervention.

[Report Bugs](https://github.com/jarylc/go-chromedpproxy/issues/new) · [Request Features](https://github.com/jarylc/go-chromedpproxy/issues/new)

## About The Project
### Built With
* [golang](https://golang.org/)
* [chromedp/chromedp](https://github.com/chromedp/chromedp)
* [gofiber/fiber](https://github.com/gofiber/fiber)
    * [gofiber/websocket](https://github.com/gofiber/websocket)
      * [fasthttp/websocket](https://github.com/fasthttp/websocket)
### Examples
Please proceed to the [examples directory](examples)
- [Google Recaptcha v2](examples/google_recaptcha_v2)
- [hcaptcha](examples/hcaptcha)

## Getting Started
### Installing
```shell
go get -u github.com/jarylc/go-chromedpproxy
```
### Usage
#### Basic
```go
// step 1 - prepare proxy
// change `:9222` to your desired Chrome remote debugging port
// change `:9221` to your desired webserver port for the front-end
// append and/or remove `chromedp.DisableGPU` with your desired allocated executor options as additional arguments
chromedpproxy.PrepareProxy(":9222", ":9221", chromedp.DisableGPU)

// step 2 - launch a tab target
// append and/or remove `chromedp.WithLogf(log.Printf)` with your desired context options as additional arguments
targetID, err := chromedpproxy.NewTab("https://www.google.com/recaptcha/api2/demo", chromedp.WithLogf(log.Printf))
if err != nil && !errors.Is(err, context.Canceled) {
    log.Fatal(err)
}
defer chromedpproxy.CloseTarget(targetID)

// step 3 - get context and do whatever you need, you can refer to the examples directory of the project
ctx := chromedpproxy.GetTarget(targetID)
// ...
```
#### ChromeDP
Please refer to [chromedp/chromedp](https://github.com/chromedp/chromedp) repository for general `chromedp` usage


## Roadmap
See the [open issues](https://github.com/jarylc/go-chromedpproxy/issues) for a list of proposed features (and known
issues).


## Contributing
Feel free to fork the repository and submit pull requests.


## License
Distributed under the GNU GENERAL PUBLIC LICENSE V3. See `LICENSE` for more information.


## Contact
Jaryl Chng - git@jarylchng.com

https://jarylchng.com
