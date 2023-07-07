# Wyze Motion Monitor

Wyze Motion Monitor is a Go application that monitors the latest image for motion events and the latest motion recordings in specified directories on Wyze cameras. It sends the paths of the latest JPG and MP4 files as a JSON payload to a webhook URL. This is particularly useful when using gtxaspec/wz_mini_hacks on Wyze cameras for local integration with Home Assistant.

## Prerequisites

- Go programming language (version 1.16 or later)
- Linux operating system (tested on amd64 and mipsle architecture)
- A destination webhook address (probably Home Assistant)

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
   ./wyze-motion-monitor <cameraName> <webhookURL>
