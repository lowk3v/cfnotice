# CFNotice

Stands for Cloudflare Notice
A tool lists, detects, notifies when Cloudflare DNS of targets changed

## Feature

## Usage

```
Usage of cfnotice:
  -c string
        a cookie (file or string) of your Cloudflare access. Default is empty
  -debug
        Enable debugging. Default is false
  -i int
         an interval to re-check. Disable by default
  -k string
        an API Key of your Cloudflare access. Default is empty
  -s string
        Load a specific storage path or set the CF_NOTICE_PATH environment. Default is ~/.config/cf-notice.json
  -zid string
        Cloudflare zone id
  -zno int
        Cloudflare zone number. Default is 0

```

## Example

- Get all CF Dns by a cookie file: `cfnotice -c ./cookie`
- Pick the second zone id: `cfnotice -c ./cookie -zno 2`
- Specific zone id: `cfnotice -c ./cookie -zid abcdefghijklmno`
- Run script daily with api key: `cfnotice -k secretapikey -i 1`

## Demo


## Install

`go install github.com/lowk3v/cfnotice@latest`

## Disclaimer

This tool is for educational purposes only. You are responsible for your own actions. If you mess something up or break any laws while using this software, it's your fault, and your fault only.

## License

`CFNotice` is made with ♥ by [@LowK3v](github.com/LowK3v) and it is released under the MIT license.

## Donate

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](paypal.me/lpdat)

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://buymeacoffee.com/lowk)
