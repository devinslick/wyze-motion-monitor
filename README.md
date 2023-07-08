# Wyze Motion Monitor

Wyze Motion Monitor is a Go application for Wyze cameras that monitors the latest event detection images, copies them to the locally available webserver and, optionally, notifies a webhook address that the camera has a new image.  Inspired by gtxaspec/wz_mini_hacks and intended to be used for local integration with Home Assistant.

## Prerequisites

- Go 1.18 installed on a system where you can compile the binary
- A Wyzecam that you have SSH access to.  Check out gtxaspec/wz_mini_hacks if you need help with this.  Tested on v3 cameras, but should work on others.
- Optionally, a destination webhook address (probably Home Assistant) to notify of the new JPG.

## Installation

1. Clone the repository on a system with Go installed
   ```shell
   git clone https://github.com/devinslick/wyze-motion-monitor.git
2. Build
   ```shell
   GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -o wyze-motion-monitor
3. Copy the resulting binary (wyze-motion-monitor) to your Wyze camera(s) using scp, WinSCP, or your sdcard.
4. Test the execution from a shell
   ```shell
   Without a webhook:
   ./wyze-motion-monitor
   With a webhook notification:
   ./wyze-motion-monitor <cameraName> <webhookURL>
5. Install a cron job to start the process if it isn't running
   ```shell
   Without a webhook:
   echo '* * * * * /media/mmc/wz_mini/wyze-motion-monitor' >> /media/mmc/wz_mini/etc/cron/root
   With a webhook notification:
   echo '* * * * * /media/mmc/wz_mini/wyze-motion-monitor FrontPorchCamera https://myhomeassistant.domain.com/api/webhook/secretWebHookAddress' >> /media/mmc/wz_mini/etc/cron/root
