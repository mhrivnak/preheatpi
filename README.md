# preheatpi

This app runs on a raspberry pi or similar that is attached to one or more relays via
GPIO pins. It turns those relays on and off based on data received from
[preheatbot](https://github.com/mhrivnak/preheatbot/).

### Config

You must set these environment variables:

`PREHEATBOTURL`: The URL for the preheatbot API.

`PREHEATBOTUSERNAME`: Your username with preheatbot.

`RELAYS`: A comma-separated string alternating GPIO pin and an identifier. Example
with two relays: `8,N12345-engine,10,N12345-cabin`